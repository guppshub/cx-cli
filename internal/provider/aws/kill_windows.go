//go:build windows

package aws

import (
	"os/exec"
	"syscall"
	"time"

	"github.com/guppshub/cx-cli/internal/connection"
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
	connection.TerminateProcessGroup(cmd.Process.Pid, 2000*time.Millisecond)
}
