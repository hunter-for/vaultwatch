package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/alert"
	"github.com/user/vaultwatch/internal/monitor"
)

func baseWebhookAlert() monitor.Alert {
	return monitor.Alert{
		LeaseID:  "secret/data/db#abc123",
		TTL:      2 * time.Hour,
		Severity: monitor.SeverityWarning,
		Message:  "lease expiring soon",
		Timestamp: time.Now().UTC(),
	}
}

func TestWebhookSender_NotNil(t *testing.T) {
	s := alert.NewWebhookSender("http://example.com/hook")
	if s == nil {
		t.Fatal("expected non-nil WebhookSender")
	}
}

func TestWebhookSender_SendSuccess(t *testing.T) {
	received := make(chan map[string]interface{}, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		received <- payload
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sender := alert.NewWebhookSender(ts.URL)
	a := baseWebhookAlert()

	if err := sender.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payload := <-received
	if payload["level"] != string(monitor.SeverityWarning) {
		t.Errorf("expected level %q, got %q", monitor.SeverityWarning, payload["level"])
	}
	if payload["lease_id"] != a.LeaseID {
		t.Errorf("expected lease_id %q, got %q", a.LeaseID, payload["lease_id"])
	}
}

func TestWebhookSender_SendFailsOnBadStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	sender := alert.NewWebhookSender(ts.URL)
	if err := sender.Send(baseWebhookAlert()); err == nil {
		t.Fatal("expected error on non-2xx response")
	}
}

func TestWebhookSender_SendFailsOnBadURL(t *testing.T) {
	sender := alert.NewWebhookSender("http://127.0.0.1:0/nope")
	if err := sender.Send(baseWebhookAlert()); err == nil {
		t.Fatal("expected error on unreachable URL")
	}
}
