// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"net"

	"github.com/k0sproject/bootloose/pkg/api/docker/network"
)

const (
	ipv4Length = 32
)

// NewRuntimeNetworks returns a slice of networks
func NewRuntimeNetworks(networks map[string]*network.EndpointSettings) []*RuntimeNetwork {
	rnList := make([]*RuntimeNetwork, 0, len(networks))
	for key, value := range networks {
		mask := net.CIDRMask(value.IPPrefixLen, ipv4Length)
		maskIP := net.IP(mask).String()
		rnNetwork := &RuntimeNetwork{
			Name:    key,
			IP:      value.IPAddress,
			Mask:    maskIP,
			Gateway: value.Gateway,
		}
		rnList = append(rnList, rnNetwork)
	}
	return rnList
}

// RuntimeNetwork contains information about the network
type RuntimeNetwork struct {
	// Name of the network
	Name string `json:"name,omitempty"`
	// IP of the container
	IP string `json:"ip,omitempty"`
	// Mask of the network
	Mask string `json:"mask,omitempty"`
	// Gateway of the network
	Gateway string `json:"gateway,omitempty"`
}
