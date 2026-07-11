//go:build windows

package connection

import (
	"syscall"
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
