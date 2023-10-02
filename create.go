// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package bootloose

import (
	"github.com/spf13/cobra"

	"github.com/k0sproject/bootloose/pkg/cluster"
)

func NewCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a cluster",
		RunE:  create,
	}
}

func create(cmd *cobra.Command, _ []string) error {
	cluster, err := cluster.NewFromFile(clusterConfigFile(cmd))
	if err != nil {
		return err
	}
	return cluster.Create()
}
