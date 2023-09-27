package main

import (
	"fmt"

	_ "github.com/carlmjohnson/versioninfo" // Needed to set version info during go install
	"github.com/k0sproject/bootloose/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print bootloose version",
	Run:   showVersion,
}

func init() {
	versionCmd.Flags().BoolP("long", "l", false, "Print long version")
	bootloose.AddCommand(versionCmd)
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
