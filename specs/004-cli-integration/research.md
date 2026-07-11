# Research: CLI Integration

## 1. Dynamic YAML Resource Parsing

To preserve the provider-agnostic boundary of the `config` package, workspaces store provider details and resources in an inline `map[string]any` called `Raw`. 

To extract the databases from this raw map:
- **Approach**: Leverage `gopkg.in/yaml.v3` (already in our dependencies).
- **Logic**: Marshal `Workspace.Raw` to a YAML byte array, and then unmarshal it into a strictly typed local structure containing our database resources.
  ```go
  bytes, err := yaml.Marshal(workspace.Raw)
  if err != nil {
      return nil, err
  }
  var result struct {
      Resources struct {
          Databases []DatabaseResource `yaml:"databases"`
      } `yaml:"resources"`
  }
  err = yaml.Unmarshal(bytes, &result)
  ```
- **Benefit**: Zero reflection boilerplate, zero new third-party dependencies, and maximum compatibility with standard configuration schemas.

---

## 2. Graceful Shutdown & Signal Trapping

To ensure that listener sockets are closed and connections are cleared from `state.json` when the user terminates `cx db` (e.g. via `Ctrl+C`):
- **Approach**: Use Go's standard library `os/signal` and `context` package.
- **Logic**: Create a signal-notifying context using `signal.NotifyContext`:
  ```go
  ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
  defer cancel()
  ```
- **Forwarding Execution**: We pass this context directly to `db.Controller.Start()`. When the interrupt fires, the context is cancelled, which triggers the accepted listener closure and state deregistration defer block cleanly.
