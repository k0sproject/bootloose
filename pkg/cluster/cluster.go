// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"context"
	"crypto/ed25519"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/ghodss/yaml"
	"github.com/k0sproject/bootloose/pkg/config"
	"github.com/k0sproject/bootloose/pkg/docker"
	"github.com/k0sproject/bootloose/pkg/exec"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// Container represents a running machine.
type Container struct {
	ID string
}

// Cluster is a running cluster.
type Cluster struct {
	spec     config.Config
	keyStore *KeyStore
}

// New creates a new cluster. It takes as input the description of the cluster
// and its machines.
func New(conf config.Config) (*Cluster, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	return &Cluster{
		spec: conf,
	}, nil
}

// NewFromYAML creates a new Cluster from a YAML serialization of its
// configuration available in the provided string.
func NewFromYAML(data []byte) (*Cluster, error) {
	spec := config.Config{}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	return New(spec)
}

// NewFromFile creates a new Cluster from a YAML serialization of its
// configuration available in the provided file.
func NewFromFile(path string) (*Cluster, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewFromYAML(data)
}

// SetKeyStore provides a store where to persist public keys for this Cluster.
func (c *Cluster) SetKeyStore(keyStore *KeyStore) *Cluster {
	c.keyStore = keyStore
	return c
}

// Name returns the cluster name.
func (c *Cluster) Name() string {
	return c.spec.Cluster.Name
}

// Save writes the Cluster configure to a file.
func (c *Cluster) Save(path string) error {
	data, err := yaml.Marshal(c.spec)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0666)
}

func f(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func (c *Cluster) containerName(machine *config.Machine) string {
	return fmt.Sprintf("%s-%s", c.spec.Cluster.Name, machine.Name)
}

func (c *Cluster) containerNameWithIndex(machine *config.Machine, i int) string {
	format := "%s-" + machine.Name
	return f(format, c.spec.Cluster.Name, i)
}

// NewMachine creates a new Machine in the cluster.
func (c *Cluster) NewMachine(spec *config.Machine) *Machine {
	return &Machine{
		spec:     spec,
		name:     c.containerName(spec),
		hostname: spec.Name,
	}
}

func (c *Cluster) machine(spec *config.Machine, i int) *Machine {
	return &Machine{
		spec:     spec,
		name:     c.containerNameWithIndex(spec, i),
		hostname: f(spec.Name, i),
	}
}

func (c *Cluster) forEachMachine(do func(*Machine, int) error) error {
	machineIndex := 0
	for _, template := range c.spec.Machines {
		for i := 0; i < template.Count; i++ {
			// machine name indexed with i
			machine := c.machine(template.Spec, i)
			// but to prevent port collision, we use machineIndex for the real machine creation
			if err := do(machine, machineIndex); err != nil {
				return err
			}
			machineIndex++
		}
	}
	return nil
}

func (c *Cluster) forSpecificMachines(do func(*Machine, int) error, machineNames []string) error {
	// machineToStart map is used to track machines to make actions and non existing machines
	machineToStart := make(map[string]bool)
	for _, machine := range machineNames {
		machineToStart[machine] = false
	}
	for _, template := range c.spec.Machines {
		for i := 0; i < template.Count; i++ {
			machine := c.machine(template.Spec, i)
			_, ok := machineToStart[machine.name]
			if ok {
				if err := do(machine, i); err != nil {
					return err
				}
				machineToStart[machine.name] = true
			}
		}
	}
	// log warning for non existing machines
	for key, value := range machineToStart {
		if !value {
			log.Warnf("machine %v does not exist", key)
		}
	}
	return nil
}

func (c *Cluster) ensureSSHKey() error {
	if c.spec.Cluster.PrivateKey == "" {
		return nil
	}
	path, err := expandHomedir(c.spec.Cluster.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to expand private key path: %w", err)
	}

	if _, err := os.Stat(path); err == nil {
		return nil
	}

	// Generate the Ed25519 private and public key pair
	// and convert it into an SSH key pair.
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return fmt.Errorf("failed to generate new Ed25519 key: %w", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return fmt.Errorf("failed to convert Ed25519 public key into SSH public key: %w", err)
	}
	privPEM, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return fmt.Errorf("failed to convert Ed25519 private key into PEM block: %w", err)
	}
	sshPubBytes, sshPrivBytes := ssh.MarshalAuthorizedKey(sshPub), pem.EncodeToMemory(privPEM)

	// Save the key pair (unencrypted).
	if err := os.WriteFile(path, sshPrivBytes, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}
	if err := os.WriteFile(path+".pub", sshPubBytes, 0644); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	return nil
}

