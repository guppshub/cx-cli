package tunnel

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// Controller binds to a local port and proxies client connections to the target via the dialer.
type Controller struct {
	dialer    Dialer
	localAddr net.Addr
	mu        sync.RWMutex
}

// NewController creates a new Controller instance.
func NewController(dialer Dialer) *Controller {
	return &Controller{
		dialer: dialer,
	}
}

// LocalAddress returns the bound network address of the controller.
func (c *Controller) LocalAddress() net.Addr {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.localAddr
}

// Start binds to the local port and proxies connection streams.
func (c *Controller) Start(ctx context.Context, localPort int, target *Target) error {
	addr := fmt.Sprintf("localhost:%d", localPort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		// Port conflict fallback
		log.Printf("Port %d in use, falling back to random open port...\n", localPort)
		l, err = net.Listen("tcp", "localhost:0")
		if err != nil {
			return fmt.Errorf("failed to bind fallback port: %w", err)
		}
	}

	c.mu.Lock()
	c.localAddr = l.Addr()
	c.mu.Unlock()

	// Watch for context cancellation to close the listener
	go func() {
		<-ctx.Done()
		_ = l.Close()
	}()

	for {
		clientConn, err := l.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				log.Printf("Error accepting client connection: %v\n", err)
				continue
			}
		}

		if tcpConn, ok := clientConn.(*net.TCPConn); ok {
			_ = tcpConn.SetKeepAlive(true)
			_ = tcpConn.SetKeepAlivePeriod(1 * time.Minute)
		}

		go c.handleClient(ctx, clientConn, target)
	}
}

func (c *Controller) handleClient(ctx context.Context, clientConn net.Conn, target *Target) {
	defer func() { _ = clientConn.Close() }()

	tunnelConn, err := c.dialer.DialTunnel(ctx, target)
	if err != nil {
		log.Printf("Error dialing tunnel target: %v\n", err)
		return
	}
	defer func() { _ = tunnelConn.Close() }()

	var wg sync.WaitGroup
	wg.Add(2)

	// Pipe client to tunnel
	go func() {
		defer wg.Done()
		_, _ = io.Copy(tunnelConn, clientConn)
		_ = tunnelConn.SetDeadline(time.Now()) // Unblock Read on other side
	}()

	// Pipe tunnel to client
	go func() {
		defer wg.Done()
		_, _ = io.Copy(clientConn, tunnelConn)
		_ = clientConn.SetDeadline(time.Now()) // Unblock Read on other side
	}()

	wg.Wait()
}
