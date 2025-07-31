// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/k0sproject/bootloose/pkg/api/docker/network"
	"github.com/k0sproject/bootloose/pkg/config"
	"github.com/k0sproject/bootloose/pkg/docker"
)

// Machine is a single machine.
type Machine struct {
	spec *config.Machine

	// container name.
	name string
	// container hostname.
	hostname string
	// container ip.
	ip string

	runtimeNetworks []*RuntimeNetwork
	// Fields that are cached from the docker daemon.

	ports map[int]int
	// maps containerPort -> hostPort.
}

// ContainerName is the name of the running container corresponding to this
// Machine.
func (m *Machine) ContainerName() string {
	return m.name
}

// Hostname is the machine hostname.
func (m *Machine) Hostname() string {
	return m.hostname
}

// IsCreated returns if a machine is has been created. A created machine could
// either be running or stopped.
func (m *Machine) IsCreated() bool {
	res, _ := docker.Inspect(m.name, "{{.Name}}")
	if len(res) > 0 && len(res[0]) > 0 {
		return true
	}
	return false
}

// IsStarted returns if a machine is currently started or not.
func (m *Machine) IsStarted() bool {
	res, _ := docker.Inspect(m.name, "{{.State.Running}}")
	parsed, _ := strconv.ParseBool(strings.Trim(res[0], `'`))
	return parsed
}

// HostPort returns the host port corresponding to the given container port.
func (m *Machine) HostPort(containerPort int) (int, error) {
	if !m.IsCreated() {
		return -1, fmt.Errorf("hostport: container %s is not created", m.name)
	}

	if !m.IsStarted() {
		return -1, fmt.Errorf("hostport: container %s is not started", m.name)
	}

	// Use the cached version first
	if hostPort, ok := m.ports[containerPort]; ok {
		return hostPort, nil
	}

	var hostPort int

	// retrieve the specific port mapping using docker inspect
	lines, err := docker.Inspect(m.ContainerName(), fmt.Sprintf("{{(index (index .NetworkSettings.Ports \"%d/tcp\") 0).HostPort}}", containerPort))
	if err != nil {
		return -1, fmt.Errorf("hostport: failed to inspect container: %v: %w", lines, err)
	}
	if len(lines) != 1 {
		return -1, fmt.Errorf("hostport: should only be one line, got %d lines", len(lines))
	}

	port := strings.Replace(lines[0], "'", "", -1)
	if hostPort, err = strconv.Atoi(port); err != nil {
		return -1, fmt.Errorf("hostport: failed to parse string to int: %w", err)
	}

	if m.ports == nil {
		m.ports = make(map[int]int)
	}

	// Cache the result
	m.ports[containerPort] = hostPort
	return hostPort, nil
}

func (m *Machine) networks() ([]*RuntimeNetwork, error) {
	if len(m.runtimeNetworks) != 0 {
		return m.runtimeNetworks, nil
	}

	var networks map[string]*network.EndpointSettings
	if err := docker.InspectObject(m.name, ".NetworkSettings.Networks", &networks); err != nil {
		return nil, err
	}
	m.runtimeNetworks = NewRuntimeNetworks(networks)
	return m.runtimeNetworks, nil
}

func (m *Machine) dockerStatus(s *MachineStatus) error {
	var ports []port
	if m.IsCreated() {
		for _, v := range m.spec.PortMappings {
			hPort, err := m.HostPort(int(v.ContainerPort))
			if err != nil {
				hPort = 0
			}
			p := port{
				Host:  hPort,
				Guest: int(v.ContainerPort),
			}
			ports = append(ports, p)
		}
	}
	if len(ports) < 1 {
		for _, p := range m.spec.PortMappings {
			ports = append(ports, port{Host: 0, Guest: int(p.ContainerPort)})
		}
	}
	s.Ports = ports

	s.RuntimeNetworks, _ = m.networks()

	return nil
}

// Status returns the machine status.
func (m *Machine) Status() *MachineStatus {
	s := MachineStatus{}
	s.Container = m.ContainerName()
	s.Image = m.spec.Image
	s.Command = m.spec.Cmd
	s.Spec = m.spec
	s.Hostname = m.Hostname()
	s.IP = m.ip
	state := NotCreated

	if m.IsCreated() {
		state = Stopped
		if m.IsStarted() {
			state = Running
		}
	}
	s.State = state

	_ = m.dockerStatus(&s)

	return &s
}
