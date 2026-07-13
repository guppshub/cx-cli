//go:build !windows

package connection

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/guppshub/cx-cli/internal/state"
)

// SpawnDaemon launches the background daemon on Unix-like operating systems.
func SpawnDaemon(binPath, command, resourceName string, port int, logDir string) (*Daemon, error) {
	logPath := filepath.Join(logDir, resourceName+".log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	// We close our write handle in the parent process after launching the child.
	defer func() { _ = logFile.Close() }()

	args := []string{command, resourceName, "--server", "--port", fmt.Sprint(port)}
	cmd := exec.Command(binPath, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Configure Unix-specific process detachment (new process group PGID)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start background daemon: %w", err)
	}

	return &Daemon{
		binPath:      binPath,
		command:      command,
		resourceName: resourceName,
		port:         port,
		logPath:      logPath,
		errLogPath:   logPath, // Unix logs both stdout/stderr to the same file
		pid:          cmd.Process.Pid,
	}, nil
}

// matches performs strict name and PID validation on Unix.
func (d *Daemon) matches(conn *state.ConnectionMetadata) bool {
	return conn.Name == d.resourceName && conn.Pid == d.pid
}
