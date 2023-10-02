// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package bootloose

import (
	"fmt"
	"os"

	"github.com/k0sproject/bootloose/pkg/cluster"
	"github.com/k0sproject/bootloose/pkg/config"

	"github.com/spf13/cobra"
)

type configCreateOptions struct {
	override bool
	config   config.Config
}

func NewConfigCreateCommand() *cobra.Command {
	opts := &configCreateOptions{config: config.DefaultConfig()}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a cluster configuration",
		RunE:  opts.create,
	}

	cmd.Flags().BoolVar(&opts.override, "override", false, "Override configuration file if it exists")

	name := &opts.config.Cluster.Name
	cmd.Flags().StringVarP(name, "name", "n", *name, "Name of the cluster")

	private := &opts.config.Cluster.PrivateKey
	cmd.Flags().StringVarP(private, "key", "k", *private, "Name of the private and public key files")

	networks := &opts.config.Machines[0].Spec.Networks
	cmd.Flags().StringSliceVar(networks, "networks", *networks, "Networks names the machines are assigned to")

	replicas := &opts.config.Machines[0].Count
	cmd.Flags().IntVarP(replicas, "replicas", "r", *replicas, "Number of machine replicas")

	image := &opts.config.Machines[0].Spec.Image
	cmd.Flags().StringVarP(image, "image", "i", *image, "Docker image to use in the containers")

	privileged := &opts.config.Machines[0].Spec.Privileged
	cmd.Flags().BoolVar(privileged, "privileged", *privileged, "Create privileged containers")

	containerCmd := &opts.config.Machines[0].Spec.Cmd
	cmd.Flags().StringVarP(containerCmd, "cmd", "d", *containerCmd, "The command to execute on the container")

	return cmd
}

// configExists checks whether a configuration file has already been created.
// Returns false if not true if it already exists.
func configExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) || os.IsPermission(err) {
		return false
	}
	return !info.IsDir()
}

func (opts *configCreateOptions) create(cmd *cobra.Command, args []string) error {
	cluster, err := cluster.New(opts.config)
	if err != nil {
		return err
	}
	cfgFile := clusterConfigFile(cmd)
	if configExists(cfgFile) && !opts.override {
		return fmt.Errorf("configuration file at %s already exists", cfgFile)
	}
	return cluster.Save(cfgFile)
}

