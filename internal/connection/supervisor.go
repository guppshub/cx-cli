package connection

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// SupervisorConfig holds configuration for a Supervisor instance.
type SupervisorConfig struct {
	// Name is the human-readable resource name (e.g. "sequr", "mercury").
	Name string
	// Type is the resource type (e.g. "database", "redis").
	Type string
	// Dialer creates new connections to the target resource.
	Dialer Dialer
	// Policy governs retry behavior after connection failures.
	Policy RestartPolicy
	// Logger receives structured lifecycle events. If nil, log.Default() is used.
	Logger *log.Logger
	// OnStateChange is called whenever the supervisor state transitions.
	// It receives the new state and current metadata snapshot.
	OnStateChange func(meta Metadata)
	// HealthCheckInterval is the interval between active TCP health checks.
	// If 0, a default of 5 seconds is used. If negative, active health checks are disabled.
	HealthCheckInterval time.Duration
}

// Metadata holds runtime information about a supervised connection.
type Metadata struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	State       ConnectionState `json:"state"`
	Port        int             `json:"port"`
	PID         int             `json:"pid"`
	Restarts    int             `json:"restarts"`
	StartedAt   time.Time       `json:"started_at"`
	LastFailure string          `json:"last_failure"`
	LastRestart time.Time       `json:"last_restart"`
	SessionID   string          `json:"session_id"`
}

// Supervisor manages the complete lifecycle of a long-running connection.
type Supervisor struct {
	cfg  SupervisorConfig
	meta Metadata

	mu       sync.Mutex
	conn     Connection
	state    ConnectionState
	stopChan chan struct{}
	stopped  chan struct{}
}

// NewSupervisor creates a new Supervisor with the given configuration.
func NewSupervisor(cfg SupervisorConfig) *Supervisor {
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	if cfg.Policy == nil {
		cfg.Policy = &NoRetry{}
	}
	return &Supervisor{
		cfg:      cfg,
		state:    StateStopped,
		stopChan: make(chan struct{}),
		stopped:  make(chan struct{}),
		meta: Metadata{
			Name:  cfg.Name,
			Type:  cfg.Type,
			State: StateStopped,
		},
	}
}

// Start begins the supervisor loop. It dials the initial connection and then
// monitors it, restarting on transient failures according to the restart policy.
// Start returns immediately; the supervisor runs in a background goroutine.
// Call Stop() to terminate it.
func (s *Supervisor) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.state != StateStopped {
		s.mu.Unlock()
		return fmt.Errorf("supervisor already running (state: %s)", s.state)
	}
	s.setState(StateStarting)
	s.meta.StartedAt = time.Now().UTC()
	s.mu.Unlock()

	go s.run(ctx)
	return nil
}

