package main

import (
	"testing"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/config"
)

func TestBuildSenders_StdoutOnly(t *testing.T) {
	cfg := &config.Config{}
	senders := buildSenders(cfg)

	if len(senders) != 1 {
		t.Fatalf("expected 1 sender (stdout), got %d", len(senders))
	}
}

func TestBuildSenders_WithSlack(t *testing.T) {
	cfg := &config.Config{}
	cfg.Alerts.Slack.WebhookURL = "https://hooks.slack.com/test"

	senders := buildSenders(cfg)
	if len(senders) != 2 {
		t.Fatalf("expected 2 senders (stdout + slack), got %d", len(senders))
	}
}

func TestBuildSenders_WithWebhook(t *testing.T) {
	cfg := &config.Config{}
	cfg.Alerts.Webhook.URL = "https://example.com/hook"

	senders := buildSenders(cfg)
	if len(senders) != 2 {
		t.Fatalf("expected 2 senders (stdout + webhook), got %d", len(senders))
	}
}

func TestBuildSenders_AllConfigured(t *testing.T) {
	cfg := &config.Config{}
	cfg.Alerts.Slack.WebhookURL = "https://hooks.slack.com/test"
	cfg.Alerts.Webhook.URL = "https://example.com/hook"
	cfg.Alerts.Email.Host = "smtp.example.com"
	cfg.Alerts.Email.Port = 587
	cfg.Alerts.Email.From = "from@example.com"
	cfg.Alerts.Email.To = "to@example.com"

	senders := buildSenders(cfg)
	if len(senders) != 4 {
		t.Fatalf("expected 4 senders, got %d", len(senders))
	}
}

func TestBuildSenders_ReturnsAlertSenderInterface(t *testing.T) {
	cfg := &config.Config{}
	senders := buildSenders(cfg)

	for i, s := range senders {
		if s == nil {
			t.Errorf("sender at index %d is nil", i)
		}
		if _, ok := s.(alert.Sender); !ok {
			t.Errorf("sender at index %d does not implement alert.Sender", i)
		}
	}
}
