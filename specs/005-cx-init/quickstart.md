# Quickstart: Validating cx init

This guide details the validation steps to verify the `cx init` command.

---

## 1. Prerequisites

Before running the tests, compile the latest CLI binary:
```bash
go build -o cx .
```

---

## 2. Test Scenarios

### Scenario A: Initialization in a clean environment
1. **Clean up any existing configuration**:
   ```bash
   rm -f ~/.config/cx/config.yaml
   ```
2. **Execute initialization**:
   ```bash
   ./cx init
   ```
3. **Verify outcome**:
   * CLI outputs: `Success! Created starter configuration at /Users/shubham/.config/cx/config.yaml`
   * Check that the file was created and contains the starter comments and database examples.
   * Run a workspace check: `./cx use dev` (should succeed with `Now using workspace "dev"`).

---

### Scenario B: Accidental overwrite protection
1. **Ensure a configuration file exists**:
   ```bash
   ls ~/.config/cx/config.yaml
   ```
2. **Execute initialization without override**:
   ```bash
   ./cx init
   ```
3. **Verify outcome**:
   * CLI exits with code 1.
   * CLI outputs: `Error: configuration file already exists at /Users/shubham/.config/cx/config.yaml. Use --force to overwrite.`
   * Verify the file was not modified or corrupted.

---

### Scenario C: Force override
1. **Ensure a configuration file exists**:
   ```bash
   ls ~/.config/cx/config.yaml
   ```
2. **Execute initialization with force flag**:
   ```bash
   ./cx init --force
   ```
3. **Verify outcome**:
   * CLI exits with code 0.
   * CLI outputs: `Success! Overwrote starter configuration at /Users/shubham/.config/cx/config.yaml`
   * Check that the file has been reset back to default starter comments.
