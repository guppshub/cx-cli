# Research: Workspace Management

## 1. Workspace Persistence Map

We researched the best way to persist workspaces inside the configuration file (`config.yaml`) without violating the provider-agnostic constraints.

### Context to Workspace Mapping
A workspace is represented by a `Context` in our configuration schema:
- The workspace **Name** maps to the key in the `contexts` map of `config.yaml`.
- The workspace **Provider** maps to the `provider` field.
- The workspace **Provider Configuration** maps to the `raw` map inline.

### Active Workspace Representation
To track the currently active workspace in `config.yaml`, we will add a `Current` field to the `Config` struct:
```go
type Config struct {
	Version     string             `yaml:"version"`
	Current     string             `yaml:"current"`
	Contexts    map[string]*Context `yaml:"contexts"`
	Preferences map[string]string  `yaml:"preferences"`
}
```
This aligns with the functional requirement:
```yaml
current: staging
```

---

## 2. Validation & Ordering

### Duplicate Checks
Workspace names must be validated case-sensitively. When adding a workspace, we check `cfg.Contexts[name]` to verify if it already exists, returning a wrapped error if it does.

### Deletion and Rename Protection
- If the workspace to delete matches `cfg.Current`, the operation is rejected.
- If the active workspace matches the renamed workspace, `cfg.Current` is updated to the new name before saving.

### Deterministic Listing
Go maps do not guarantee iteration order. To return workspaces in a deterministic alphabetical order:
1. Extract all workspace names from `cfg.Contexts`.
2. Sort the names using `sort.Strings()`.
3. Build the output list by iterating over the sorted keys.
