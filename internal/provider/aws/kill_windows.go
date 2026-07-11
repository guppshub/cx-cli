//go:build windows

package aws

import (
	"fmt"
	"os/exec"
	"syscall"
)

func prepareCmd(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP
}

func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	// Use taskkill to kill the process and all its children (/T) forcefully (/F)
	killCmd := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", cmd.Process.Pid))
	if err := killCmd.Run(); err != nil {
		_ = cmd.Process.Kill()
	}
}