const initScript = `
set -e
rm -f /run/nologin
sshdir=/root/.ssh
test -d "$sshdir" || mkdir $sshdir
chmod 700 $sshdir
touch $sshdir/authorized_keys; chmod 600 $sshdir/authorized_keys
`

func (c *Cluster) publicKey(machine *Machine) ([]byte, error) {
	// Prefer the machine public key over the cluster-wide key.
	if machine.spec.PublicKey != "" && c.keyStore != nil {
		data, err := c.keyStore.Get(machine.spec.PublicKey)
		if err != nil {
			return nil, err
		}
		data = append(data, byte('\n'))
		return data, err
	}

	// Cluster global key
	if c.spec.Cluster.PrivateKey == "" {
		return nil, errors.New("no SSH key provided")
	}

	path, err := expandHomedir(c.spec.Cluster.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to expand private key path: %w", err)
	}
	return os.ReadFile(path + ".pub")
}

// CreateMachine creates and starts a new machine in the cluster.
func (c *Cluster) CreateMachine(machine *Machine, i int) error {
	name := machine.ContainerName()

	publicKey, err := c.publicKey(machine)
	if err != nil {
		return err
	}

	// Start the container.
	log.Infof("Creating machine: %s ...", name)

	if machine.IsCreated() {
		log.Infof("Machine %s is already created...", name)
		return nil
	}

	cmd := "/sbin/init"
	if machine.spec.Cmd != "" {
		cmd = machine.spec.Cmd
	}

	runArgs := c.createMachineRunArgs(machine, name, i)
	_, err = docker.Create(machine.spec.Image,
		runArgs,
		[]string{cmd},
	)
	if err != nil {
		return err
	}

	if len(machine.spec.Networks) > 1 {
		for _, network := range machine.spec.Networks[1:] {
			log.Infof("Connecting %s to the %s network...", name, network)
			if network == "bridge" {
				if err := docker.ConnectNetwork(name, network); err != nil {
					return err
				}
			} else {
				if err := docker.ConnectNetworkWithAlias(name, network, machine.Hostname()); err != nil {
					return err
				}
			}
		}
	}

	if err := docker.Start(name); err != nil {
		return err
	}

	// Initial provisioning.
	if err := containerRunShell(name, initScript); err != nil {
		return err
	}
	if err := copy(name, publicKey, "/root/.ssh/authorized_keys"); err != nil {
		return err
	}

	return nil
}

