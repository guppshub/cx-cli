package aws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/guppshub/cx-cli/internal/config"
)

// STSCredentials represents the temporary session credentials returned by STS.
type STSCredentials struct {
	AccessKeyID     string    `json:"access_key_id"`
	SecretAccessKey string    `json:"secret_access_key"`
	SessionToken    string    `json:"session_token"`
	Expiration      time.Time `json:"expiration"`
}

// STSCache represents the structure of the local secure credentials cache.
type STSCache struct {
	Credentials map[string]STSCredentials `json:"credentials"` // key is "workspaceName/profile"
}

// STSPath resolves the path to sts_cache.json.
func STSPath() (string, error) {
	cPath, err := config.Path()
	if err != nil {
		return "", fmt.Errorf("resolving config path for STS cache: %w", err)
	}
	return filepath.Join(filepath.Dir(cPath), "sts_cache.json"), nil
}

// LoadSTSCache loads the temporary credentials cache.
func LoadSTSCache() (*STSCache, error) {
	path, err := STSPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &STSCache{Credentials: make(map[string]STSCredentials)}, nil
		}
		return nil, fmt.Errorf("reading STS cache file: %w", err)
	}

	var cache STSCache
	if err := json.Unmarshal(data, &cache); err != nil {
		// Return clean empty cache if corrupt
		return &STSCache{Credentials: make(map[string]STSCredentials)}, nil
	}

	if cache.Credentials == nil {
		cache.Credentials = make(map[string]STSCredentials)
	}
	return &cache, nil
}

// SaveSTSCache writes the credentials cache back to disk securely (0600 permissions).
func SaveSTSCache(cache *STSCache) error {
	path, err := STSPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling STS cache to JSON: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Write atomically with owner-only 0600 permissions
	tmpFile, err := os.CreateTemp(dir, "sts_cache.*.json.tmp")
	if err != nil {
		return fmt.Errorf("creating temporary STS cache file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}()

	// Set owner-only permissions on temp file before writing data
	if err := os.Chmod(tmpPath, 0600); err != nil {
		return fmt.Errorf("securing temporary STS cache file: %w", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("writing temporary STS cache file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("syncing temporary STS cache file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temporary STS cache file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("moving temporary STS cache file to destination: %w", err)
	}

	return nil
}

// ListMFADevices queries IAM to automatically discover the user's active MFA serial ARN.
func (p *Provider) ListMFADevices(ctx context.Context) (string, error) {
	if _, err := p.lookPathFunc("aws"); err != nil {
		return "", fmt.Errorf("aws CLI not found in PATH: %w", err)
	}

	args := []string{"iam", "list-mfa-devices", "--output", "json"}
	if p.profile != "" {
		args = append(args, "--profile", p.profile)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("aws iam list-mfa-devices failed: %w (stderr: %q)", err, stderr.String())
	}

	var output struct {
		MFADevices []struct {
			SerialNumber string `json:"serialNumber"`
		} `json:"mfaDevices"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return "", fmt.Errorf("failed to parse IAM MFA devices JSON: %w", err)
	}

	if len(output.MFADevices) == 0 {
		return "", fmt.Errorf("no registered MFA devices found for the active AWS profile. Please configure mfa_serial manually in config.yaml")
	}

	return output.MFADevices[0].SerialNumber, nil
}

// GetSessionToken performs the STS authentication token exchange using the MFA code.
func (p *Provider) GetSessionToken(ctx context.Context, serialNumber, tokenCode string, duration int) (*STSCredentials, error) {
	if _, err := p.lookPathFunc("aws"); err != nil {
		return nil, fmt.Errorf("aws CLI not found in PATH: %w", err)
	}

	if duration <= 0 {
		duration = 3600 // Default to 1 hour
	}

	args := []string{
		"sts", "get-session-token",
		"--serial-number", serialNumber,
		"--token-code", tokenCode,
		"--duration-seconds", fmt.Sprint(duration),
		"--output", "json",
	}

	// Session token fetches are region-agnostic but require active profile if set
	if p.profile != "" {
		args = append(args, "--profile", p.profile)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("aws sts get-session-token failed: %w (stderr: %q)", err, stderr.String())
	}

	var output struct {
		Credentials struct {
			AccessKeyId     string    `json:"accessKeyId"`
			SecretAccessKey string    `json:"secretAccessKey"`
			SessionToken    string    `json:"sessionToken"`
			Expiration      time.Time `json:"expiration"`
		} `json:"credentials"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, fmt.Errorf("failed to parse STS credentials JSON: %w", err)
	}

	return &STSCredentials{
		AccessKeyID:     output.Credentials.AccessKeyId,
		SecretAccessKey: output.Credentials.SecretAccessKey,
		SessionToken:    output.Credentials.SessionToken,
		Expiration:      output.Credentials.Expiration,
	}, nil
}

// UpdateAWSCredentialsFile updates the AWS credentials file (~/.aws/credentials) with temporary STS keys.
func UpdateAWSCredentialsFile(profileName string, creds *STSCredentials) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolving home directory: %w", err)
	}
	path := filepath.Join(home, ".aws", "credentials")

	var lines []string
	data, err := os.ReadFile(path)
	if err == nil {
		lines = strings.Split(string(data), "\n")
	}

	sectionHeader := fmt.Sprintf("[%s]", profileName)
	var newLines []string
	inSection := false
	sectionFound := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if inSection {
				newLines = append(newLines, fmt.Sprintf("aws_access_key_id = %s", creds.AccessKeyID))
				newLines = append(newLines, fmt.Sprintf("aws_secret_access_key = %s", creds.SecretAccessKey))
				newLines = append(newLines, fmt.Sprintf("aws_session_token = %s", creds.SessionToken))
				newLines = append(newLines, "")
				inSection = false
			}

			if trimmed == sectionHeader {
				inSection = true
				sectionFound = true
				newLines = append(newLines, line)
				continue
			}
		}

		if inSection {
			if strings.HasPrefix(trimmed, "aws_access_key_id") ||
				strings.HasPrefix(trimmed, "aws_secret_access_key") ||
				strings.HasPrefix(trimmed, "aws_session_token") {
				continue
			}
		}

		newLines = append(newLines, line)
	}

	if inSection {
		newLines = append(newLines, fmt.Sprintf("aws_access_key_id = %s", creds.AccessKeyID))
		newLines = append(newLines, fmt.Sprintf("aws_secret_access_key = %s", creds.SecretAccessKey))
		newLines = append(newLines, fmt.Sprintf("aws_session_token = %s", creds.SessionToken))
	}

	if !sectionFound {
		if len(newLines) > 0 && newLines[len(newLines)-1] != "" {
			newLines = append(newLines, "")
		}
		newLines = append(newLines, sectionHeader)
		newLines = append(newLines, fmt.Sprintf("aws_access_key_id = %s", creds.AccessKeyID))
		newLines = append(newLines, fmt.Sprintf("aws_secret_access_key = %s", creds.SecretAccessKey))
		newLines = append(newLines, fmt.Sprintf("aws_session_token = %s", creds.SessionToken))
		newLines = append(newLines, "")
	}

	output := strings.Join(newLines, "\n")
	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, 0755)

	return os.WriteFile(path, []byte(output), 0600)
}
