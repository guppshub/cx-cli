package connection

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/guppshub/cx-cli/internal/state"
	"github.com/guppshub/cx-cli/internal/tunnel"
)

// Manager handles the lifecycle of connections (registration, duplicates, liveness checks, and handshakes).
type Manager struct {
	stateStore *state.Manager
}

// NewManager creates a new Manager instance.
func NewManager(stateStore *state.Manager) *Manager {
	return &Manager{
		stateStore: stateStore,
	}
}

// RegisterState registers an active connection in state.json.
func (m *Manager) RegisterState(resourceType, name string, port int, connID string) error {
	s, err := m.stateStore.Load()
	if err != nil {
		return err
	}

	s.ActiveConnections[connID] = &state.ConnectionMetadata{
		Type:         resourceType,
		Name:         name,
		LocalPort:    port,
		ConnectionID: connID,
		ConnectedAt:  time.Now().Format(time.RFC3339),
		Pid:          os.Getpid(),
	}

	return m.stateStore.Save(s)
}

// DeregisterState removes an active connection from state.json.
func (m *Manager) DeregisterState(connID string) error {
	s, err := m.stateStore.Load()
	if err != nil {
		return err
	}

	delete(s.ActiveConnections, connID)
	return m.stateStore.Save(s)
}

// GetActiveConnection checks if an active and alive connection already exists for the resource.
func (m *Manager) GetActiveConnection(resourceName string) (*state.ConnectionMetadata, error) {
	s, err := m.stateStore.Load()
	if err != nil {
		return nil, err
	}

	for _, conn := range s.ActiveConnections {
		if conn.Name == resourceName && IsProcessAlive(conn.Pid) {
			return conn, nil
		}
	}

	return nil, nil
}

// PreflightHandshake verifies that the target destination is reachable before starting the session.
func (m *Manager) PreflightHandshake(ctx context.Context, dialer tunnel.Dialer, target *tunnel.Target, protocol string) error {
	// Use a random port during handshake to prevent blocking/occupying the target port
	handshakeTarget := *target
	handshakeTarget.PreferredLocalPort = 0

	handshakeCtx, handshakeCancel := context.WithTimeout(ctx, 15*time.Second)
	testConn, err := dialer.DialTunnel(handshakeCtx, &handshakeTarget)
	handshakeCancel()
	if err != nil {
		return fmt.Errorf("connection handshake failed: %w", err)
	}
	defer func() { _ = testConn.Close() }()

	// Verify connection over the bound port using our native verifier
	err = VerifyConnection(protocol, handshakeTarget.PreferredLocalPort, 10*time.Second)
	if err != nil {
		// If verification fails, read any exit/failure output of the process if available
		type failureConn interface {
			FailureMessage() string
		}
		if fc, ok := testConn.(failureConn); ok {
			if errMsg := fc.FailureMessage(); errMsg != "" {
				return fmt.Errorf("connection verification failed:\n%s", errMsg)
			}
		}
		return fmt.Errorf("connection verification failed: %w", err)
	}

	return nil
}
