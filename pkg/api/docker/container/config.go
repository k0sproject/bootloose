// SPDX-FileCopyrightText: 2025 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package container

// Subset of Config from
// https://github.com/moby/moby/blob/v28.3.3/api/types/container/config.go#L44
type Config struct {
	Cmd []string // Originally type 'strslice' but we can use a simple slice of strings
}