// Stop gracefully shuts down the supervisor and its active connection.
func (s *Supervisor) Stop() {
	s.mu.Lock()
	if s.state == StateStopped {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	close(s.stopChan)
	<-s.stopped
}

// State returns the current supervisor state.
func (s *Supervisor) State() ConnectionState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// Meta returns a snapshot of the current connection metadata.
func (s *Supervisor) Meta() Metadata {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.meta
}

// Wait blocks until the supervisor exits (either via Stop or permanent failure).
func (s *Supervisor) Wait() {
	<-s.stopped
}

// Done returns a channel that is closed when the supervisor exits.
func (s *Supervisor) Done() <-chan struct{} {
	return s.stopped
}

// setState transitions state and triggers the callback.
func (s *Supervisor) setState(newState ConnectionState) {
	s.state = newState
	s.meta.State = newState
	if s.cfg.OnStateChange != nil {
		s.cfg.OnStateChange(s.meta)
	}
}

// run is the main supervisor loop.
func (s *Supervisor) run(ctx context.Context) {
	defer close(s.stopped)
	defer func() {
		s.mu.Lock()
		s.setState(StateStopped)
		s.mu.Unlock()
	}()

	for {
		// Check for stop signal before dialing
		select {
		case <-s.stopChan:
			return
		default:
		}

		// Dial a new connection
		s.cfg.Logger.Printf("[%s] Dialing connection...", s.cfg.Name)
		conn, err := s.cfg.Dialer.Dial(ctx)
		if err != nil {
			if !s.handleDialError(err) {
				return
			}
			continue
		}

		// Successfully connected
		s.mu.Lock()
		s.conn = conn
		s.meta.PID = conn.PID()
		s.meta.Port = conn.Port()
		s.meta.StartedAt = time.Now().UTC()
		if si, ok := conn.(interface{ SessionID() string }); ok {
			s.meta.SessionID = si.SessionID()
		}
		s.setState(StateHealthy)
		s.cfg.Policy.Reset()
		s.mu.Unlock()

		s.cfg.Logger.Printf("[%s] Connection healthy (PID: %d, Port: %d)", s.cfg.Name, conn.PID(), conn.Port())

		// Monitor: wait for connection exit or stop signal
		exitErr := s.monitor(conn)

		// Clean up the dead connection
		_ = conn.Close()
		s.mu.Lock()
		s.conn = nil
		s.mu.Unlock()

		if exitErr == nil {
			// Process exited cleanly (e.g. via Stop)
			return
		}

		// Process exited with error — attempt restart
		if !s.handleDialError(exitErr) {
			return
		}
	}
}

// monitor blocks until the connection exits, the supervisor is stopped, or the port active check fails.
// Returns nil if stopped by the user, or the exit/health error if the connection dropped.
func (s *Supervisor) monitor(conn Connection) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- conn.Wait()
	}()

	interval := s.cfg.HealthCheckInterval
	if interval == 0 {
		interval = 5 * time.Second
	}

	var tickerCh <-chan time.Time
	if interval > 0 {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		tickerCh = ticker.C
	}

	for {
		select {
		case <-s.stopChan:
			// Graceful shutdown requested
			_ = conn.Close()
			return nil
		case err := <-errChan:
			// Check if we were stopped by the user
			select {
			case <-s.stopChan:
				return nil
			default:
				if err == nil {
					return fmt.Errorf("tunnel process exited unexpectedly")
				}
				return err
			}
		case <-tickerCh:
			// Run active TCP liveness check to handle silent session drops/zombies
			if !s.checkLiveness(conn) {
				return fmt.Errorf("active health check failed: port %d not responding", conn.Port())
			}
		}
	}
}

// checkLiveness performs a lightweight TCP dial check to ensure the local port is open.
func (s *Supervisor) checkLiveness(conn Connection) bool {
	address := fmt.Sprintf("127.0.0.1:%d", conn.Port())
	d := net.Dialer{Timeout: 1 * time.Second}
	c, err := d.Dial("tcp", address)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

// handleDialError processes a connection failure and decides whether to retry.
// Returns true if the supervisor should retry, false if it should stop.
func (s *Supervisor) handleDialError(err error) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.meta.LastFailure = err.Error()
	s.cfg.Logger.Printf("[%s] Connection failed: %v", s.cfg.Name, err)

	if !s.cfg.Policy.ShouldRetry(err) {
		s.cfg.Logger.Printf("[%s] Permanent failure or retries exhausted. Giving up.", s.cfg.Name)
		s.setState(StateFailed)
		return false
	}

	s.meta.Restarts++
	s.meta.LastRestart = time.Now().UTC()
	s.setState(StateRestarting)

	delay := s.cfg.Policy.NextDelay()
	s.cfg.Logger.Printf("[%s] Restarting in %s (attempt %d)...", s.cfg.Name, delay, s.meta.Restarts)

	// Wait for delay or stop signal
	s.mu.Unlock()
	select {
	case <-s.stopChan:
		s.mu.Lock()
		return false
	case <-time.After(delay):
	}
	s.mu.Lock()

	s.setState(StateStarting)
	return true
}
