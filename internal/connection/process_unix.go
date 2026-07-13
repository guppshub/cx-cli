//go:build !windows

package connection

import (
	"os"
	"syscall"
	"time"
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

// TerminateProcessGroup attempts to gracefully terminate a process group on Unix.
// It sends syscall.SIGINT first, waits up to the timeout duration for the process to exit,
// and falls back to syscall.SIGKILL if it's still running.
func TerminateProcessGroup(pid int, timeout time.Duration) {
	if pid <= 0 {
		return
	}
	pgid := -pid
	// 1. Try sending SIGINT first to let the process clean up gracefully
	if err := syscall.Kill(pgid, syscall.SIGINT); err == nil {
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			if !IsProcessAlive(pid) {
				return // Exited cleanly!
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	// 2. Fallback to SIGKILL
	_ = syscall.Kill(pgid, syscall.SIGKILL)
}
