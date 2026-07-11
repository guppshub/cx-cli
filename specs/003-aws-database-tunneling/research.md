# Research: AWS Database Tunneling

## 1. Subprocess Pipe to net.Conn Adapter

To avoid rewriting the complex AWS Session Manager WebSocket protocol (normally handled by the `session-manager-plugin`), we will wrap the plugin in a standard Go `net.Conn` struct.

### Design Pattern: Stream Redirection
- **Launch Command**: `aws ssm start-session --target <instance-id> --document-name AWS-StartPortForwardingSessionToRemoteHost --parameters host=<db-host>,portNumber=<db-port>,localPortNumber=<ignored>`
- **Stream Binding**: We run this process using `os/exec.CommandContext`. We redirect stdin using `cmd.StdinPipe()` and stdout using `cmd.StdoutPipe()`.
- **Custom Wrapper Struct**:
  ```go
  type ProcessConn struct {
      cmd    *exec.Cmd
      stdin  io.WriteCloser
      stdout io.ReadCloser
  }
  func (c *ProcessConn) Read(b []byte) (int, error)  { return c.stdout.Read(b) }
  func (c *ProcessConn) Write(b []byte) (int, error) { return c.stdin.Write(b) }
  func (c *ProcessConn) Close() error {
      c.stdin.Close()
      c.stdout.Close()
      return c.cmd.Process.Kill()
  }
  ```

---

## 2. Port Forwarding Core loop

The CLI core orchestrates local TCP listener binding and session proxying:
1. **Local Socket Listen**: Bind `net.Listen("tcp", "localhost:<port>")`. If the port is in use, fallback to random binding (`localhost:0`) and report the allocated port.
2. **Accept Connection**: For every accepted client connection `clientConn`:
   - Start a goroutine.
   - Invoke `DialTunnel()` on the active dialer.
   - Forward data using two concurrent `io.Copy(clientConn, tunnelConn)` and `io.Copy(tunnelConn, clientConn)` routines.
3. **Connection Lifecycle**: If the client disconnects or the subprocess closes, clean up the resources and active state records.

---

## 3. Headless Authentication Prompt Callback

We researched the best way to handle MFA and browser authorization without polluting the provider library with UI logic.
- **Standard Go pattern**: Implement a prompt callback interface or type.
  ```go
  type Prompter func(prompt string, secret bool) (string, error)
  ```
- **Provider interaction**: The provider is passed the `Prompter` function during `EnsureCredentials()`. It calls this callback if it needs to prompt for an MFA code or output a login instruction link.
