// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package bootloose

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

// ConfigFile is the name of the default configuration file.
const ConfigFile = "bootloose.yaml"

type contextKey string

const configFileKey contextKey = "configFile"

func NewRootCommand(ctx context.Context) *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:           "bootloose",
		Short:         "bootloose - Container Machines",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.SetContext(ctx)

	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", ConfigFile, "Cluster configuration file")

	cmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		if flag := cmd.Flags().Lookup("config"); flag != nil && !flag.Hidden {
			cmd.SetContext(context.WithValue(cmd.Context(), configFileKey, configFile))
		}
	}

	cmd.AddCommand(
		NewConfigCommand(),
		NewCreateCommand(),
		NewShowCommand(),
		NewDeleteCommand(),
		NewStartCommand(),
		NewStopCommand(),
		NewSSHCommand(),
	)

	// hide config flag from commands that do not need it
	for _, configlessCmd := range []*cobra.Command{
		NewVersionCommand(),
	} {
		cmd.AddCommand(configlessCmd)
		if flag := configlessCmd.Flags().Lookup("config"); flag != nil {
			flag.Hidden = true
		}
	}

	return cmd
}

func configFile(f string) string {
	env := os.Getenv("BOOTLOOSE_CONFIG")
	if env != "" && f == ConfigFile {
		return env
	}
	return f
}

func clusterConfigFile(cmd *cobra.Command) string {
	cfg, ok := cmd.Context().Value(configFileKey).(string)
	if !ok {
		return configFile(ConfigFile)
	}
	return configFile(cfg)
}

