package aws

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/guppshub/cx-cli/internal/state"
	"github.com/guppshub/cx-cli/internal/tunnel"
)

// ProcessConn wraps a spawned subprocess stdin/stdout streams as a net.Conn.
type ProcessConn struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.Reader
	rawStdout io.ReadCloser
	stderrBuf *bytes.Buffer
	sessionID string
	provider  *Provider
}

// IsAlive checks if the underlying subprocess is still running.
// This relies on DialTunnel's background goroutine calling cmd.Wait(),
// which populates cmd.ProcessState as soon as the child process exits.
func (c *ProcessConn) IsAlive() bool {
	if c.cmd == nil || c.cmd.Process == nil {
		return false
	}
	if c.cmd.ProcessState != nil {
		return false
	}
	return true
}

// Stderr returns any captured stderr output.
func (c *ProcessConn) Stderr() string {
	if c.stderrBuf == nil {
		return ""
	}
	return strings.TrimSpace(c.stderrBuf.String())
}

// FailureMessage drains and returns stdout/stderr of a failed/exited process.
func (c *ProcessConn) FailureMessage() string {
	var buf bytes.Buffer
	// Set a read deadline to prevent potential deadlocks/blocking if the OS pipe is not closed.
	if f, ok := c.rawStdout.(*os.File); ok {
		_ = f.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	}
	// Since the process has exited, we can safely drain the remaining stdout bytes.
	if c.stdout != nil {
		_, _ = io.Copy(&buf, c.stdout)
	}
	msg := strings.TrimSpace(buf.String())

	stderrStr := c.Stderr()
	if stderrStr != "" {
		if msg != "" {
			msg = msg + "\n" + stderrStr
		} else {
			msg = stderrStr
		}
	}
	return msg
}

// Read reads bytes from the process stdout.
func (c *ProcessConn) Read(b []byte) (int, error) {
	return c.stdout.Read(b)
}

// Write writes bytes to the process stdin.
func (c *ProcessConn) Write(b []byte) (int, error) {
	return c.stdin.Write(b)
}

