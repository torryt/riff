//go:build !windows

package internal

import (
	"os/exec"
	"syscall"
)

// detachProcess configures cmd to run in its own process group so it
// survives after the parent exits.
func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
