// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package bootloose

import (
	"fmt"
	"os"

	"github.com/k0sproject/bootloose/pkg/cluster"
	"github.com/spf13/cobra"
)

type showOptions struct {
	output string
}

func NewShowCommand() *cobra.Command {
	opts := &showOptions{}
	cmd := &cobra.Command{
		Use:     "show [HOSTNAME]",
		Aliases: []string{"status"},
		Short:   "Show all running machines or a single machine with a given hostname.",
		Long: `Provides information about machines created by bootloose in JSON or Table format.
	Optionally, provide show with a hostname to look for a specific machine. Exp: 'show node0'.`,
		RunE: opts.show,
		Args: cobra.MaximumNArgs(1),
	}
	cmd.Flags().StringVarP(&opts.output, "output", "o", "table", "Output formatting options: {json,table}.")
	return cmd
}

// show will show all machines in a given cluster.
func (opts *showOptions) show(cmd *cobra.Command, args []string) error {
	c, err := cluster.NewFromFile(configFile(clusterConfigFile(cmd)))
	if err != nil {
		return err
	}
	var formatter cluster.Formatter
	switch opts.output {
	case "json":
		formatter = new(cluster.JSONFormatter)
	case "table":
		formatter = new(cluster.TableFormatter)
	default:
		return fmt.Errorf("unknown formatter '%s'", opts.output)
	}
	machines, err := c.Inspect(args)
	if err != nil {
		return err
	}
	return formatter.Format(os.Stdout, machines)
}

