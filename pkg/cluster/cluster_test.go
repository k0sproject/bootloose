// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchFilter(t *testing.T) {
	const refused = "ssh: connect to host 172.17.0.2 port 22: Connection refused"

	filter := matchFilter{
		writer: io.Discard,
		regexp: connectRefused,
	}

	_, err := filter.Write([]byte("foo\n"))
	require.NoError(t, err)
	assert.Equal(t, false, filter.matched)

	_, err = filter.Write([]byte(refused))
	require.NoError(t, err)
	assert.Equal(t, false, filter.matched)
}

func TestNewClusterWithHostPort(t *testing.T) {
	cluster, err := NewFromYAML([]byte(`cluster:
  name: cluster
  privateKey: cluster-key
machines:
- count: 2
  spec:
    image: quay.io/k0sproject/bootloose-ubuntu20.04:latest
    name: node%d
    portMappings:
    - containerPort: 22
      hostPort: 2222
`))
	require.NoError(t, err)
	require.NotNil(t, cluster)
	assert.Equal(t, 1, len(cluster.spec.Machines))
	template := cluster.spec.Machines[0]
	assert.Equal(t, 2, template.Count)
	assert.Equal(t, 1, len(template.Spec.PortMappings))
	portMapping := template.Spec.PortMappings[0]
	assert.Equal(t, uint16(22), portMapping.ContainerPort)
	assert.Equal(t, uint16(2222), portMapping.HostPort)

	machine0 := cluster.machine(template.Spec, 0)
	args0 := cluster.createMachineRunArgs(machine0, machine0.ContainerName(), 0)
	i := indexOf("-p", args0)
	assert.NotEqual(t, -1, i)
	assert.Equal(t, "2222:22", args0[i+1])

	machine1 := cluster.machine(template.Spec, 1)
	args1 := cluster.createMachineRunArgs(machine1, machine1.ContainerName(), 1)
	i = indexOf("-p", args1)
	assert.NotEqual(t, -1, i)
	assert.Equal(t, "2223:22", args1[i+1])
}

func indexOf(element string, array []string) int {
	for k, v := range array {
		if element == v {
			return k
		}
	}
	return -1 // element not found.
}

func TestCluster_EnsureSSHKeys(t *testing.T) {
	underTest := Cluster{}

	t.Run("do_nothing_if_key_is_empty", func(t *testing.T) {
		assert.NoError(t, underTest.ensureSSHKey())
	})

	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "the-key")

	var pubStat, privStat fs.FileInfo
	created := t.Run("create_new_key", func(t *testing.T) {
		underTest.spec.Cluster.PrivateKey = keyPath

		err := underTest.ensureSSHKey()
		assert.NoError(t, err)

		privStat, err = os.Stat(keyPath)
		if assert.NoError(t, err, "failed to stat private key file") {
			assert.Equal(t, privStat.Mode().Perm(), os.FileMode(0o600), "private key file has wrong permissions")
		}

		pubStat, err = os.Stat(keyPath + ".pub")
		if assert.NoError(t, err, "failed to stat public key file") {
			assert.Equal(t, pubStat.Mode().Perm(), os.FileMode(0o644), "public key file has wrong permissions")
		}
	})

	if !created {
		return
	}

	t.Run("retains_existing_key", func(t *testing.T) {
		err := underTest.ensureSSHKey()
		assert.NoError(t, err)

		newPrivStat, err := os.Stat(keyPath)
		if assert.NoError(t, err, "failed to stat private key file") {
			assert.Equal(t, privStat, newPrivStat, "private key has been tampered with")
		}

		newPubStat, err := os.Stat(keyPath + ".pub")
		if assert.NoError(t, err, "failed to stat public key file") {
			assert.Equal(t, pubStat, newPubStat, "public key has been tampered with")
		}
	})
}
