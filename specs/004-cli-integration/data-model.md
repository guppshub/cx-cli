# Data Model: CLI Integration

This document defines the schema parsed from raw workspaces to resolve database resource targets.

## 1. Domain Entities

```go
package resource

// DatabaseResource represents the configuration for a target database resource.
type DatabaseResource struct {
	Name              string `yaml:"name"`
	Engine            string `yaml:"engine"`
	Endpoint          string `yaml:"endpoint"`
	Port              int    `yaml:"port"`
	LocalPort         int    `yaml:"local_port"`
	BastionInstanceID string `yaml:"bastion_instance_id"`
}
```

## 2. Parsing Target Mappings
The database configuration resides in the active workspace's configuration block:
- **Location**: `workspaces.<active_workspace_name>.resources.databases`
- **Schema Mapping**: The fields map directly from YAML to the `DatabaseResource` struct.
