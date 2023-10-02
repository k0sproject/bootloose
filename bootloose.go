// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Bootloose is the default name of the bootloose file.
const Bootloose = "bootloose.yaml"

var bootloose = &cobra.Command{
	Use:           "bootloose",
	Short:         "bootloose - Container Machines",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func configFile(f string) string {
	env := os.Getenv("BOOTLOOSE_CONFIG")
	if env != "" && f == Bootloose{
		return env
	}
	return f
}

func main() {
	if err := bootloose.Execute(); err != nil {
		log.Fatal(err)
	}
}
