//go:build windows

package connection

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/guppshub/cx-cli/internal/state"
)

// SpawnDaemon launches the background daemon on Windows using a PowerShell hidden wrapper.
func SpawnDaemon(binPath, command, resourceName string, port int, logDir string) (*Daemon, error) {
	logPath := filepath.Join(logDir, resourceName+".log")
	errLogPath := filepath.Join(logDir, resourceName+"_err.log")

	// Start-Process detaches the child process from the parent console session natively.
	// We split stdout and stderr into separate files to avoid sharing locks on Windows.
	psCmd := fmt.Sprintf(
		"Start-Process -FilePath '%s' -ArgumentList '%s', '%s', '--server', '--port', '%d' -WindowStyle Hidden -RedirectStandardOutput '%s' -RedirectStandardError '%s'",
		binPath, command, resourceName, port, logPath, errLogPath,
	)

	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-Command", psCmd)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start powershell launcher: %w", err)
	}

	// Wait for the powershell starter process to exit (it exits quickly after launching the daemon)
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("failed to start background daemon via powershell: %w", err)
	}

	return &Daemon{
		binPath:      binPath,
		command:      command,
		resourceName: resourceName,
		port:         port,
		logPath:      logPath,
		errLogPath:   errLogPath,
		pid:          cmd.Process.Pid, // Note: This is powershell's PID, not the daemon's PID
	}, nil
}

// matches ignores PID check on Windows and performs validation by connection name.
func (d *Daemon) matches(conn *state.ConnectionMetadata) bool {
	return conn.Name == d.resourceName
}
