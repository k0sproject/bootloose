package main

import (
	"github.com/spf13/cobra"

	"github.com/k0sproject/bootloose/pkg/cluster"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a cluster",
	RunE:  delete,
}

var deleteOptions struct {
	config string
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteOptions.config, "config", "c", Bootloose, "Cluster configuration file")
	bootloose.AddCommand(deleteCmd)
}

func delete(cmd *cobra.Command, args []string) error {
	cluster, err := cluster.NewFromFile(configFile(deleteOptions.config))
	if err != nil {
		return err
	}
	return cluster.Delete()
}
