// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func NewConfigFromYAML(data []byte) (*Config, error) {
	spec := Config{}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

func NewConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewConfigFromYAML(data)
}

// MachineReplicas are a number of machine following the same specification.
type MachineReplicas struct {
	Spec  *Machine `json:"spec"`
	Count int      `json:"count"`
}

// Cluster is a set of Machines.
type Cluster struct {
	// Name is the cluster name. Defaults to "cluster".
	Name string `json:"name"`

	// PrivateKey is the path to the private SSH key used to login into the cluster
	// machines. Can be expanded to user homedir if ~ is found. Ex. ~/.ssh/id_rsa.
	//
	// This field is optional. If absent, machines are expected to have a public
	// key defined.
	PrivateKey string `json:"privateKey,omitempty"`
}

// Config is the top level config object.
type Config struct {
	// Cluster describes cluster-wide configuration.
	Cluster Cluster `json:"cluster"`
	// Machines describe the machines we want created for this cluster.
	Machines []MachineReplicas `json:"machines"`
}

// validate checks basic rules for MachineReplicas's fields
func (conf MachineReplicas) validate() error {
	return conf.Spec.validate()
}

// Validate checks basic rules for Config's fields
func (conf Config) Validate() error {
	valid := true
	for _, machine := range conf.Machines {
		err := machine.validate()
		if err != nil {
			valid = false
			log.Fatalf(err.Error())
		}
	}
	if !valid {
		return fmt.Errorf("Configuration file non valid")
	}
	return nil
}

func DefaultConfig() Config {
	return Config{
		Cluster: Cluster{
			Name:       "cluster",
			PrivateKey: "cluster-key",
		},
		Machines: []MachineReplicas{
			{
				Count: 1,
				Spec: &Machine{
					Name:  "node%d",
					Image: "quay.io/k0sproject/bootloose-ubuntu20.04",
					PortMappings: []PortMapping{
						{ContainerPort: 22},
					},
				},
			},
		},
	}
}
