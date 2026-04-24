package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration for vaultwatch.
type Config struct {
	Vault   VaultConfig   `yaml:"vault"`
	Alerts  AlertsConfig  `yaml:"alerts"`
	Monitor MonitorConfig `yaml:"monitor"`
}

// VaultConfig contains Vault connection settings.
type VaultConfig struct {
	Address   string `yaml:"address"`
	Token     string `yaml:"token"`
	Namespace string `yaml:"namespace"`
}

// AlertsConfig defines alerting thresholds and channels.
type AlertsConfig struct {
	WarnBefore  time.Duration `yaml:"warn_before"`
	SlackWebhook string       `yaml:"slack_webhook"`
	Email        string       `yaml:"email"`
}

// MonitorConfig controls polling behavior.
type MonitorConfig struct {
	Interval time.Duration `yaml:"interval"`
	Paths    []string      `yaml:"paths"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// validate checks that required fields are present and sensible.
func (c *Config) validate() error {
	if c.Vault.Address == "" {
		return fmt.Errorf("vault.address is required")
	}
	if c.Vault.Token == "" {
		return fmt.Errorf("vault.token is required")
	}
	if c.Monitor.Interval <= 0 {
		c.Monitor.Interval = 60 * time.Second
	}
	if c.Alerts.WarnBefore <= 0 {
		c.Alerts.WarnBefore = 24 * time.Hour
	}
	return nil
}
