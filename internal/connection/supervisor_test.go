package connection

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// mockConnection implements Connection for testing.
type mockConnection struct {
	pid     int
	port    int
	waitCh  chan struct{}
	waitErr error
	closeMu sync.Mutex
	closed  bool
}

func newMockConnection(pid, port int) *mockConnection {
	return &mockConnection{
		pid:    pid,
		port:   port,
		waitCh: make(chan struct{}),
	}
}

func (m *mockConnection) Wait() error {
	<-m.waitCh
	return m.waitErr
}

func (m *mockConnection) Close() error {
	m.closeMu.Lock()
	defer m.closeMu.Unlock()
	m.closed = true
	// Unblock Wait if not already unblocked
	select {
	case <-m.waitCh:
	default:
		close(m.waitCh)
	}
	return nil
}

func (m *mockConnection) PID() int  { return m.pid }
func (m *mockConnection) Port() int { return m.port }

// simulateExit simulates a process exit with the given error.
func (m *mockConnection) simulateExit(err error) {
	m.waitErr = err
	select {
	case <-m.waitCh:
	default:
		close(m.waitCh)
	}
}

// mockDialer implements Dialer for testing.
type mockDialer struct {
	mu    sync.Mutex
	calls int
	conns []*mockConnection
	errs  []error
}

func (d *mockDialer) Dial(_ context.Context) (Connection, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	idx := d.calls
	d.calls++
	if idx < len(d.errs) && d.errs[idx] != nil {
		return nil, d.errs[idx]
	}
	if idx < len(d.conns) {
		return d.conns[idx], nil
	}
	// Default: return a new mock connection
	c := newMockConnection(1000+idx, 5432)
	return c, nil
}

func (d *mockDialer) callCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.calls
}

// --- Restart Policy Tests ---

func TestIsPermanentError(t *testing.T) {
	tests := []struct {
		err       error
		permanent bool
	}{
		{nil, false},
		{errors.New("network timeout"), false},
		{errors.New("broken pipe"), false},
		{errors.New("AccessDenied: not authorized"), true},
		{errors.New("could not find profile named test"), true},
		{errors.New("ExpiredTokenException: token expired"), true},
		{errors.New("InvalidInstanceId: i-bad"), true},
	}

	for _, tt := range tests {
		got := IsPermanentError(tt.err)
		if got != tt.permanent {
			errStr := "<nil>"
			if tt.err != nil {
				errStr = tt.err.Error()
			}
			t.Errorf("IsPermanentError(%q) = %v, want %v", errStr, got, tt.permanent)
		}
	}
}

func TestFixedBackoff_ShouldRetry(t *testing.T) {
	fb := NewFixedBackoff(1*time.Second, 3)

	// Should allow 3 retries for transient errors
	for i := 0; i < 3; i++ {
		if !fb.ShouldRetry(errors.New("network timeout")) {
			t.Fatalf("expected ShouldRetry to return true on attempt %d", i+1)
		}
	}
	// 4th attempt should be rejected
	if fb.ShouldRetry(errors.New("network timeout")) {
		t.Fatal("expected ShouldRetry to return false after max retries")
	}

	// Reset and verify
	fb.Reset()
	if !fb.ShouldRetry(errors.New("network timeout")) {
		t.Fatal("expected ShouldRetry to return true after Reset()")
	}
}

func TestFixedBackoff_PermanentError(t *testing.T) {
	fb := NewFixedBackoff(1*time.Second, 100)
	if fb.ShouldRetry(errors.New("AccessDenied: not authorized")) {
		t.Fatal("expected ShouldRetry to return false for permanent error")
	}
}

func TestFixedBackoff_NextDelay(t *testing.T) {
	fb := NewFixedBackoff(5*time.Second, 10)
	if fb.NextDelay() != 5*time.Second {
		t.Fatalf("expected 5s delay, got %v", fb.NextDelay())
	}
}

func TestNoRetry(t *testing.T) {
	nr := &NoRetry{}
	if nr.ShouldRetry(errors.New("any error")) {
		t.Fatal("NoRetry.ShouldRetry should always return false")
	}
	if nr.NextDelay() != 0 {
		t.Fatal("NoRetry.NextDelay should return 0")
	}
}

// --- Supervisor Tests ---

