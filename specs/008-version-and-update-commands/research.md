# Research: Version & Update Commands

This document contains research, design decisions, and alternatives considered for implementing the `version` and `update` commands in `cx`.

## 1. Build-Time Version Injection in Go

### The Decision:
Use Go's compiler linker flags (`-ldflags`) to inject version metadata at compile time during build.
Specifically, we will define string variables in our `cmd` package:
```go
package cmd

var (
	Version   = "v0.0.0-dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
)
```
During compilation (e.g. in GitHub Actions release workflows), we compile using:
```bash
go build -ldflags "-X github.com/guppshub/cx-cli/cmd.Version=${TAG} -X github.com/guppshub/cx-cli/cmd.CommitSHA=${COMMIT_SHA} -X github.com/guppshub/cx-cli/cmd.BuildTime=${TIMESTAMP}"
```

### Rationale:
* **Clean Code**: Keeps version logic simple without requiring separate text files or run-time filesystem checks.
* **Dev Friendly**: Defaults to `v0.0.0-dev` when running locally via `go run` or standard `go build`, letting developers know they are not on an official release.

---

## 2. GitHub Releases API & Update Checks

### The Decision:
Use Go's standard library `net/http` to send a GET request to:
`https://api.github.com/repos/guppshub/cx-cli/releases/latest`

We will define a custom `User-Agent` header (e.g. `cx-cli-updater`) as required by GitHub API guidelines, and set a strict timeout of `5` seconds to prevent the CLI from hanging when offline.

### Rationale:
* **Zero Dependencies**: Standard library `encoding/json` and `net/http` are highly optimized and stable. Third-party GitHub API SDKs would bloat our binary size and introduce unnecessary update dependencies.
* **Simplicity**: The JSON structure returned by the endpoint is simple and easily parsed using a minimal Go struct:
```go
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}
```

---

## 3. Windows Executable Self-Replacement (Rename-Dance)

### The Decision:
On Windows, a running executable file is locked by the OS kernel, preventing other processes from overwriting or deleting it. However, Windows **does** allow renaming the active running executable within the same directory.
We will implement the following platform-safe self-replacement logic:

1. **Download**: Download the new binary asset to the same directory as the current binary with a `.tmp` extension (e.g., `cx.exe.tmp`).
2. **Rename Old**: Rename the currently running binary (resolved via `os.Executable()`) to a `.old` extension (e.g., `cx.exe.old`).
3. **Rename New**: Rename the `.tmp` file to the original binary name (e.g., `cx.exe.tmp` -> `cx.exe`).
4. **Cleanup**: Try to delete the `.old` file. If the OS file lock is still present, we swallow the error silently and let the next execution of `cx` clean up any `.old` files in its directory context.

On Unix-like systems (Linux, macOS), we can overwrite the binary directly, but we must make sure to preserve/restore execute permissions (`0755`) on the replaced file.

### Rationale:
* **EDR-Friendly**: Does not spawn external processes to execute scripting loops, keeping EDR systems happy.
* **Robustness**: If step 1 fails, the active binary is untouched. If step 2 fails, the active binary is untouched. This prevents half-downloaded or corrupted upgrades from breaking the user's installation.
