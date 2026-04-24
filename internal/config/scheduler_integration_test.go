package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/config"
)

func TestLoad_SchedulerInterval(t *testing.T) {
	const yaml = `
vault:
  address: "http://127.0.0.1:8200"
monitor:
  interval: 30s
  critical_seconds: 120
  warning_seconds: 900
`
	f, err := os.CreateTemp("", "cfg-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(yaml)
	f.Close()

	cfg, err := config.Load(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Monitor.Interval != 30*time.Second {
		t.Errorf("expected 30s interval, got %s", cfg.Monitor.Interval)
	}
	if cfg.Monitor.CriticalSeconds != 120 {
		t.Errorf("expected critical=120, got %d", cfg.Monitor.CriticalSeconds)
	}
	if cfg.Monitor.WarningSeconds != 900 {
		t.Errorf("expected warning=900, got %d", cfg.Monitor.WarningSeconds)
	}
}

func TestLoad_SchedulerDefaults(t *testing.T) {
	const yaml = `
vault:
  address: "http://127.0.0.1:8200"
`
	f, err := os.CreateTemp("", "cfg-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(yaml)
	f.Close()

	cfg, err := config.Load(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Monitor.Interval != 60*time.Second {
		t.Errorf("expected default 60s, got %s", cfg.Monitor.Interval)
	}
	if cfg.Monitor.CriticalSeconds != 300 {
		t.Errorf("expected default critical=300, got %d", cfg.Monitor.CriticalSeconds)
	}
	if cfg.Monitor.WarningSeconds != 3600 {
		t.Errorf("expected default warning=3600, got %d", cfg.Monitor.WarningSeconds)
	}
}
