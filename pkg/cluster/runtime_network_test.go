// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2025 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"testing"

	"github.com/k0sproject/bootloose/pkg/api/docker/network"
	"github.com/stretchr/testify/assert"
)

func TestNewRuntimeNetworks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		networks := map[string]*network.EndpointSettings{}
		networks["mynetwork"] = &network.EndpointSettings{
			Gateway:     "172.17.0.1",
			IPAddress:   "172.17.0.4",
			IPPrefixLen: 16,
		}
		res := NewRuntimeNetworks(networks)

		expectedRuntimeNetworks := []*RuntimeNetwork{
			&RuntimeNetwork{Name: "mynetwork", Gateway: "172.17.0.1", IP: "172.17.0.4", Mask: "255.255.0.0"}}
		assert.Equal(t, expectedRuntimeNetworks, res)
	})
}
