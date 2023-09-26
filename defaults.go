package main

import "github.com/k0sproject/footloose/pkg/config"

// defaultKeyStore is the path where to store the public keys.
const defaultKeyStorePath = "keys"

var defaultConfig = config.Config{
	Cluster: config.Cluster{
		Name:       "cluster",
		PrivateKey: "cluster-key",
	},
	Machines: []config.MachineReplicas{{
		Count: 1,
		Spec: config.Machine{
			Name:  "node%d",
			Image: "quay.io/k0sproject/footloose-ubuntu20.04",
			PortMappings: []config.PortMapping{{
				ContainerPort: 22,
			}},
			Backend: "docker",
		},
	}},
}
