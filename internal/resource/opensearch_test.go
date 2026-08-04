package resource

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guppshub/cx-cli/internal/config"
)

func TestResolveOpenSearch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cx_test_os_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configPath := filepath.Join(tmpDir, "config.yaml")
	_ = os.Setenv("CX_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("CX_CONFIG") }()

	yamlContent := `
version: "1"
current: dev
workspaces:
  dev:
    provider: aws
    profile: dev-profile
    region: us-east-1
    bastion_instance_id: i-0d1d909c5fea48c31
    mfa_serial: arn:aws:iam::123456789012:mfa/ws-mfa
    mfa_duration: 3600
    resources:
      opensearch:
        - name: logs-dev
          endpoint: vpc-my-dev-domain.us-east-1.es.amazonaws.com
        - name: secure-os
          endpoint: secure.us-east-1.es.amazonaws.com
          port: 8443
          local_port: 9201
          mfa: true
          mfa_serial: arn:aws:iam::123456789012:mfa/override-mfa
          mfa_duration: 7200
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write mock config: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load mock config: %v", err)
	}

	ws := cfg.Workspaces["dev"]

	// 1. Test logs-dev with defaults and workspace fallbacks
	osDev, err := ResolveOpenSearch(ws, "logs-dev")
	if err != nil {
		t.Fatalf("failed to resolve logs-dev: %v", err)
	}

	if osDev.Name != "logs-dev" || osDev.Endpoint != "vpc-my-dev-domain.us-east-1.es.amazonaws.com" {
		t.Errorf("unexpected logs-dev resource resolved: %+v", osDev)
	}

	if osDev.Port != 443 || osDev.LocalPort != 9200 {
		t.Errorf("expected default ports 443/9200, got %d/%d", osDev.Port, osDev.LocalPort)
	}

	if osDev.MFASerial != "arn:aws:iam::123456789012:mfa/ws-mfa" || osDev.MFADuration != 3600 {
		t.Errorf("expected workspace-level fallback MFA parameters, got %s/%d", osDev.MFASerial, osDev.MFADuration)
	}

	// 2. Test secure-os override
	osSec, err := ResolveOpenSearch(ws, "secure-os")
	if err != nil {
		t.Fatalf("failed to resolve secure-os: %v", err)
	}

	if osSec.Port != 8443 || osSec.LocalPort != 9201 {
		t.Errorf("expected overridden ports, got %d/%d", osSec.Port, osSec.LocalPort)
	}

	if osSec.MFASerial != "arn:aws:iam::123456789012:mfa/override-mfa" || osSec.MFADuration != 7200 {
		t.Errorf("expected resource-level overridden MFA parameters, got %s/%d", osSec.MFASerial, osSec.MFADuration)
	}

	// 3. Test FetchOpenSearch
	list, err := FetchOpenSearch(ws)
	if err != nil {
		t.Fatalf("failed to fetch OpenSearch list: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("expected 2 OpenSearch resources, got %d", len(list))
	}
}
