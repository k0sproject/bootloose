// SPDX-FileCopyrightText: 2025 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package container

// Subset of MountPoint from
// https://github.com/moby/moby/blob/v28.3.3/api/types/container/container.go#L62
type MountPoint struct {
	// Type was originally a `Type` string subtype, but since we dont' use any
	// of the mount types, we can just use a plain string.
	Type        string `json:"Type"`
	Source      string `json:",omitempty"`
	Destination string
	RW          bool
}

// Subset of InspectResponse from
// https://github.com/moby/moby/blob/v28.3.3/api/types/container/container.go#L179-L188
type InspectResponse struct {
	Mounts          []MountPoint
	Config          *Config
	NetworkSettings *NetworkSettings
}
