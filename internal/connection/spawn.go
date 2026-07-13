package connection

import (
	"errors"
	"time"

	"github.com/guppshub/cx-cli/internal/state"
)

// Daemon represents a spawned background tunnel daemon.
type Daemon struct {
	binPath      string
	command      string
	resourceName string
	port         int
	logPath      string
	errLogPath   string
	pid          int // PID of the process spawned by Go
}

// LogPath returns the path to the standard daemon output log.
func (d *Daemon) LogPath() string {
	return d.logPath
}

// ErrorLogPath returns the path to the daemon error log.
func (d *Daemon) ErrorLogPath() string {
	return d.errLogPath
}

// VerifyRegistration polls the state store to confirm that the daemon has successfully registered.
// Returns the bound local port, or an error if the timeout is reached.
func (d *Daemon) VerifyRegistration(stateStore *state.Manager, timeout time.Duration) (int, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)

		s, err := stateStore.Load()
		if err != nil {
			continue
		}

		for _, conn := range s.ActiveConnections {
			if d.matches(conn) {
				return conn.LocalPort, nil
			}
		}
	}

	return 0, errors.New("timeout waiting for background daemon to register in state.json")
}