func (c *Cluster) createMachineRunArgs(machine *Machine, name string, i int) []string {
	runArgs := []string{
		"-it",
		"--label", "io.k0sproject.bootloose.owner=bootloose",
		"--label", "io.k0sproject.bootloose.cluster=" + c.spec.Cluster.Name,
		"--name", name,
		"--hostname", machine.Hostname(),
		"--tmpfs", "/run",
		"--tmpfs", "/run/lock",
		"--tmpfs", "/tmp:exec,mode=777",
	}
	if docker.CgroupVersion() == "2" {
		runArgs = append(runArgs, "--cgroupns", "private")

		if !machine.spec.Privileged {
			// Non-privileged containers will have their /sys/fs/cgroup folder
			// mounted read-only, even when running in private cgroup
			// namespaces. This is a bummer for init systems. Containers could
			// probably remount the cgroup fs in read-write mode, but that would
			// require CAP_SYS_ADMIN _and_ a custom logic in the container's
			// entry point. Podman has `--security-opt unmask=/sys/fs/cgroup`,
			// but that's not a thing for Docker. The only other way to get a
			// writable cgroup fs inside the container is to explicitly mount
			// it. Some references:
			//   - https://github.com/moby/moby/issues/42275
			//   - https://serverfault.com/a/1054414

			// Docker will use cgroups like
			//   <cgroup-parent>/docker-{{ContainerID}}.scope.
			//
			// Ideally, we could mount those to /sys/fs/cgroup inside the
			// containers. But there's some chicken-and-egg problem, as we only
			// know the container ID _after_ the container creation. As a
			// duct-tape solution, we mount our own cgroup as the root, which is
			// unrelated to the Docker-managed one:
			//   <cgroup-parent>/cluster-{{ClusterID}}.scope/machine-{{MachineID}}.scope

			// FIXME: How to clean this up? Especially when Docker is being run
			// on a different machine?

			// Just assume that the cgroup fs is mounted at its default
			// location. We could try to figure this out via
			// /proc/self/mountinfo, but it's really not worth the hassle.
			const cgroupMountpoint = "/sys/fs/cgroup"

			// Use this as the parent cgroup for everything. Note that if Docker
			// uses the systemd cgroup driver, the cgroup name has to end with
			// .slice. This is not a requirement for the cgroupfs driver; it
			// won't care. Hence, just always use the .slice suffix, no matter
			// if it's required or not.
			const cgroupParent = "bootloose.slice"

			cg := path.Join(
				cgroupMountpoint, cgroupParent,
				fmt.Sprintf("cluster-%s.scope", c.spec.Cluster.Name),
				fmt.Sprintf("machine-%s.scope", name),
			)

			runArgs = append(runArgs,
				"--cgroup-parent", cgroupParent,
				"-v", fmt.Sprintf("%s:%s:rw", cg, cgroupMountpoint),
			)
		}
	} else {
		runArgs = append(runArgs, "-v", "/sys/fs/cgroup:/sys/fs/cgroup")
	}

	for _, volume := range machine.spec.Volumes {
		mount := f("type=%s", volume.Type)
		if volume.Source != "" {
			mount += f(",src=%s", volume.Source)
		}
		mount += f(",dst=%s", volume.Destination)
		if volume.ReadOnly {
			mount += ",readonly"
		}
		runArgs = append(runArgs, "--mount", mount)
	}

	for _, mapping := range machine.spec.PortMappings {
		publish := ""
		if mapping.Address != "" {
			publish += f("%s:", mapping.Address)
		}
		if mapping.HostPort != 0 {
			publish += f("%d:", int(mapping.HostPort)+i)
		}
		publish += f("%d", mapping.ContainerPort)
		if mapping.Protocol != "" {
			publish += f("/%s", mapping.Protocol)
		}
		runArgs = append(runArgs, "-p", publish)
	}

	if machine.spec.Privileged {
		runArgs = append(runArgs, "--privileged")
	}

	if len(machine.spec.Networks) > 0 {
		network := machine.spec.Networks[0]
		log.Infof("Connecting %s to the %s network...", name, network)
		runArgs = append(runArgs, "--network", machine.spec.Networks[0])
		if network != "bridge" {
			runArgs = append(runArgs, "--network-alias", machine.Hostname())
		}
	}

	return append(runArgs, machine.spec.ExtraArgs...)
}

// Create creates the cluster.
func (c *Cluster) Create() error {
	if err := c.ensureSSHKey(); err != nil {
		return err
	}
	if err := docker.IsRunning(); err != nil {
		return err
	}
	for _, template := range c.spec.Machines {
		if _, err := docker.PullIfNotPresent(template.Spec.Image, 2); err != nil {
			return err
		}
	}
	return c.forEachMachine(c.CreateMachine)
}

// DeleteMachine remove a Machine from the cluster.
func (c *Cluster) DeleteMachine(machine *Machine, i int) error {
	name := machine.ContainerName()
	if !machine.IsCreated() {
		log.Infof("Machine %s hasn't been created...", name)
		return nil
	}

	if machine.IsStarted() {
		log.Infof("Machine %s is started, stopping and deleting machine...", name)
		err := docker.Kill("KILL", name)
		if err != nil {
			return err
		}
		cmd := exec.Command(
			"docker", "rm", "--volumes",
			name,
		)
		return cmd.Run()
	}
	log.Infof("Deleting machine: %s ...", name)
	cmd := exec.Command(
		"docker", "rm", "--volumes",
		name,
	)
	return cmd.Run()
}

// Delete deletes the cluster.
func (c *Cluster) Delete() error {
	if err := docker.IsRunning(); err != nil {
		return err
	}
	return c.forEachMachine(c.DeleteMachine)
}

