# Data Model: Connection Metadata

This document describes the schema updates for `state.json` to support rich supervisor state tracking.

---

## 1. Updated Connection Metadata Schema

We will add fields to the existing `ConnectionMetadata` struct in `internal/state/state.go` to support state tracking, restart counts, and failure tracking.

### Struct Definition:
```go
package state

// ConnectionMetadata represents metadata for active background connections.
type ConnectionMetadata struct {
	Type         string `json:"type"`          // "database" or "redis"
	Name         string `json:"name"`          // Resource name (e.g. sequr)
	LocalPort    int    `json:"local_port"`    // Bound local port
	ConnectionID string `json:"connection_id"` // Unique identifier (e.g. cx-conn-sequr-5432)
	ConnectedAt  string `json:"connected_at"`  // First establishment timestamp
	Pid          int    `json:"pid"`           // OS Process ID of the daemon process

	// Supervisor-added fields
	State        string `json:"state"`         // "Healthy", "Restarting", "Failed", etc.
	Restarts     int    `json:"restarts"`      // Number of restarts attempted
	LastFailure  string `json:"last_failure"`  // Error message of the last failure
	LastRestart  string `json:"last_restart"`  // Timestamp of the last restart
}
```

---

## 2. Backward Compatibility

The Go `json` standard decoder will gracefully handle older `state.json` files that lack the `state`, `restarts`, `last_failure`, or `last_restart` fields by leaving those fields as their zero-values (empty string or `0`). 

When `cx status` reads these zero-values:
* If `State == ""`, it will default to displaying `Healthy` (assuming older tunnels are healthy since they are running).
* If `Restarts` is missing, it displays `0`.
* This keeps the migration transparent and risk-free.
