package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const LatestReleaseURL = "https://api.github.com/repos/guppshub/cx-cli/releases/latest"

// GitHubRelease represents a release structure from the GitHub API.
type GitHubRelease struct {
	TagName    string         `json:"tag_name"`
	Prerelease bool           `json:"prerelease"`
	Draft      bool           `json:"draft"`
	Assets     []ReleaseAsset `json:"assets"`
}

// ReleaseAsset represents an asset uploaded to a GitHub release.
type ReleaseAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

// FetchLatestRelease queries the GitHub API for the latest release of cx-cli.
func FetchLatestRelease(ctx context.Context, apiURL string) (*GitHubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// GitHub API guidelines require setting a User-Agent
	req.Header.Set("User-Agent", "cx-cli-updater")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %s", resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return &release, nil
}

// ParseVersion parses a semver string (e.g. "v1.2.3" or "1.2.3-rc1") into major, minor, patch components.
func ParseVersion(v string) (int, int, int, error) {
	if len(v) > 0 && (v[0] == 'v' || v[0] == 'V') {
		v = v[1:]
	}
	// Strip pre-release suffix (e.g. -rc1)
	if idx := strings.Index(v, "-"); idx >= 0 {
		v = v[:idx]
	}

	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid semver format: %s", v)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %w", err)
	}

	return major, minor, patch, nil
}

// IsNewerVersion returns true if the latest version is strictly greater than the current version.
func IsNewerVersion(current, latest string) bool {
	// Dev/unknown versions are always considered older than any valid release.
	if current == "v0.0.0-dev" || current == "unknown" || current == "" {
		_, _, _, err := ParseVersion(latest)
		return err == nil
	}

	cMaj, cMin, cPat, err := ParseVersion(current)
	if err != nil {
		return true // Default to update if current is unparseable
	}

	lMaj, lMin, lPat, err := ParseVersion(latest)
	if err != nil {
		return false // Cannot update to invalid release version
	}

	if lMaj != cMaj {
		return lMaj > cMaj
	}
	if lMin != cMin {
		return lMin > cMin
	}
	return lPat > cPat
}

// FindAsset returns the download URL and name for the asset matching the current OS and architecture.
func FindAsset(release *GitHubRelease) (string, string, error) {
	expectedName := fmt.Sprintf("cx-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		expectedName += ".exe"
	}

	for _, asset := range release.Assets {
		if strings.EqualFold(asset.Name, expectedName) {
			return asset.DownloadURL, asset.Name, nil
		}
	}

	return "", "", fmt.Errorf("no release asset found for platform %s-%s (expected name: %s)", runtime.GOOS, runtime.GOARCH, expectedName)
}

// SelfUpdate downloads the new binary from the specified URL and replaces the currently running executable.
func SelfUpdate(ctx context.Context, downloadURL string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve active executable path: %w", err)
	}
	return selfUpdateWithPath(ctx, downloadURL, execPath)
}

func selfUpdateWithPath(ctx context.Context, downloadURL, execPath string) error {
	// 2. Resolve target directory
	execDir := filepath.Dir(execPath)

	// 3. Create temp file in same directory to avoid partition boundaries issues
	tmpFile, err := os.CreateTemp(execDir, "cx_new_*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath) // Cleanup if we exit early
	}()

	// 4. Download asset
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("User-Agent", "cx-cli-updater")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download release binary: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download server returned status %s", resp.Status)
	}

	// 5. Write to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write downloaded binary: %w", err)
	}
	_ = tmpFile.Close()

	// 6. Set execute permissions (especially on Unix/macOS)
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to set execute permissions on temp binary: %w", err)
	}

	// 7. Perform the rename-dance
	oldPath := execPath + ".old"
	// Silently remove any existing stale .old file from previous update attempts
	_ = os.Remove(oldPath)

	// Rename current binary to .old
	if err := os.Rename(execPath, oldPath); err != nil {
		return fmt.Errorf("failed to rename current binary to backup: %w", err)
	}

	// Rename temp file to original binary name
	if err := os.Rename(tmpPath, execPath); err != nil {
		// Rollback if possible
		_ = os.Rename(oldPath, execPath)
		return fmt.Errorf("failed to replace active binary: %w", err)
	}

	// 8. Try to clean up the .old file (swallowed if locked/fails, will clean up next time)
	_ = os.Remove(oldPath)

	return nil
}
