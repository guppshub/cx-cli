# Research: Connection Supervisor State Machine & Interfaces

This document outlines the core interfaces, state transitions, and process synchronization logic for the Connection Supervisor.

---

## 1. ConnectionState Machine

The supervisor transition table is defined as follows:

| From State | Event | Action | To State |
|------------|-------|--------|----------|
| **Stopped** | `Start()` called | Initialize state, invoke dialer | **Starting** |
| **Starting**| Dial success & health PASS | Register in `state.json`, spawn waiter | **Healthy** |
| **Starting**| Dial failure (Transient) | Increment restarts, trigger backoff | **Restarting**|
| **Starting**| Dial failure (Permanent) | Log error, stop retries | **Failed** |
| **Healthy** | `Wait()` returns error / TCP fails | Log failure, check policy, trigger backoff | **Restarting**|
| **Healthy** | `Stop()` called | Terminate process group, deregister state | **Stopped** |
| **Restarting**| Dial success & health PASS | Reset restart policy, update state | **Healthy** |
| **Restarting**| Dial failure & policy allows | Increment restarts, trigger backoff | **Restarting**|
| **Restarting**| Dial failure & policy denies | Log exhaust, stop retries | **Failed** |
| **Restarting**| `Stop()` called | Terminate active dial, deregister state | **Stopped** |
| **Failed**   | `Stop()` called | Deregister state | **Stopped** |

---

## 2. Core Go Interfaces

### A. Connection State Enum
```go
type ConnectionState string

const (
	StateStopped    ConnectionState = "Stopped"
	StateStarting   ConnectionState = "Starting"
	StateHealthy    ConnectionState = "Healthy"
	StateRestarting ConnectionState = "Restarting"
	StateFailed     ConnectionState = "Failed"
)
```

### B. Dialer & Connection Interfaces
The supervisor delegates the creation of the connection to a `Dialer`. The returned `Connection` wraps the actual subprocess.

```go
type Connection interface {
	Wait() error   // Blocks until process exits, returns exit code error
	Close() error  // Terminates the process group cleanly
	PID() int      // Returns the OS process ID
	Port() int     // Returns the bound local port
}

type Dialer interface {
	Dial(ctx context.Context) (Connection, error)
}
```

### C. Restart Policy Interface
Governs whether to retry and the backoff duration.

```go
type RestartPolicy interface {
	ShouldRetry(err error) bool // Returns false on permanent errors or limit exhaust
	NextDelay() time.Duration   // Returns backoff delay
	Reset()                     // Resets the restart counter and delay
}
```

---

## 3. Event-Driven Loop (No Polling)

Instead of running a sleep/polling loop, the supervisor monitors the connection lifecycle by launching a goroutine that blocks on `conn.Wait()`:

```go
func (s *supervisor) monitorConnection(conn Connection) {
	errChan := make(chan error, 1)
	go func() {
		errChan <- conn.Wait()
	}()

	select {
	case <-s.stopChan:
		// Normal stop, close connection
		_ = conn.Close()
	case err := <-errChan:
		// Subprocess exited!
		s.handleDisconnect(err)
	}
}
```
This guarantees immediate, zero-CPU overhead detection of process dropouts.
