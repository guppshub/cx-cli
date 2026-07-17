# Quickstart & Validation: Version & Update Commands

This guide provides step-by-step instructions to test and validate the `version` and `update` commands.

## 1. Validating Version Command

### Local Development Version (Default)
Run:
```bash
go run . version
```
**Expected Outcome**:
```text
cx version v0.0.0-dev (commit: unknown, built: unknown)
```

### Injected Production Version
Compile the binary manually using `-ldflags` injection:
```bash
go build -ldflags "-X github.com/guppshub/cx-cli/cmd.Version=v9.9.9 -X github.com/guppshub/cx-cli/cmd.CommitSHA=abc1234 -X github.com/guppshub/cx-cli/cmd.BuildTime=2026-07-14T09:00:00Z" -o cx_test .
```
Run the compiled test binary:
```bash
./cx_test version
```
**Expected Outcome**:
```text
cx version v9.9.9 (commit: abc1234, built: 2026-07-14T09:00:00Z)
```

---

## 2. Validating Update Check

Check for updates without downloading or installing:
```bash
./cx_test update --check
```
**Expected Outcome**:
Since the current version is `v9.9.9` (which is newer than any release on GitHub), the CLI should report:
```text
You are already running the latest version of cx (v9.9.9).
```

### Mocking Older Version
Compile a version that is guaranteed to be older than the latest release:
```bash
go build -ldflags "-X github.com/guppshub/cx-cli/cmd.Version=v0.0.1" -o cx_test .
./cx_test update --check
```
**Expected Outcome**:
The CLI should contact GitHub, identify the latest release, and output:
```text
A newer version of cx is available: v0.1.16 (current: v0.0.1).
```

---

## 3. Validating Self-Upgrade

### Test Interactive Prompt
Run `update` without `--yes`:
```bash
./cx_test update
```
**Expected Outcome**:
```text
A newer version of cx is available: v0.1.16 (current: v0.0.1).
Would you like to upgrade? (y/N): 
```
* If you type `n` and press enter, it exits with `Update canceled.`
* If you type `y` and press enter, it downloads, updates the executable, and prints:
  `Successfully updated cx to v0.1.16!`

### Test Non-Interactive Auto-Upgrade
Run `update` with `--yes`:
```bash
./cx_test update --yes
```
**Expected Outcome**:
Instantly downloads the matching binary, performs the rename replacement, and exits with:
```text
Downloading latest release...
Successfully updated cx to v0.1.16!
```
Verify the active version is now updated:
```bash
./cx_test version
```
*(Should now report version `v0.1.16`).*
