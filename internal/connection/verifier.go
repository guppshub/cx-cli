package connection

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"
)

// VerifyConnection polls the target local port and performs a protocol-specific handshake
// to verify end-to-end connectivity.
func VerifyConnection(protocol string, port int, timeout time.Duration) error {
	addr := fmt.Sprintf("localhost:%d", port)
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		switch strings.ToLower(protocol) {
		case "postgres":
			lastErr = pingPostgres(addr)
		case "redis":
			lastErr = pingRedis(addr)
		case "mysql":
			lastErr = pingMySQL(addr)
		case "http", "opensearch":
			lastErr = pingHTTP(addr)
		default:
			// Fallback: simple TCP connection check
			lastErr = pingTCP(addr)
		}

		if lastErr == nil {
			return nil // Success!
		}
		time.Sleep(1500 * time.Millisecond) // Retry after a short sleep
	}

	return fmt.Errorf("connectivity verification failed after %v: %w", timeout, lastErr)
}

func pingTCP(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

func pingRedis(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))

	// Send standard Redis RESP PING query
	_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	if err != nil {
		return err
	}

	buf := make([]byte, 64)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	resp := string(buf[:n])
	// Valid Redis RESP responses start with +, -, :, $, or *
	if strings.HasPrefix(resp, "+") || strings.HasPrefix(resp, "-") || strings.HasPrefix(resp, ":") || strings.HasPrefix(resp, "$") || strings.HasPrefix(resp, "*") {
		return nil
	}
	return fmt.Errorf("unexpected Redis response: %q", resp)
}

func pingPostgres(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))

	// Send Postgres SSLRequest packet: length 8, code 80877103 (0x04D2162F)
	_, err = conn.Write([]byte{0x00, 0x00, 0x00, 0x08, 0x04, 0xd2, 0x16, 0x2f})
	if err != nil {
		return err
	}

	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err != nil {
		return err
	}

	// Postgres replies with 'S' (Supports SSL) or 'N' (Does not support SSL)
	if buf[0] == 'S' || buf[0] == 'N' {
		return nil
	}
	return fmt.Errorf("unexpected Postgres response: %c", buf[0])
}

func pingMySQL(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))

	buf := make([]byte, 10)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	// MySQL protocol immediately sends a handshake packet on connect (at least 4 bytes)
	if n >= 4 {
		return nil
	}
	return fmt.Errorf("unexpected MySQL handshake length: %d", n)
}

func pingHTTP(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(1 * time.Second))

	// Send basic HTTP request
	_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"))
	if err != nil {
		return err
	}

	buf := make([]byte, 12)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	// Valid HTTP responses start with HTTP/1. or HTTP/2
	if bytes.HasPrefix(buf[:n], []byte("HTTP/1.")) || bytes.HasPrefix(buf[:n], []byte("HTTP/2")) {
		return nil
	}
	return fmt.Errorf("unexpected HTTP response: %q", string(buf[:n]))
}
