package tunnel

import (
	"context"
	"net"
)

// Dialer establishes network connection tunnels to target endpoints.
type Dialer interface {
	DialTunnel(ctx context.Context, target *Target) (net.Conn, error)
}