// Inspect will generate information about running or stopped machines.
func (c *Cluster) Inspect(hostnames []string) ([]*Machine, error) {
	if err := docker.IsRunning(); err != nil {
		return nil, err
	}
	machines, err := c.gatherMachines()
	if err != nil {
		return nil, err
	}
	if len(hostnames) > 0 {
		return c.machineFilering(machines, hostnames), nil
	}
	return machines, nil
}

func (c *Cluster) machineFilering(machines []*Machine, hostnames []string) []*Machine {
	// machinesToKeep map is used to know not found machines
	machinesToKeep := make(map[string]bool)
	for _, machine := range hostnames {
		machinesToKeep[machine] = false
	}
	// newMachines is the filtered list
	newMachines := make([]*Machine, 0)
	for _, m := range machines {
		if _, ok := machinesToKeep[m.hostname]; ok {
			machinesToKeep[m.hostname] = true
			newMachines = append(newMachines, m)
		}
	}
	for hostname, found := range machinesToKeep {
		if !found {
			log.Warnf("machine with hostname %s not found", hostname)
		}
	}
	return newMachines
}

func (c *Cluster) gatherMachines() (machines []*Machine, err error) {
	// Bootloose has no machines running. Falling back to display
	// cluster related data.
	machines = c.gatherMachinesByCluster()
	for _, m := range machines {
		if !m.IsCreated() {
			continue
		}

		var inspect types.ContainerJSON
		if err := docker.InspectObject(m.name, ".", &inspect); err != nil {
			return machines, err
		}

		// Set Ports
		ports := make([]config.PortMapping, 0)
		for k, v := range inspect.NetworkSettings.Ports {
			if len(v) < 1 {
				continue
			}
			p := config.PortMapping{}
			hostPort, _ := strconv.Atoi(v[0].HostPort)
			p.HostPort = uint16(hostPort)
			p.ContainerPort = uint16(k.Int())
			p.Address = v[0].HostIP
			ports = append(ports, p)
		}
		m.spec.PortMappings = ports
		// Volumes
		var volumes []config.Volume
		for _, mount := range inspect.Mounts {
			v := config.Volume{
				Type:        string(mount.Type),
				Source:      mount.Source,
				Destination: mount.Destination,
				ReadOnly:    mount.RW,
			}
			volumes = append(volumes, v)
		}
		m.spec.Volumes = volumes
		m.spec.Cmd = strings.Join(inspect.Config.Cmd, ",")
		m.ip = inspect.NetworkSettings.IPAddress
		m.runtimeNetworks = NewRuntimeNetworks(inspect.NetworkSettings.Networks)

	}
	return
}

func (c *Cluster) gatherMachinesByCluster() (machines []*Machine) {
	for _, template := range c.spec.Machines {
		for i := 0; i < template.Count; i++ {
			s := template.Spec
			machine := c.machine(s, i)
			machines = append(machines, machine)
		}
	}
	return
}

func (c *Cluster) startMachine(machine *Machine, i int) error {
	name := machine.ContainerName()
	if !machine.IsCreated() {
		log.Infof("Machine %s hasn't been created...", name)
		return nil
	}
	if machine.IsStarted() {
		log.Infof("Machine %s is already started...", name)
		return nil
	}
	log.Infof("Starting machine: %s ...", name)

	// Run command while sigs.k8s.io/kind/pkg/container/docker doesn't
	// have a start command
	cmd := exec.Command(
		"docker", "start",
		name,
	)
	return cmd.Run()
}

// Start starts the machines in cluster.
func (c *Cluster) Start(machineNames []string) error {
	if err := docker.IsRunning(); err != nil {
		return err
	}
	if len(machineNames) < 1 {
		return c.forEachMachine(c.startMachine)
	}
	return c.forSpecificMachines(c.startMachine, machineNames)
}

// StartMachines starts specific machines(s) in cluster
func (c *Cluster) StartMachines(machineNames []string) error {
	return c.forSpecificMachines(c.startMachine, machineNames)
}

