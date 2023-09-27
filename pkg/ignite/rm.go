package ignite

import "github.com/k0sproject/bootloose/pkg/exec"

// Remove removes an Ignite VM
func Remove(name string) error {
	return exec.CommandWithLogging(execName, "rm", "-f", name)
}
