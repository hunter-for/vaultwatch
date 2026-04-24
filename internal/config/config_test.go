package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "vaultwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	content := `
vault:
  address: "https://vault.example.com"
  token: "s.abc123"
  namespace: "admin"
monitor:
  interval: 30s
  paths:
    - secret/myapp
alerts:
  warn_before: 48h
  slack_webhook: "https://hooks.slack.com/test"
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "https://vault.example.com" {
		t.Errorf("expected vault address, got %q", cfg.Vault.Address)
	}
	if cfg.Monitor.Interval != 30*time.Second {
		t.Errorf("expected 30s interval, got %v", cfg.Monitor.Interval)
	}
	if cfg.Alerts.WarnBefore != 48*time.Hour {
		t.Errorf("expected 48h warn_before, got %v", cfg.Alerts.WarnBefore)
	}
}

func TestLoad_MissingAddress(t *testing.T) {
	content := `
vault:
  token: "s.abc123"
`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing vault address, got nil")
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	content := `
vault:
  address: "https://vault.example.com"
  token: "s.abc123"
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Monitor.Interval != 60*time.Second {
		t.Errorf("expected default interval 60s, got %v", cfg.Monitor.Interval)
	}
	if cfg.Alerts.WarnBefore != 24*time.Hour {
		t.Errorf("expected default warn_before 24h, got %v", cfg.Alerts.WarnBefore)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
