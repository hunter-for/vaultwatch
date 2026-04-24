package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

const pagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutySender sends alerts to PagerDuty via the Events API v2.
type PagerDutySender struct {
	integrationKey string
	client         *http.Client
	eventsURL      string
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary  string `json:"summary"`
	Source   string `json:"source"`
	Severity string `json:"severity"`
	Timestamp string `json:"timestamp"`
}

// NewPagerDutySender creates a new PagerDutySender with the given integration key.
func NewPagerDutySender(integrationKey string) *PagerDutySender {
	return &PagerDutySender{
		integrationKey: integrationKey,
		client:         &http.Client{Timeout: 10 * time.Second},
		eventsURL:      pagerDutyEventsURL,
	}
}

// Send dispatches an alert event to PagerDuty.
func (p *PagerDutySender) Send(a monitor.Alert) error {
	severity := pdSeverity(a.Severity)
	body := pdPayload{
		RoutingKey:  p.integrationKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:   fmt.Sprintf("Vault lease expiring: %s (TTL: %s)", a.LeaseID, a.TTL),
			Source:    "vaultwatch",
			Severity:  severity,
			Timestamp: a.Timestamp.UTC().Format(time.RFC3339),
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal payload: %w", err)
	}

	resp, err := p.client.Post(p.eventsURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func pdSeverity(s monitor.Severity) string {
	switch s {
	case monitor.SeverityCritical:
		return "critical"
	case monitor.SeverityWarning:
		return "warning"
	default:
		return "info"
	}
}
