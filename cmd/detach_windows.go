//go:build windows

package cmd

import (
	"os/exec"
	"syscall"
)

const (
	createNewConsole = 0x00000010
	createNoWindow   = 0x08000000
)

func detachCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | createNewConsole | createNoWindow,
	}
}
