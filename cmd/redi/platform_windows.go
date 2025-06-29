//go:build windows

package main

import (
	"os/exec"
)

func setPlatformSpecificAttributes(cmd *exec.Cmd) {
	// Windows doesn't need Setsid, process management is different
	// The process will be detached by default when parent exits
}