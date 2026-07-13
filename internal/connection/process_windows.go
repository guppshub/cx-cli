//go:build windows

package connection

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

// IsProcessAlive returns true if the process with the given PID is running.
func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	const PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	h, err := syscall.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		if err == syscall.Errno(syscall.ERROR_ACCESS_DENIED) {
			return true
		}
		return false
	}
	defer syscall.CloseHandle(h)

	var exitCode uint32
	err = syscall.GetExitCodeProcess(h, &exitCode)
	if err != nil {
		return false
	}
	const STILL_ACTIVE = 259
	return exitCode == STILL_ACTIVE
}

// TerminateProcessGroup attempts to gracefully terminate a process group on Windows.
// It tries a soft taskkill (without /F) first, waits up to the timeout duration,
// and falls back to a forceful taskkill (/F) if it's still running.
func TerminateProcessGroup(pid int, timeout time.Duration) {
	if pid <= 0 {
		return
	}
	// Try soft kill first to let the process clean up gracefully
	killCmd := exec.Command("taskkill", "/T", "/PID", fmt.Sprintf("%d", pid))
	_ = killCmd.Run()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !IsProcessAlive(pid) {
			return // Exited cleanly!
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Fallback to force kill
	killCmdForce := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", pid))
	_ = killCmdForce.Run()
}
