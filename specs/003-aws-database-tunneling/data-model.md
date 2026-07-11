# Data Model: AWS Database Tunneling

This document defines the structs required for establishing database connection tunnels.

## 1. Domain Entities

```go
package workflow

// TunnelTarget defines the destination targets for the tunnel dialer.
type TunnelTarget struct {
	BastionInstanceID string `json:"bastion_instance_id"`
	RemoteHost        string `json:"remote_host"`
	RemotePort        int    `json:"remote_port"`
}

// Endpoint represents a resolved local or remote network socket destination.
type Endpoint struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}
```

## 2. State Mapping
Once a database tunnel is successfully established, the active tunnel properties are persisted inside `state.json` via the state package's `ConnectionMetadata` structure:
```json
{
  "type": "database",
  "name": "mercury",
  "local_port": 5432,
  "connection_id": "cx-conn-staging-db-mercury",
  "connected_at": "2026-07-11T02:00:00Z"
}
```
The state manager updates the machine-managed connection list accordingly.
