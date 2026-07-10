# Quickstart: Workspace Management Validation Guide

This document describes how to validate the Workspace Management package.

## Prerequisites
- Go installed (`1.22.5` or later)
- Access to terminal

## Validation Scenarios

### Scenario 1: Run Unit Tests
To run all unit tests verifying adding, deleting, renaming, selecting, listing, and active workspace protection:

```bash
# Execute tests in the workspace package
go test -v -race ./internal/workspace/...
```

**Expected Outcome:**
- All tests pass successfully (no failures).
- Output reports test coverage.

### Scenario 2: Validation of Workspace Add & List
To verify that workspaces are added and listed alphabetically:
1. Initialize a clean configuration using the verify tool:
   ```bash
   go run specs/002-workspace-management/scratch/verify.go --action=init
   ```
2. Add workspaces `staging` and `production`:
   ```bash
   go run specs/002-workspace-management/scratch/verify.go --action=add --name=staging --provider=aws
   go run specs/002-workspace-management/scratch/verify.go --action=add --name=production --provider=aws
   ```
3. List all workspaces:
   ```bash
   go run specs/002-workspace-management/scratch/verify.go --action=list
   ```
4. **Expected Output:**
   - List shows `production` and `staging` sorted alphabetically.
   - Status indicates neither is active yet (unless selected).

### Scenario 3: Validation of Active Workspace Protection
To verify that the active workspace cannot be deleted:
1. Select `staging` as the active workspace:
   ```bash
   go run specs/002-workspace-management/scratch/verify.go --action=use --name=staging
   ```
2. Attempt to delete `staging`:
   ```bash
   go run specs/002-workspace-management/scratch/verify.go --action=delete --name=staging
   ```
3. **Expected Output:**
   - Command fails with: `cannot delete active workspace "staging"`.
