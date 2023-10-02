// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package ignite

import (
	"github.com/k0sproject/bootloose/pkg/exec"
)

// Start starts an Ignite VM
func Start(name string) error {
	return exec.CommandWithLogging(execName, "start", name)
}