func (c *Cluster) stopMachine(machine *Machine, i int) error {
	name := machine.ContainerName()

	if !machine.IsCreated() {
		log.Infof("Machine %s hasn't been created...", name)
		return nil
	}
	if !machine.IsStarted() {
		log.Infof("Machine %s is already stopped...", name)
		return nil
	}
	log.Infof("Stopping machine: %s ...", name)

	// Run command while sigs.k8s.io/kind/pkg/container/docker doesn't
	// have a start command
	cmd := exec.Command(
		"docker", "stop",
		name,
	)
	return cmd.Run()
}

// Stop stops the machines in cluster.
func (c *Cluster) Stop(machineNames []string) error {
	if err := docker.IsRunning(); err != nil {
		return err
	}
	if len(machineNames) < 1 {
		return c.forEachMachine(c.stopMachine)
	}
	return c.forSpecificMachines(c.stopMachine, machineNames)
}

// io.Writer filter that writes that it receives to writer. Keeps track if it
// has seen a write matching regexp.
type matchFilter struct {
	writer       io.Writer
	writeMatched bool // whether the filter should write the matched value or not.

	regexp  *regexp.Regexp
	matched bool
}

func (f *matchFilter) Write(p []byte) (n int, err error) {
	if f.regexp.Match(p) {
		f.matched = true
		if !f.writeMatched {
			return len(p), err
		}
	}
	// Write as is if no match
	return f.writer.Write(p)
}

// Matches:
// ssh_exchange_identification: read: Connection reset by peer
var connectRefused = regexp.MustCompile("(?m)(ssh|kex)_exchange_identification:.+?$")

// execSSH returns true if the command should be tried again.
func execSSH(args []string) (bool, error) {
	cmd := exec.Command("ssh", args...)

	refusedFilter := &matchFilter{
		writer:       os.Stderr,
		writeMatched: false,
		regexp:       connectRefused,
	}

	cmd.SetStdin(os.Stdin)
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(refusedFilter)

	err := cmd.Run()
	if err != nil && refusedFilter.matched {
		return true, err
	}
	return false, err
}

func (c *Cluster) machineFromHostname(hostname string) (*Machine, error) {
	for _, template := range c.spec.Machines {
		for i := 0; i < template.Count; i++ {
			if hostname == f(template.Spec.Name, i) {
				return c.machine(template.Spec, i), nil
			}
		}
	}
	return nil, fmt.Errorf("%s: invalid machine hostname", hostname)
}

func mappingFromPort(spec *config.Machine, containerPort int) (*config.PortMapping, error) {
	for i := range spec.PortMappings {
		if int(spec.PortMappings[i].ContainerPort) == containerPort {
			return &spec.PortMappings[i], nil
		}
	}
	return nil, fmt.Errorf("unknown containerPort %d", containerPort)
}

// SSH logs into the name machine with SSH.
func (c *Cluster) SSH(nodename string, username string, remoteArgs ...string) error {
	machine, err := c.machineFromHostname(nodename)
	if err != nil {
		return err
	}

	hostPort, err := machine.HostPort(22)
	if err != nil {
		return err
	}
	mapping, err := mappingFromPort(machine.spec, 22)
	if err != nil {
		return err
	}
	remote := "localhost"
	if mapping.Address != "" {
		remote = mapping.Address
	}
	path, err := expandHomedir(c.spec.Cluster.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to expand private key path: %w", err)
	}
	args := []string{
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking=no",
		"-o", "IdentitiesOnly=yes",
		"-o", "LogLevel=error",
		"-i", path,
		"-p", f("%d", hostPort),
		"-l", username,
		remote,
	}
	args = append(args, remoteArgs...)
	// If we ssh in a bit too quickly after the container creation, ssh errors out
	// with:
	//   ssh_exchange_identification: read: Connection reset by peer
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var lastErr error
	for {
		select {
		case <-ctx.Done():
			if lastErr == nil {
				return fmt.Errorf("ssh connection failed: %w", ctx.Err())
			}
			return fmt.Errorf("ssh connection failed: %s: %w", ctx.Err(), lastErr)
		case <-ticker.C:
			retry, lastErr := execSSH(args)
			if lastErr == nil {
				return nil
			}

			if !retry {
				return lastErr
			}
		}
	}
}

func expandHomedir(path string) (string, error) {
	// Needs to be either `~` or `~/...` (or also `~\...` on Windows)
	if path == "" || path[0] != '~' || (len(path) > 1 && !os.IsPathSeparator(path[1])) {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, path[1:]), nil
}
