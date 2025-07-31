// SPDX-FileCopyrightText: 2025 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/docker/go-connections/nat"
	"github.com/k0sproject/bootloose/pkg/api/docker/network"
)

// Subset of NetworkSettings from
// https://github.com/moby/moby/blob/v28.3.3/api/types/container/network_settings.go#L8
type NetworkSettings struct {
	Ports     nat.PortMap
	IPAddress string
	Networks  map[string]*network.EndpointSettings
}
