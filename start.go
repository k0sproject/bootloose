// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/k0sproject/bootloose/pkg/cluster"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start cluster machines",
	RunE:  start,
}

var startOptions struct {
	config string
}

func init() {
	startCmd.Flags().StringVarP(&startOptions.config, "config", "c", Bootloose, "Cluster configuration file")
	bootloose.AddCommand(startCmd)
}

func start(cmd *cobra.Command, args []string) error {
	cluster, err := cluster.NewFromFile(configFile(startOptions.config))
	if err != nil {
		return err
	}
	return cluster.Start(args)
}