func TestSupervisor_StartAndHealthy(t *testing.T) {
	conn := newMockConnection(1234, 5432)
	dialer := &mockDialer{conns: []*mockConnection{conn}}

	var stateChanges []ConnectionState
	var mu sync.Mutex

	sv := NewSupervisor(SupervisorConfig{
		Name:                "test-db",
		Type:                "database",
		Dialer:              dialer,
		Policy:              &NoRetry{},
		HealthCheckInterval: -1,
		OnStateChange: func(meta Metadata) {
			mu.Lock()
			stateChanges = append(stateChanges, meta.State)
			mu.Unlock()
		},
	})

	ctx := context.Background()
	if err := sv.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for the supervisor to reach Healthy
	deadline := time.After(2 * time.Second)
	for sv.State() != StateHealthy {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for Healthy state, got %s", sv.State())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	meta := sv.Meta()
	if meta.PID != 1234 {
		t.Errorf("expected PID 1234, got %d", meta.PID)
	}
	if meta.Port != 5432 {
		t.Errorf("expected Port 5432, got %d", meta.Port)
	}

	sv.Stop()

	mu.Lock()
	defer mu.Unlock()
	// State transitions: Starting -> Healthy -> Stopped
	if len(stateChanges) < 2 {
		t.Fatalf("expected at least 2 state changes, got %d: %v", len(stateChanges), stateChanges)
	}
	if stateChanges[0] != StateStarting {
		t.Errorf("expected first state Starting, got %s", stateChanges[0])
	}
	if stateChanges[1] != StateHealthy {
		t.Errorf("expected second state Healthy, got %s", stateChanges[1])
	}
}

func TestSupervisor_RestartsOnTransientError(t *testing.T) {
	conn1 := newMockConnection(1001, 5432)
	conn2 := newMockConnection(1002, 5432)
	dialer := &mockDialer{conns: []*mockConnection{conn1, conn2}}

	sv := NewSupervisor(SupervisorConfig{
		Name:                "test-db",
		Type:                "database",
		Dialer:              dialer,
		Policy:              NewFixedBackoff(50*time.Millisecond, 5),
		HealthCheckInterval: -1,
	})

	ctx := context.Background()
	if err := sv.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for healthy
	deadline := time.After(2 * time.Second)
	for sv.State() != StateHealthy {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for initial Healthy state")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Simulate connection drop with transient error
	conn1.simulateExit(errors.New("broken pipe"))

	// Wait for second connection to become healthy
	deadline = time.After(5 * time.Second)
	for {
		meta := sv.Meta()
		if meta.Restarts >= 1 && sv.State() == StateHealthy {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for restart, state=%s restarts=%d", sv.State(), meta.Restarts)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	if dialer.callCount() < 2 {
		t.Errorf("expected at least 2 Dial calls, got %d", dialer.callCount())
	}

	sv.Stop()
}

func TestSupervisor_FailsOnPermanentError(t *testing.T) {
	dialer := &mockDialer{
		errs: []error{errors.New("AccessDenied: not authorized")},
	}

	sv := NewSupervisor(SupervisorConfig{
		Name:                "test-db",
		Type:                "database",
		Dialer:              dialer,
		Policy:              NewFixedBackoff(50*time.Millisecond, 5),
		HealthCheckInterval: -1,
	})

	ctx := context.Background()
	if err := sv.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Should transition to Failed
	deadline := time.After(2 * time.Second)
	for sv.State() != StateFailed && sv.State() != StateStopped {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for Failed state, got %s", sv.State())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Should NOT have retried
	if dialer.callCount() != 1 {
		t.Errorf("expected exactly 1 Dial call for permanent error, got %d", dialer.callCount())
	}

	sv.Stop()
}

func TestSupervisor_DoubleStartReturnsError(t *testing.T) {
	conn := newMockConnection(1234, 5432)
	dialer := &mockDialer{conns: []*mockConnection{conn}}

	sv := NewSupervisor(SupervisorConfig{
		Name:                "test-db",
		Type:                "database",
		Dialer:              dialer,
		Policy:              &NoRetry{},
		HealthCheckInterval: -1,
	})

	ctx := context.Background()
	if err := sv.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for healthy
	deadline := time.After(2 * time.Second)
	for sv.State() != StateHealthy {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for Healthy state")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Second Start should fail
	if err := sv.Start(ctx); err == nil {
		t.Fatal("expected error on double Start")
	}

	sv.Stop()
}
