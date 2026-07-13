package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/guppshub/cx-cli/internal/connection"
	"github.com/guppshub/cx-cli/internal/tunnel"
)

// TunnelConnection wraps a ProcessConn to satisfy the connection.Connection interface.
type TunnelConnection struct {
	proc     *ProcessConn
	port     int
	waitDone chan struct{}
}

// newTunnelConnection wraps an existing ProcessConn into a Connection.
// NOTE: DialTunnel already starts a goroutine calling cmd.Wait(), so we must NOT
// call cmd.Wait() again. Instead we poll cmd.ProcessState to detect process exit.
func newTunnelConnection(proc *ProcessConn, port int) *TunnelConnection {
	tc := &TunnelConnection{
		proc:     proc,
		port:     port,
		waitDone: make(chan struct{}),
	}
	// Monitor process exit by checking ProcessState, which is populated by
	// the goroutine in DialTunnel that already calls cmd.Wait().
	go func() {
		for tc.proc.cmd != nil && tc.proc.cmd.ProcessState == nil {
			time.Sleep(50 * time.Millisecond)
		}
		close(tc.waitDone)
	}()
	return tc
}

// Wait blocks until the underlying process exits.
func (tc *TunnelConnection) Wait() error {
	<-tc.waitDone
	if tc.proc.cmd != nil && tc.proc.cmd.ProcessState != nil {
		if !tc.proc.cmd.ProcessState.Success() {
			return fmt.Errorf("process exited with code %d", tc.proc.cmd.ProcessState.ExitCode())
		}
	}
	return nil
}

// Close terminates the process group cleanly.
func (tc *TunnelConnection) Close() error {
	return tc.proc.Close()
}

// PID returns the OS process ID.
func (tc *TunnelConnection) PID() int {
	if tc.proc.cmd != nil && tc.proc.cmd.Process != nil {
		return tc.proc.cmd.Process.Pid
	}
	return 0
}

// Port returns the bound local port.
func (tc *TunnelConnection) Port() int {
	return tc.port
}

// SessionID returns the AWS SSM session ID.
func (tc *TunnelConnection) SessionID() string {
	if tc.proc != nil {
		return tc.proc.sessionID
	}
	return ""
}

// Compile-time check that TunnelConnection satisfies Connection.
var _ connection.Connection = (*TunnelConnection)(nil)

// TunnelDialer implements connection.Dialer by calling DialTunnel on the AWS provider.
type TunnelDialer struct {
	provider *Provider
	target   *tunnel.Target
}

// NewTunnelDialer creates a Dialer that establishes AWS SSM port-forwarding tunnels.
func NewTunnelDialer(provider *Provider, target *tunnel.Target) *TunnelDialer {
	return &TunnelDialer{provider: provider, target: target}
}

// Dial establishes a new tunnel connection.
func (d *TunnelDialer) Dial(ctx context.Context) (connection.Connection, error) {
	conn, err := d.provider.DialTunnel(ctx, d.target)
	if err != nil {
		return nil, err
	}
	proc, ok := conn.(*ProcessConn)
	if !ok {
		_ = conn.Close()
		return nil, connection.ErrInvalidConnection
	}
	return newTunnelConnection(proc, d.target.PreferredLocalPort), nil
}

// Compile-time check that TunnelDialer satisfies Dialer.
var _ connection.Dialer = (*TunnelDialer)(nil)
