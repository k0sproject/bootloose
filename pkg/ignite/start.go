package ignite

import (
	"github.com/k0sproject/bootloose/pkg/exec"
)

// Start starts an Ignite VM
func Start(name string) error {
	return exec.CommandWithLogging(execName, "start", name)
}
