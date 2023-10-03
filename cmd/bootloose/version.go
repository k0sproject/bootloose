// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package bootloose

import (
	"fmt"

	_ "github.com/carlmjohnson/versioninfo" // Needed to set version info during go install
	"github.com/k0sproject/bootloose/version"

	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print bootloose version",
		Run:   showVersion,
	}
	cmd.Flags().BoolP("long", "l", false, "Print long version")
	return cmd
}

func showVersion(cmd *cobra.Command, _ []string) {
	if long, err := cmd.Flags().GetBool("long"); err == nil && long {
		fmt.Println("version:", version.Version)
		fmt.Printf("commit: %s\n", version.GitCommit)
		fmt.Printf("environment: %s\n", version.Environment)
		return
	}
	fmt.Println(version.Version)
}

