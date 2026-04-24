package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

func basePDAlert() monitor.Alert {
	return monitor.Alert{
		LeaseID:   "aws/creds/my-role/abc123",
		TTL:       5 * time.Minute,
		Severity:  monitor.SeverityCritical,
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
}

func TestPagerDutySender_NotNil(t *testing.T) {
	s := NewPagerDutySender("test-key")
	if s == nil {
		t.Fatal("expected non-nil PagerDutySender")
	}
}

func TestPagerDutySender_SendSuccess(t *testing.T) {
	var received map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	s := NewPagerDutySender("test-integration-key")
	s.eventsURL = srv.URL

	if err := s.Send(basePDAlert()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["routing_key"] != "test-integration-key" {
		t.Errorf("expected routing_key=test-integration-key, got %v", received["routing_key"])
	}
	if received["event_action"] != "trigger" {
		t.Errorf("expected event_action=trigger, got %v", received["event_action"])
	}
}

func TestPagerDutySender_SeverityMapping(t *testing.T) {
	cases := []struct {
		severity monitor.Severity
		want     string
	}{
		{monitor.SeverityCritical, "critical"},
		{monitor.SeverityWarning, "warning"},
		{monitor.SeverityInfo, "info"},
	}
	for _, tc := range cases {
		got := pdSeverity(tc.severity)
		if got != tc.want {
			t.Errorf("pdSeverity(%v) = %q, want %q", tc.severity, got, tc.want)
		}
	}
}

func TestPagerDutySender_SendFailsOnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := NewPagerDutySender("key")
	s.eventsURL = srv.URL

	if err := s.Send(basePDAlert()); err == nil {
		t.Fatal("expected error on non-2xx status")
	}
}

func TestPagerDutySender_SendFailsOnBadURL(t *testing.T) {
	s := NewPagerDutySender("key")
	s.eventsURL = "http://127.0.0.1:0"

	if err := s.Send(basePDAlert()); err == nil {
		t.Fatal("expected error on unreachable URL")
	}
}
