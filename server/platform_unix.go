//go:build !windows

package server

import (
	"os/exec"
	"syscall"
)

// setPlatformSpecificAttributes sets Unix-specific process attributes
func setPlatformSpecificAttributes(cmd *exec.Cmd) {
	// Create new session to detach from terminal
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}