package main

import (
	"github.com/spf13/cobra"

	"github.com/k0sproject/bootloose/pkg/cluster"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cluster",
	RunE:  create,
}

var createOptions struct {
	config string
}

func init() {
	createCmd.Flags().StringVarP(&createOptions.config, "config", "c", Bootloose, "Cluster configuration file")
	bootloose.AddCommand(createCmd)
}

func create(cmd *cobra.Command, args []string) error {
	cluster, err := cluster.NewFromFile(configFile(createOptions.config))
	if err != nil {
		return err
	}
	return cluster.Create()
}
