/*
Copyright 2022 The Kubernetes Authors.
Copyright 2022 Weaveworks Ltd.
Copyright 2023 bootloose authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package docker

import (
	"github.com/k0sproject/bootloose/pkg/exec"
)

// Info return system-wide information
func Info(format string) ([]string, error) {
	cmd := exec.Command("docker", "info",
		"-f", // format
		format,
	)
	return exec.CombinedOutputLines(cmd)
}

// InfoObject is similar to Inspect but deserializes the JSON output to a struct.
func CgroupVersion() string {
	res, err := Info("{{.CgroupVersion}}")
	if err != nil || len(res) == 0 {
		return ""
	}
	return res[0]
}
