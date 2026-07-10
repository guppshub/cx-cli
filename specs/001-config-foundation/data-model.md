# Data Model: Configuration Foundation

This document defines the structs and data models for configuration and runtime state representation.

## 1. Configuration Data Model (`config.yaml`)

```go
type Config struct {
	Version     string               `yaml:"version"`
	Contexts    map[string]*Context  `yaml:"contexts"`
	Preferences map[string]string    `yaml:"preferences"`
}

type Context struct {
	Provider  string                 `yaml:"provider"`
	AWS       *AWSConfig             `yaml:"aws,omitempty"`
	Resources *Resources             `yaml:"resources,omitempty"`
}

type AWSConfig struct {
	Profile string `yaml:"profile"`
	Region  string `yaml:"region"`
}

type Resources struct {
	Databases []DatabaseResource `yaml:"databases,omitempty"`
	Caches    []CacheResource    `yaml:"caches,omitempty"`
}

type DatabaseResource struct {
	Name              string `yaml:"name"`
	Engine            string `yaml:"engine"`
	Endpoint          string `yaml:"endpoint"`
	Port              int    `yaml:"port"`
	LocalPort         int    `yaml:"local_port"`
	BastionInstanceID string `yaml:"bastion_instance_id"`
}

type CacheResource struct {
	Name              string `yaml:"name"`
	Endpoint          string `yaml:"endpoint"`
	Port              int    `yaml:"port"`
	LocalPort         int    `yaml:"local_port"`
	BastionInstanceID string `yaml:"bastion_instance_id"`
}
```

### Validation Rules
- `Version` MUST be equal to `"1"` or `"1.0"`.
- `Contexts` map keys MUST NOT contain duplicate context names (enforced by YAML parser natively, but programmatically verified).
- Context names MUST NOT contain spaces or special characters except hyphens and underscores.

---

## 2. Runtime State Data Model (`state.json`)

```go
type State struct {
	CurrentContext    string                         `json:"current_context"`
	ActiveConnections map[string]*ConnectionMetadata `json:"active_connections"`
}

type ConnectionMetadata struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	LocalPort    int    `json:"local_port"`
	ConnectionID string `json:"connection_id"`
	ConnectedAt  string `json:"connected_at"` // RFC3339 format
}
```
