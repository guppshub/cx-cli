package aws

import (
	"context"
	"io"
	"os/exec"
	"testing"

	"github.com/guppshub/cx-cli/internal/tunnel"
)

func TestProcessConn(t *testing.T) {
	// Spawns cat which echoes standard input back to stdout
	cmd := exec.Command("cat")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start cat process: %v", err)
	}

	conn := &ProcessConn{
		cmd:       cmd,
		stdin:     stdin,
		stdout:    stdout,
		rawStdout: stdout,
	}
	defer func() { _ = conn.Close() }()

	// Verify Write/Read
	msg := []byte("hello conn wrapper\n")
	n, err := conn.Write(msg)
	if err != nil || n != len(msg) {
		t.Fatalf("failed to write to ProcessConn: %v", err)
	}

	buf := make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("failed to read from ProcessConn: %v", err)
	}

	if string(buf[:n]) != string(msg) {
		t.Errorf("expected echo %q, got %q", msg, string(buf[:n]))
	}

	// Verify Close kills process
	err = conn.Close()
	if err != nil {
		t.Fatalf("failed to close connection: %v", err)
	}

	// Process should be terminated
	_ = cmd.Wait()
}

func TestEnsureCredentialsWithPrompter(t *testing.T) {
	provider := New("default", "us-east-1")
	promptCalled := false
	prompter := func(prompt string, secret bool) (string, error) {
		promptCalled = true
		return "123456", nil
	}

	// For unit tests, we want to mock the credentials lookup or run it offline.
	// We can let p.EnsureCredentials call our mock if we have a mock flag or if we just test the prompt mechanism.
	// Let's implement the prompt flow in provider.go and test it by setting an internal flag or test hook!
	// We'll define a test hook in provider: p.checkSessionFunc = func(...) ...
	provider.checkSessionFunc = func(prompt func(string, bool) (string, error)) error {
		_, _ = prompt("Enter MFA Code: ", false)
		return nil
	}

	err := provider.EnsureCredentials(context.Background(), prompter)
	if err != nil {
		t.Fatalf("unexpected error during EnsureCredentials: %v", err)
	}

	if !promptCalled {
		t.Error("expected prompter callback to be invoked, but it wasn't")
	}
}

func TestDialTunnelLookPathChecks(t *testing.T) {
	provider := New("default", "us-east-1")

	// Set LookPath mock to simulate missing binaries
	provider.lookPathFunc = func(file string) (string, error) {
		return "", exec.ErrNotFound
	}

	target := &tunnel.Target{
		BastionInstanceID: "i-012345",
		RemoteHost:        "rds.staging.local",
		RemotePort:        5432,
	}

	_, err := provider.DialTunnel(context.Background(), target)
	if err == nil {
		t.Error("expected DialTunnel to fail when binaries are missing, but it succeeded")
	}
}

func TestFetchEC2InstancesLookPathChecks(t *testing.T) {
	provider := New("default", "us-east-1")
	provider.lookPathFunc = func(file string) (string, error) {
		return "", exec.ErrNotFound
	}

	_, err := provider.FetchEC2Instances(context.Background())
	if err == nil {
		t.Error("expected FetchEC2Instances to fail when aws CLI is missing, but it succeeded")
	}
}
