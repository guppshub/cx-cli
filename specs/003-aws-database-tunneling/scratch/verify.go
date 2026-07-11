package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"time"
)

// ProcessConn is a duplicate definition for scratch testing.
type ProcessConn struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (c *ProcessConn) Read(b []byte) (int, error)  { return c.stdout.Read(b) }
func (c *ProcessConn) Write(b []byte) (int, error) { return c.stdin.Write(b) }
func (c *ProcessConn) Close() error {
	_ = c.stdin.Close()
	_ = c.stdout.Close()
	return c.cmd.Process.Kill()
}

func main() {
	action := flag.String("action", "listen", "action to perform: listen, subprocess-mock")
	port := flag.Int("port", 5432, "target local port")
	flag.Parse()

	switch *action {
	case "listen":
		addr := fmt.Sprintf("localhost:%d", *port)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Printf("Port %d in use, falling back to random port...\n", *port)
			l, err = net.Listen("tcp", "localhost:0")
			if err != nil {
				log.Fatalf("Error binding fallback port: %v", err)
			}
		}
		defer func() { _ = l.Close() }()

		fmt.Printf("Successfully bound to %s\n", l.Addr().String())

		// Keep alive for 3 seconds for tests
		time.Sleep(3 * time.Second)

	case "subprocess-mock":
		// Mock by spawning cat which echoes standard input
		cmd := exec.Command("cat")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatalf("Error creating stdin pipe: %v", err)
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("Error creating stdout pipe: %v", err)
		}

		if err := cmd.Start(); err != nil {
			log.Fatalf("Error starting mock process: %v", err)
		}

		conn := &ProcessConn{
			cmd:    cmd,
			stdin:  stdin,
			stdout: stdout,
		}
		defer func() { _ = conn.Close() }()

		message := []byte("hello network stream\n")
		_, err = conn.Write(message)
		if err != nil {
			log.Fatalf("Error writing to process conn: %v", err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("Error reading from process conn: %v", err)
		}

		fmt.Printf("Received echo: %s", string(buf[:n]))

	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}
