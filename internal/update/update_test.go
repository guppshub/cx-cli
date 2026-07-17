package update

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input   string
		wantMaj int
		wantMin int
		wantPat int
		wantErr bool
	}{
		{"v1.2.3", 1, 2, 3, false},
		{"1.2.3", 1, 2, 3, false},
		{"v0.1.16", 0, 1, 16, false},
		{"v0.1.16-rc1", 0, 1, 16, false},
		{"12.34.56-beta.2", 12, 34, 56, false},
		{"invalid", 0, 0, 0, true},
		{"1.2", 0, 0, 0, true},
		{"1.2.3.4", 0, 0, 0, true},
		{"a.b.c", 0, 0, 0, true},
	}

	for _, tt := range tests {
		maj, min, pat, err := ParseVersion(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseVersion(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr {
			if maj != tt.wantMaj || min != tt.wantMin || pat != tt.wantPat {
				t.Errorf("ParseVersion(%q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.input, maj, min, pat, tt.wantMaj, tt.wantMin, tt.wantPat)
			}
		}
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"v0.1.15", "v0.1.16", true},
		{"v0.1.16", "v0.1.16", false},
		{"v0.2.0", "v0.1.16", false},
		{"v1.0.0", "v2.0.0", true},
		{"v1.1.0", "v1.2.0", true},
		{"v1.1.1", "v1.1.2", true},
		{"v0.0.0-dev", "v0.1.16", true},
		{"unknown", "v0.1.16", true},
		{"", "v0.1.16", true},
		{"v0.1.16", "invalid", false},
		{"invalid", "v0.1.16", true},
	}

	for _, tt := range tests {
		got := IsNewerVersion(tt.current, tt.latest)
		if got != tt.want {
			t.Errorf("IsNewerVersion(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestFetchLatestRelease(t *testing.T) {
	// Mock server returning a successful release payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"tag_name": "v9.9.9",
			"prerelease": false,
			"draft": false,
			"assets": [
				{
					"name": "cx-linux-amd64",
					"browser_download_url": "https://example.com/cx-linux-amd64"
				}
			]
		}`))
	}))
	defer server.Close()

	release, err := FetchLatestRelease(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error fetching release: %v", err)
	}

	if release.TagName != "v9.9.9" {
		t.Errorf("expected tag_name v9.9.9, got %q", release.TagName)
	}

	if len(release.Assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(release.Assets))
	}

	if release.Assets[0].Name != "cx-linux-amd64" {
		t.Errorf("expected asset name cx-linux-amd64, got %q", release.Assets[0].Name)
	}
}

func TestFindAsset(t *testing.T) {
	expectedName := fmt.Sprintf("cx-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		expectedName += ".exe"
	}

	release := &GitHubRelease{
		Assets: []ReleaseAsset{
			{Name: "cx-invalid-arch", DownloadURL: "https://example.com/invalid"},
			{Name: expectedName, DownloadURL: "https://example.com/match"},
		},
	}

	url, name, err := FindAsset(release)
	if err != nil {
		t.Fatalf("unexpected error finding asset: %v", err)
	}

	if name != expectedName {
		t.Errorf("expected asset name %q, got %q", expectedName, name)
	}

	if url != "https://example.com/match" {
		t.Errorf("expected download URL https://example.com/match, got %q", url)
	}
}

func TestSelfUpdate(t *testing.T) {
	// Create mock server returning mock binary content
	mockContent := "new-compiled-binary-data"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockContent))
	}))
	defer server.Close()

	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "cx-update-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create dummy "active binary"
	execPath := filepath.Join(tempDir, "cx")
	if runtime.GOOS == "windows" {
		execPath += ".exe"
	}

	err = os.WriteFile(execPath, []byte("old-binary-data"), 0755)
	if err != nil {
		t.Fatalf("failed to write dummy active binary: %v", err)
	}

	// Run selfUpdateWithPath
	err = selfUpdateWithPath(context.Background(), server.URL, execPath)
	if err != nil {
		t.Fatalf("selfUpdateWithPath failed: %v", err)
	}

	// Verify old binary was replaced
	content, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("failed to read updated binary: %v", err)
	}

	if string(content) != mockContent {
		t.Errorf("expected content %q, got %q", mockContent, string(content))
	}

	// Verify .old file was cleaned up or exists
	oldPath := execPath + ".old"
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Logf(".old file remained, this is acceptable if OS locks were simulated, but check if we can remove it manually")
	}
}
