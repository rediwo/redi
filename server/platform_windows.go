//go:build windows

package server

import (
	"os/exec"
	"syscall"
)

// setPlatformSpecificAttributes sets Windows-specific process attributes
func setPlatformSpecificAttributes(cmd *exec.Cmd) {
	// Detach from parent console on Windows
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}