# Quickstart: Validating Picker & EC2 SSH

This document describes the manual validation procedures to verify the TUI resource picker and `cx ec2` integration.

---

## 1. Compile and Ready CLI

First, build the latest CLI:
```bash
go build -o cx .
```

---

## 2. Test Scenarios

### Scenario A: Reusable Picker component validation
To verify the UI component in isolation, we will run a scratch command loaded with mock server entries:

1. **Run verification script**:
   ```bash
   go run specs/006-tui-picker-ec2-ssh/scratch/verify_picker.go
   ```
2. **Verify keyboard navigation**:
   * Use `↑` and `↓` arrow keys to move the selection cursor.
   * Verify the highlighted row is styled in bold/underlined.
3. **Verify search input**:
   * Press `/` to enter search mode (or verify search is active by default).
   * Type `redis`. Verify the list instantly filters down to items containing `redis` in any field.
   * Type backspace. Verify the list restores previous elements.
4. **Verify cancellation**:
   * Press `Esc`. The picker should close and print `Selection cancelled`.
5. **Verify confirmation**:
   * Re-run the script, select a mock item, and press `Enter`.
   * Verify the picker closes and prints the selected item ID (e.g., `i-09f87c4f1c901844a`).

---

### Scenario B: Active `cx ec2` session execution
1. **Choose your workspace**:
   ```bash
   ./cx use dev
   ```
2. **Execute EC2 command**:
   ```bash
   ./cx ec2
   ```
3. **Verify picker rendering**:
   * Check that all EC2 instances from the workspace are retrieved and displayed in columns (Name, ID, State, IP).
4. **Execute connection**:
   * Select a running instance and press `Enter`.
   * Verify that the terminal transitions immediately and cleanly into the interactive SSM shell session.
   * Run commands inside the shell (e.g. `whoami`, `ls`).
   * Type `exit` and verify you exit back to your local shell, and `cx` exits cleanly with code 0.
