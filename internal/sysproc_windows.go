//go:build windows

package internal

import "os/exec"

// detachProcess is a no-op on Windows; the child process already outlives
// the parent by default.
func detachProcess(cmd *exec.Cmd) {}
