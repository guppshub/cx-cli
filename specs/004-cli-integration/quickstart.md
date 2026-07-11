# Quickstart: CLI Integration Validation Guide

This document describes how to validate the CLI commands `cx use` and `cx db`.

## Prerequisites
- Go installed (`1.22.5` or later)
- Access to terminal

## Validation Scenarios

### Scenario 1: Setup Mock Configuration
1. Create a `config.yaml` at your OS user config location (`~/.config/cx/config.yaml` or macOS equivalent):
   ```yaml
   version: "1"
   current: ""
   workspaces:
     staging:
       provider: aws
       region: us-east-1
       profile: staging-admin
       resources:
         databases:
           - name: mercury
             engine: postgres
             endpoint: staging-db.xxxx.us-east-1.rds.amazonaws.com
             port: 5432
             local_port: 5432
             bastion_instance_id: i-1234567890
   ```

### Scenario 2: Validate Workspace Selector
1. Run the `use` command to switch active workspace:
   ```bash
   go run main.go use staging
   ```
2. **Expected Output:**
   - Logs: `Now using workspace "staging"`.
   - `config.yaml` file now contains: `current: staging`.

### Scenario 3: Validate DB Command Execution
1. Run the `db` command:
   ```bash
   go run main.go db mercury
   ```
2. **Expected Output:**
   - Command detects that `aws` and `session-manager-plugin` are missing or present. (If present, starts connection).
   - If mock AWS is simulated or real credentials present, establishes tunnel:
     `Tunneling mercury (postgres) through local port 5432...`
3. Press `Ctrl+C` to terminate the connection.
   - **Expected Output:**
     - Logs: `Terminating tunnel connection...`
     - Command exits cleanly with code 0.
