# Quickstart: AWS Database Tunneling Validation Guide

This document describes how to validate the AWS Database Tunneling features.

## Prerequisites
- Go installed (`1.22.5` or later)
- Access to terminal

## Validation Scenarios

### Scenario 1: Run Unit Tests
To run all unit tests verifying the subprocess connection wrapper, port listener fallback, and state json tracking:

```bash
# Execute tests in both the provider and workflow packages
go test -v -race ./internal/provider/aws/...
go test -v -race ./internal/workflow/db/...
```

**Expected Outcome:**
- All tests pass successfully (no failures).
- Output reports test coverage.

### Scenario 2: Validation of Mock Subprocess Connection
To verify that the custom Go `net.Conn` wrapper correctly forwards and encapsulates process standard streams:
1. Run the verification script:
   ```bash
   go run specs/003-aws-database-tunneling/scratch/verify.go --action=subprocess-mock
   ```
2. **Expected Output:**
   - Script runs a dummy subprocess that echoes input (e.g. `cat` or custom Go echo binary).
   - Verifies bytes written to the connection struct are correctly read back from it.

### Scenario 3: Validation of Local Socket Binding Fallback
To verify that binding falls back if a port is in use:
1. Run two instances of the listener in separate terminal windows:
   ```bash
   # Terminal 1 - starts a listener on local port 5432
   go run specs/003-aws-database-tunneling/scratch/verify.go --action=listen --port=5432
   
   # Terminal 2 - attempts to listen on port 5432 again
   go run specs/003-aws-database-tunneling/scratch/verify.go --action=listen --port=5432
   ```
2. **Expected Output:**
   - Terminal 1 binds successfully to `5432`.
   - Terminal 2 detects the conflict and automatically binds to a random open port, reporting: `Port 5432 in use, bound to <random-port>`.
