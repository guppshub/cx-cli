//go:build windows

package cmd

import (
	"os/exec"
	"syscall"
)

const detachedProcess = 0x00000008

func detachCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | detachedProcess,
	}
}
