//go:build !windows

package connection

import (
	"os"
	"syscall"
)

// IsProcessAlive returns true if the process with the given PID is running.
func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Sending signal 0 to a process checks if it exists and can receive signals.
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}
	// EPERM means the process exists but we do not have permission to send signals to it.
	if err == syscall.EPERM {
		return true
	}
	return false
}
