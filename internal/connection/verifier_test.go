package connection

import (
	"net"
	"testing"
	"time"
)

func TestVerifyConnectionRedis(t *testing.T) {
	// Start mock Redis server
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to start mock listener: %v", err)
	}
	defer func() { _ = l.Close() }()
	port := l.Addr().(*net.TCPAddr).Port

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		buf := make([]byte, 64)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		if string(buf[:n]) == "*1\r\n$4\r\nPING\r\n" {
			_, _ = conn.Write([]byte("+PONG\r\n"))
		}
	}()

	err = VerifyConnection("redis", port, 2*time.Second)
	if err != nil {
		t.Errorf("expected Redis verification to succeed, got error: %v", err)
	}
}

func TestVerifyConnectionPostgres(t *testing.T) {
	// Start mock Postgres server
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to start mock listener: %v", err)
	}
	defer func() { _ = l.Close() }()
	port := l.Addr().(*net.TCPAddr).Port

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		buf := make([]byte, 8)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		// Expect SSLRequest payload
		if n == 8 && buf[4] == 0x04 && buf[5] == 0xd2 && buf[6] == 0x16 && buf[7] == 0x2f {
			_, _ = conn.Write([]byte("S"))
		}
	}()

	err = VerifyConnection("postgres", port, 2*time.Second)
	if err != nil {
		t.Errorf("expected Postgres verification to succeed, got error: %v", err)
	}
}

func TestVerifyConnectionMySQL(t *testing.T) {
	// Start mock MySQL server
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to start mock listener: %v", err)
	}
	defer func() { _ = l.Close() }()
	port := l.Addr().(*net.TCPAddr).Port

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		// MySQL immediately sends a handshake packet
		_, _ = conn.Write([]byte{0x0a, 0x35, 0x2e, 0x36}) // Prot version + mock version bytes
	}()

	err = VerifyConnection("mysql", port, 2*time.Second)
	if err != nil {
		t.Errorf("expected MySQL verification to succeed, got error: %v", err)
	}
}

func TestVerifyConnectionHTTP(t *testing.T) {
	// Start mock HTTP server
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to start mock listener: %v", err)
	}
	defer func() { _ = l.Close() }()
	port := l.Addr().(*net.TCPAddr).Port

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		buf := make([]byte, 128)
		_, err = conn.Read(buf)
		if err != nil {
			return
		}
		_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"))
	}()

	err = VerifyConnection("http", port, 2*time.Second)
	if err != nil {
		t.Errorf("expected HTTP verification to succeed, got error: %v", err)
	}
}

func TestVerifyConnectionFailure(t *testing.T) {
	// Try to verify connection to a port where nothing is listening
	err := VerifyConnection("redis", 1, 100*time.Millisecond)
	if err == nil {
		t.Error("expected verification to fail for inactive port, but it succeeded")
	}
}
