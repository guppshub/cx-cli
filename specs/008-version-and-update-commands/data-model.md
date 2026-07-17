# Data Model: Version & Update Commands

This document describes the in-memory data structures and JSON schemas used to perform update checks and execute the self-replacement sequence.

## 1. GitHub Release Schema

We query `https://api.github.com/repos/guppshub/cx-cli/releases/latest` to fetch the latest official release version. The JSON response is mapped to the following Go struct:

```go
type GitHubRelease struct {
	TagName    string        `json:"tag_name"`    // E.g., "v0.1.16"
	Prerelease bool          `json:"prerelease"`  // True if this is a pre-release
	Draft      bool          `json:"draft"`       // True if this is a draft release
	Assets     []ReleaseAsset `json:"assets"`      // List of uploaded binary assets
}

type ReleaseAsset struct {
	Name        string `json:"name"`                 // E.g., "cx-windows-amd64.exe"
	DownloadURL string `json:"browser_download_url"` // URL to download the binary asset
}
```

## 2. Release Asset Naming Convention

To download the correct asset matching the user's OS and CPU architecture, we match the asset `Name` against the current operating system (`runtime.GOOS`) and architecture (`runtime.GOARCH`).

The asset filename format on GitHub releases MUST be:
`cx-<os>-<arch>[extension]`

### Supported Mappings:
| Operating System (`GOOS`) | Architecture (`GOARCH`) | Target Asset File Pattern |
|---------------------------|-------------------------|---------------------------|
| `darwin` (macOS)          | `amd64` (Intel)         | `cx-darwin-amd64`         |
| `darwin` (macOS)          | `arm64` (Apple Silicon) | `cx-darwin-arm64`         |
| `linux`                   | `amd64`                 | `cx-linux-amd64`          |
| `linux`                   | `arm64`                 | `cx-linux-arm64`          |
| `windows`                 | `amd64`                 | `cx-windows-amd64.exe`    |
| `windows`                 | `arm64`                 | `cx-windows-arm64.exe`    |

---

## 3. Version Representation

In memory, versions are compared using semantic version strings. We extract version tags (e.g. `v0.1.16` vs `v0.1.15`) and compare them to determine if an update is available.

* Standard semantic versioning (`semver`) comparisons are used.
* Any version prefix `v` is trimmed before comparison.
* Development builds (`v0.0.0-dev`) are always considered older than any official release, enabling developers to test updates locally.
