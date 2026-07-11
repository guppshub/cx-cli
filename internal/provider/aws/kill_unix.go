//go:build !windows

package aws

import (
	"os/exec"
	"syscall"
)

func prepareCmd(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	// The PGID is equal to the leader's PID when Setpgid is true.
	// We send SIGKILL to -Pid to target the entire process group.
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
