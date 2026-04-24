package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/vaultwatch/internal/monitor"
)

// SlackSender sends alert notifications to a Slack webhook URL.
type SlackSender struct {
	webhookURL string
	client     *http.Client
}

// slackPayload represents the JSON body sent to Slack's incoming webhook API.
type slackPayload struct {
	Text string `json:"text"`
}

// NewSlackSender creates a new SlackSender targeting the given Slack webhook URL.
func NewSlackSender(webhookURL string) *SlackSender {
	return &SlackSender{
		webhookURL: webhookURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send formats the alert and posts it to the configured Slack webhook.
func (s *SlackSender) Send(a monitor.Alert) error {
	msg := fmt.Sprintf("[%s] VaultWatch Alert\nLease: %s\nTTL: %s\nSeverity: %s",
		a.Timestamp.Format(time.RFC3339),
		a.LeaseID,
		a.TTL.String(),
		a.Severity,
	)

	payload := slackPayload{Text: msg}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: failed to marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack: unexpected status code %d", resp.StatusCode)
	}

	return nil
}
