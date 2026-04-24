package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all vaultwatch runtime configuration.
type Config struct {
	Vault   VaultConfig   `yaml:"vault"`
	Monitor MonitorConfig `yaml:"monitor"`
	Alert   AlertConfig   `yaml:"alert"`
}

// VaultConfig holds Vault connection settings.
type VaultConfig struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
}

// MonitorConfig controls polling and threshold behaviour.
type MonitorConfig struct {
	Interval        time.Duration `yaml:"interval"`
	CriticalSeconds int           `yaml:"critical_seconds"`
	WarningSeconds  int           `yaml:"warning_seconds"`
}

// AlertConfig holds optional notification channel settings.
type AlertConfig struct {
	SlackWebhook    string `yaml:"slack_webhook"`
	WebhookURL      string `yaml:"webhook_url"`
	PagerDutyKey    string `yaml:"pagerduty_key"`
	SMTPHost        string `yaml:"smtp_host"`
	SMTPPort        int    `yaml:"smtp_port"`
	SMTPFrom        string `yaml:"smtp_from"`
	SMTPTo          string `yaml:"smtp_to"`
	SMTPUsername    string `yaml:"smtp_username"`
	SMTPPassword    string `yaml:"smtp_password"`
}

const (
	defaultInterval        = 60 * time.Second
	defaultCriticalSeconds = 300
	defaultWarningSeconds  = 3600
)

// Load reads a YAML config file from path and applies defaults.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Vault.Address == "" {
		return nil, errors.New("vault.address is required")
	}

	// Apply defaults for optional numeric fields.
	if cfg.Monitor.Interval <= 0 {
		cfg.Monitor.Interval = defaultInterval
	}
	if cfg.Monitor.CriticalSeconds == 0 {
		cfg.Monitor.CriticalSeconds = defaultCriticalSeconds
	}
	if cfg.Monitor.WarningSeconds == 0 {
		cfg.Monitor.WarningSeconds = defaultWarningSeconds
	}

	return &cfg, nil
}
