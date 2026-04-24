package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/vaultwatch/internal/monitor"
)

// WebhookSender sends alerts as JSON POST requests to a configured URL.
type WebhookSender struct {
	url    string
	client *http.Client
}

type webhookPayload struct {
	Level     string    `json:"level"`
	LeaseID   string    `json:"lease_id"`
	TTL       string    `json:"ttl"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// NewWebhookSender creates a WebhookSender that posts to the given URL.
func NewWebhookSender(url string) *WebhookSender {
	return &WebhookSender{
		url: url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send serializes the alert and posts it to the webhook URL.
func (w *WebhookSender) Send(a monitor.Alert) error {
	payload := webhookPayload{
		Level:     string(a.Severity),
		LeaseID:  a.LeaseID,
		TTL:       a.TTL.String(),
		Message:   a.Message,
		Timestamp: a.Timestamp,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d from %s", resp.StatusCode, w.url)
	}

	return nil
}
