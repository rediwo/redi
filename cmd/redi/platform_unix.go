//go:build unix

package main

import (
	"os/exec"
	"syscall"
)

func setPlatformSpecificAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}