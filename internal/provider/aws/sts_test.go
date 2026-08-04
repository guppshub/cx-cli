package aws

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadSaveSTSCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cx_test_sts_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configPath := filepath.Join(tmpDir, "config.yaml")
	_ = os.Setenv("CX_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("CX_CONFIG") }()

	// 1. Load empty cache
	cache, err := LoadSTSCache()
	if err != nil {
		t.Fatalf("failed to load empty cache: %v", err)
	}
	if len(cache.Credentials) != 0 {
		t.Errorf("expected empty credentials map, got %d", len(cache.Credentials))
	}

	// 2. Modify and Save cache
	expTime := time.Now().Add(1 * time.Hour).Truncate(time.Second)
	cache.Credentials["dev/default"] = STSCredentials{
		AccessKeyID:     "AKIA...",
		SecretAccessKey: "secret...",
		SessionToken:    "token...",
		Expiration:      expTime,
	}

	err = SaveSTSCache(cache)
	if err != nil {
		t.Fatalf("failed to save STS cache: %v", err)
	}

	// 3. Reload cache and assert
	loaded, err := LoadSTSCache()
	if err != nil {
		t.Fatalf("failed to load STS cache: %v", err)
	}

	creds, ok := loaded.Credentials["dev/default"]
	if !ok {
		t.Fatal("expected 'dev/default' credentials to exist in loaded cache")
	}

	if creds.AccessKeyID != "AKIA..." || creds.SecretAccessKey != "secret..." || creds.SessionToken != "token..." {
		t.Errorf("loaded credentials value mismatch: %+v", creds)
	}

	if !creds.Expiration.Equal(expTime) {
		t.Errorf("expected expiration %v, got %v", expTime, creds.Expiration)
	}
}

func TestSTS_JSONParsing(t *testing.T) {
	sampleIAMResponse := `{
		"mfaDevices": [
			{
				"userName": "shubham",
				"serialNumber": "arn:aws:iam::123456789012:mfa/shubham",
				"enableDate": "2026-07-11T08:10:20Z"
			}
		]
	}`

	var iamOutput struct {
		MFADevices []struct {
			SerialNumber string `json:"serialNumber"`
		} `json:"mfaDevices"`
	}

	if err := json.Unmarshal([]byte(sampleIAMResponse), &iamOutput); err != nil {
		t.Fatalf("failed to unmarshal mock IAM response: %v", err)
	}

	if len(iamOutput.MFADevices) != 1 || iamOutput.MFADevices[0].SerialNumber != "arn:aws:iam::123456789012:mfa/shubham" {
		t.Errorf("unexpected parsed IAM response: %+v", iamOutput)
	}

	sampleSTSResponse := `{
		"credentials": {
			"accessKeyId": "ASIA1234567890",
			"secretAccessKey": "SECRET_KEY_abc123",
			"sessionToken": "SESSION_TOKEN_xyz789",
			"expiration": "2026-08-04T22:30:00Z"
		}
	}`

	var stsOutput struct {
		Credentials struct {
			AccessKeyId     string    `json:"accessKeyId"`
			SecretAccessKey string    `json:"secretAccessKey"`
			SessionToken    string    `json:"sessionToken"`
			Expiration      time.Time `json:"expiration"`
		} `json:"credentials"`
	}

	if err := json.Unmarshal([]byte(sampleSTSResponse), &stsOutput); err != nil {
		t.Fatalf("failed to unmarshal mock STS response: %v", err)
	}

	if stsOutput.Credentials.AccessKeyId != "ASIA1234567890" || stsOutput.Credentials.Expiration.IsZero() {
		t.Errorf("unexpected parsed STS response: %+v", stsOutput)
	}
}
