package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/monitor"
)

func baseSlackAlert() monitor.Alert {
	return monitor.Alert{
		LeaseID:   "database/creds/my-role/abc123",
		TTL:       10 * time.Minute,
		Severity:  monitor.SeverityWarning,
		Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
	}
}

func TestSlackSender_NotNil(t *testing.T) {
	s := NewSlackSender("https://hooks.slack.com/services/fake")
	if s == nil {
		t.Fatal("expected non-nil SlackSender")
	}
}

func TestSlackSender_SendSuccess(t *testing.T) {
	var received slackPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewSlackSender(server.URL)
	if err := s.Send(baseSlackAlert()); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !strings.Contains(received.Text, "database/creds/my-role/abc123") {
		t.Errorf("expected payload to contain lease ID, got: %s", received.Text)
	}
	if !strings.Contains(received.Text, "warning") {
		t.Errorf("expected payload to contain severity, got: %s", received.Text)
	}
}

func TestSlackSender_SendFailsOnBadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s := NewSlackSender(server.URL)
	if err := s.Send(baseSlackAlert()); err == nil {
		t.Fatal("expected error on non-2xx status, got nil")
	}
}

func TestSlackSender_SendFailsOnBadURL(t *testing.T) {
	s := NewSlackSender("http://127.0.0.1:0/invalid")
	if err := s.Send(baseSlackAlert()); err == nil {
		t.Fatal("expected error on bad URL, got nil")
	}
}
