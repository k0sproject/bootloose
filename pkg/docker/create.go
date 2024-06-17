/*
Copyright 2018 The Kubernetes Authors.
Copyright 2019 Weaveworks Ltd.
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
	"bytes"
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	goexec "os/exec"

	"github.com/k0sproject/bootloose/pkg/exec"
)

// Create creates a container with "docker create", with some error handling
// it will return the ID of the created container if any, even on error
func Create(image string, runArgs []string, containerArgs []string) (id string, err error) {
	args := []string{"create"}
	args = append(args, runArgs...)
	args = append(args, image)
	args = append(args, containerArgs...)
	cmd := exec.Command("docker", args...)
	cmd.SetEnv("DOCKER_LOG_LEVEL=debug")
	var stdout, stderr bytes.Buffer
	cmd.SetStdout(&stdout)
	cmd.SetStderr(&stderr)

	err = cmd.Run()
	if err != nil {
		// log error output if there was any
		// log error output if there was any
		log.Error(stderr.String())
		return "", err
	}

	// if docker created a container the id will be the first line and match
	// validate the output and get the id
	outputStr := stdout.String()
	outputLines := strings.Split(outputStr, "\n")

	if len(outputLines) < 1 {
		return "", errors.New("failed to get container id, received no output from docker run")
	}
	if !containerIDRegex.MatchString(outputLines[0]) {
		return "", fmt.Errorf("failed to get container id, output did not match: %v", outputLines)
	}
	containerID := outputLines[0]

	// Check container status
	statusCmd := goexec.Command("docker", "ps", "-a", "--filter", "id="+containerID, "--format", "{{.Status}}")
	var statusOut bytes.Buffer
	statusCmd.Stdout = &statusOut
	if err := statusCmd.Run(); err != nil {
		log.Printf("Error checking container status: %s", err)
		return "", err
	}
	log.Printf("Container status: %s", statusOut.String())

	// Capture container logs
	logCmd := goexec.Command("docker", "logs", containerID)
	var logOut bytes.Buffer
	logCmd.Stdout = &logOut
	if err := logCmd.Run(); err != nil {
		log.Printf("Error capturing container logs: %s", err)
		return "", err
	}
	log.Printf("Container logs: %s", logOut.String())

	return containerID, nil
}
