// SPDX-FileCopyrightText: 2025 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package network

// Subset of EndpointSettings
// https://github.com/moby/moby/blob/v28.3.3/api/types/network/endpoint.go#L12

type EndpointSettings struct {
	Gateway     string
	IPAddress   string
	IPPrefixLen int
}
