//go:build windows

package cmd

import (
	"os/exec"
)

func detachCmd(cmd *exec.Cmd) {
	// Detaching and hiding is handled via PowerShell's -WindowStyle Hidden,
	// so no custom Win32 SysProcAttr flags are needed here.
}
