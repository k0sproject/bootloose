package main

import (
	"errors"
	"fmt"
	"os/user"
	"strings"

	"github.com/spf13/cobra"

	"github.com/k0sproject/bootloose/pkg/cluster"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into a machine",
	Args:  validateArgs,
	RunE:  ssh,
}

var sshOptions struct {
	config  string
	verbose bool
}

func init() {
	sshCmd.Flags().StringVarP(&sshOptions.config, "config", "c", Bootloose, "Cluster configuration file")
	sshCmd.Flags().BoolVarP(&sshOptions.verbose, "verbose", "v", false, "SSH verbose output")
	bootloose.AddCommand(sshCmd)
}

func ssh(cmd *cobra.Command, args []string) error {
	cluster, err := cluster.NewFromFile(configFile(sshOptions.config))
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
	if sshOptions.verbose {
		remoteArgs = append(remoteArgs, "-v")
	}
	remoteArgs = append(remoteArgs, args[1:]...)
	return cluster.SSH(node, username, remoteArgs...)
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("missing machine name argument")
	}
	return nil
}
