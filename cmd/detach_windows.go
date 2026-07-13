//go:build windows

package cmd

import (
	"os/exec"
	"syscall"
)

func detachCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | syscall.DETACHED_PROCESS,
	}
}
