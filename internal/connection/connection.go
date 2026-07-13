package connection

import (
	"context"
	"errors"
)

// ErrInvalidConnection is returned when a dialed connection cannot be cast to the expected type.
var ErrInvalidConnection = errors.New("invalid connection type returned by dialer")

// ConnectionState represents the lifecycle state of a supervised connection.
type ConnectionState string

const (
	StateStopped    ConnectionState = "Stopped"
	StateStarting   ConnectionState = "Starting"
	StateHealthy    ConnectionState = "Healthy"
	StateRestarting ConnectionState = "Restarting"
	StateFailed     ConnectionState = "Failed"
)

// Connection represents a running, supervised process connection (e.g., an SSM tunnel).
type Connection interface {
	// Wait blocks until the underlying process exits and returns the exit error.
	Wait() error
	// Close terminates the underlying process and all its children cleanly.
	Close() error
	// PID returns the OS process ID.
	PID() int
	// Port returns the bound local port.
	Port() int
}

// Dialer creates new connections. Implementations are provider-specific (e.g., AWS SSM).
type Dialer interface {
	// Dial establishes a new connection to the target resource.
	Dial(ctx context.Context) (Connection, error)
}
