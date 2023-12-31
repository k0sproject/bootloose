// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package bootloose

import (
	"errors"
	"fmt"
	"os/user"
	"strings"

	"github.com/spf13/cobra"

	"github.com/k0sproject/bootloose/pkg/cluster"
)

type sshOptions struct {
	verbose bool
}

func NewSSHCommand() *cobra.Command {
	opts := &sshOptions{}
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "SSH into a machine",
		Args:  validateArgs,
		RunE:  opts.ssh,
	}
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "SSH verbose output")
	return cmd
}

func (opts *sshOptions) ssh(cmd *cobra.Command, args []string) error {
	cluster, err := cluster.NewFromFile(clusterConfigFile(cmd))
	if err != nil {
		return err
	}
	var node string
	var username string
	if strings.Contains(args[0], "@") {
		items := strings.Split(args[0], "@")
		if len(items) != 2 {
			return fmt.Errorf("bad syntax for user@node: %v", items)
		}
		username = items[0]
		node = items[1]
	} else {
		node = args[0]
		user, err := user.Current()
		if err != nil {
			return errors.New("error in getting current user")
		}
		username = user.Username
	}
	var remoteArgs []string
	if opts.verbose {
		remoteArgs = append(remoteArgs, "-v")
	}
	remoteArgs = append(remoteArgs, args[1:]...)
	return cluster.SSH(node, username, remoteArgs...)
}

func validateArgs(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("missing machine name argument")
	}
	return nil
}