// Close terminates the background process and closes standard streams.
func (c *ProcessConn) Close() error {
	if c.provider != nil && c.sessionID != "" {
		c.provider.terminateSession(c.sessionID)
	}

	_ = c.stdin.Close()
	_ = c.rawStdout.Close()

	// Wait up to 1.5 seconds for the process to exit naturally from EOF on stdin
	if c.cmd != nil && c.cmd.Process != nil {
		for i := 0; i < 30; i++ {
			if c.cmd.ProcessState != nil {
				return nil // Exited cleanly on EOF!
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	killProcessGroup(c.cmd)
	return nil
}

// LocalAddr returns a dummy local address.
func (c *ProcessConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}

// RemoteAddr returns a dummy remote address.
func (c *ProcessConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}

// SetDeadline is a no-op required for net.Conn compatibility.
func (c *ProcessConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline is a no-op required for net.Conn compatibility.
func (c *ProcessConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline is a no-op required for net.Conn compatibility.
func (c *ProcessConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// Provider implements cloud credential verification and network tunneling for AWS.
type Provider struct {
	profile          string
	region           string
	lookPathFunc     func(file string) (string, error)
	checkSessionFunc func(prompt func(string, bool) (string, error)) error
}

// New creates a new AWS Provider.
func New(profile, region string) *Provider {
	return &Provider{
		profile:      profile,
		region:       region,
		lookPathFunc: exec.LookPath,
	}
}

// EnsureCredentials verifies session credentials, calling prompt if authentication is required.
func (p *Provider) EnsureCredentials(ctx context.Context, prompt func(string, bool) (string, error)) error {
	if p.checkSessionFunc != nil {
		return p.checkSessionFunc(prompt)
	}

	args := []string{"sts", "get-caller-identity"}
	if p.profile != "" {
		args = append(args, "--profile", p.profile)
	}
	if p.region != "" {
		args = append(args, "--region", p.region)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderrBuf.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("AWS credential verification failed for profile %q: %s", p.profile, errMsg)
	}

	return nil
}

// terminateSession explicitly terminates the active session on AWS SSM.
func (p *Provider) terminateSession(sessionID string) {
	if sessionID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []string{"ssm", "terminate-session", "--session-id", sessionID}
	if p.profile != "" {
		args = append(args, "--profile", p.profile)
	}
	if p.region != "" {
		args = append(args, "--region", p.region)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()

	logPath, logErr := state.Path()
	if logErr == nil && logPath != "" {
		logDir := filepath.Join(filepath.Dir(logPath), "logs")
		_ = os.MkdirAll(logDir, 0755)
		f, _ := os.OpenFile(filepath.Join(logDir, "session_cleanup.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if f != nil {
			if err != nil {
				_, _ = fmt.Fprintf(f, "[%s] Failed to terminate session %s: %v. Stderr: %s\n", time.Now().Format(time.RFC3339), sessionID, err, stderrBuf.String())
			} else {
				_, _ = fmt.Fprintf(f, "[%s] Successfully terminated session %s: %s\n", time.Now().Format(time.RFC3339), sessionID, strings.TrimSpace(stdoutBuf.String()))
			}
			_ = f.Close()
		}
	}
}

func checkAndResolvePort(port int) int {
	if port <= 0 {
		l, err := net.Listen("tcp", "localhost:0")
		if err == nil {
			defer func() { _ = l.Close() }()
			return l.Addr().(*net.TCPAddr).Port
		}
		return port
	}
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err == nil {
		_ = l.Close()
		return port
	}
	l, err = net.Listen("tcp", "localhost:0")
	if err == nil {
		defer func() { _ = l.Close() }()
		return l.Addr().(*net.TCPAddr).Port
	}
	return port
}

// DialTunnel launches session-manager-plugin in the background to tunnel database traffic.
func (p *Provider) DialTunnel(ctx context.Context, target *tunnel.Target) (net.Conn, error) {
	// Verify dependencies
	if _, err := p.lookPathFunc("aws"); err != nil {
		return nil, fmt.Errorf("aws CLI not found in PATH: %w", err)
	}
	if _, err := p.lookPathFunc("session-manager-plugin"); err != nil {
		return nil, fmt.Errorf("session-manager-plugin not found in PATH: %w", err)
	}

	// Resolve local port binding (fallback to random free port if occupied)
	localPort := checkAndResolvePort(target.PreferredLocalPort)
	target.PreferredLocalPort = localPort

	args := []string{
		"ssm",
		"start-session",
		"--target", target.BastionInstanceID,
		"--document-name", "AWS-StartPortForwardingSessionToRemoteHost",
		"--parameters", fmt.Sprintf("host=%s,portNumber=%d,localPortNumber=%d", target.RemoteHost, target.RemotePort, target.PreferredLocalPort),
	}

	if p.profile != "" {
		args = append(args, "--profile", p.profile)
	}
	if p.region != "" {
		args = append(args, "--region", p.region)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	prepareCmd(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create process stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("failed to create process stdout pipe: %w", err)
	}

	// Capture stderr to read error outputs
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, fmt.Errorf("failed to start ssm tunnel process: %w", err)
	}

	// Scan stdout to ensure the session-manager-plugin initializes successfully
	stdoutReader := bufio.NewReader(stdout)
	var accumulated bytes.Buffer
	success := false
	var sessionID string

	scanCtx, scanCancel := context.WithTimeout(ctx, 5*time.Second)
	defer scanCancel()

	errChan := make(chan error, 1)
	go func() {
		for {
			line, err := stdoutReader.ReadString('\n')
			if len(line) > 0 {
				accumulated.WriteString(line)
				if strings.Contains(line, "SessionId:") {
					parts := strings.Split(line, "SessionId:")
					if len(parts) > 1 {
						sessionID = strings.TrimSpace(parts[1])
					}
				} else if strings.Contains(line, "sessionId ") {
					parts := strings.Split(line, "sessionId ")
					if len(parts) > 1 {
						words := strings.Fields(parts[1])
						if len(words) > 0 {
							sessionID = strings.TrimFunc(words[0], func(r rune) bool {
								return r == '.' || r == ',' || r == '"' || r == '\''
							})
						}
					}
				}
				// Look for standard session-manager-plugin startup success keywords
				if strings.Contains(line, "Waiting for connections") {
					success = true
					errChan <- nil
					return
				}
			}
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	select {
	case <-scanCtx.Done():
		_ = cmd.Process.Kill()
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, fmt.Errorf("connection handshake timed out: %s", strings.TrimSpace(stderrBuf.String()))
	case err := <-errChan:
		if !success {
			_ = cmd.Process.Kill()
			_ = stdin.Close()
			_ = stdout.Close()
			errMsg := strings.TrimSpace(stderrBuf.String())
			if errMsg == "" {
				errMsg = strings.TrimSpace(accumulated.String())
			}
			if errMsg == "" && err != nil {
				errMsg = err.Error()
			}
			return nil, fmt.Errorf("connection handshake failed: %s", errMsg)
		}
	}

	// Start a background goroutine to call Wait() so ProcessState is populated when the process exits
	go func() {
		_ = cmd.Wait()
	}()

	return &ProcessConn{
		cmd:       cmd,
		stdin:     stdin,
		stdout:    stdoutReader,
		rawStdout: stdout,
		stderrBuf: &stderrBuf,
		sessionID: sessionID,
		provider:  p,
	}, nil
}
