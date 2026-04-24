package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/config"
	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func main() {
	cfgPath := flag.String("config", "vaultwatch.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	client, err := vault.NewClient(cfg.Vault.Address, cfg.Vault.Token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating vault client: %v\n", err)
		os.Exit(1)
	}

	senders := buildSenders(cfg)
	sender := alert.NewMultiSender(senders...)

	m := monitor.New(client, sender, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ticker := time.NewTicker(cfg.Monitor.Interval)
	defer ticker.Stop()

	fmt.Printf("vaultwatch started (interval: %s)\n", cfg.Monitor.Interval)

	for {
		select {
		case <-ticker.C:
			if err := m.Run(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "monitor run error: %v\n", err)
			}
		case <-ctx.Done():
			fmt.Println("vaultwatch shutting down")
			return
		}
	}
}

func buildSenders(cfg *config.Config) []alert.Sender {
	var senders []alert.Sender

	senders = append(senders, alert.NewStdoutSender())

	if cfg.Alerts.Slack.WebhookURL != "" {
		senders = append(senders, alert.NewSlackSender(cfg.Alerts.Slack.WebhookURL))
	}

	if cfg.Alerts.Webhook.URL != "" {
		senders = append(senders, alert.NewWebhookSender(cfg.Alerts.Webhook.URL))
	}

	if cfg.Alerts.Email.Host != "" {
		senders = append(senders, alert.NewEmailSender(
			cfg.Alerts.Email.Host,
			cfg.Alerts.Email.Port,
			cfg.Alerts.Email.Username,
			cfg.Alerts.Email.Password,
			cfg.Alerts.Email.From,
			cfg.Alerts.Email.To,
		))
	}

	return senders
}
