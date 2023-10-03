// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/k0sproject/bootloose/cmd/bootloose"

	log "github.com/sirupsen/logrus"
)

func main() {
	if err := bootloose.NewRootCommand(context.Background()).Execute(); err != nil {
		log.Fatal(err)
	}
}
