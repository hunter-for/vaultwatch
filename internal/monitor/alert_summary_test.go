package monitor

import (
	"strings"
	"testing"
	"time"
)

func makeSummaryHistory(t *testing.T) *AlertHistory {
	t.Helper()
	h := NewAlertHistory(10)
	events := []struct {
		leaseID  string
		severity Severity
	}{
		{"lease/critical/1", SeverityCritical},
		{"lease/warning/1", SeverityWarning},
		{"lease/critical/2", SeverityCritical},
		{"lease/info/1", SeverityInfo},
	}
	for _, e := range events {
		h.Record(AlertEvent{
			LeaseID:  e.leaseID,
			Severity: e.severity,
			FiredAt:  time.Now().UTC(),
		})
	}
	return h
}

func TestAlertSummary_TotalCount(t *testing.T) {
	h := makeSummaryHistory(t)
	s := NewAlertSummary(h, nil)
	if s.TotalAlerts != 4 {
		t.Errorf("expected 4 total alerts, got %d", s.TotalAlerts)
	}
}

func TestAlertSummary_BySeverity(t *testing.T) {
	h := makeSummaryHistory(t)
	s := NewAlertSummary(h, nil)
	if s.BySeverity[SeverityCritical] != 2 {
		t.Errorf("expected 2 critical, got %d", s.BySeverity[SeverityCritical])
	}
	if s.BySeverity[SeverityWarning] != 1 {
		t.Errorf("expected 1 warning, got %d", s.BySeverity[SeverityWarning])
	}
	if s.BySeverity[SeverityInfo] != 1 {
		t.Errorf("expected 1 info, got %d", s.BySeverity[SeverityInfo])
	}
}

func TestAlertSummary_SuppressedFromDedup(t *testing.T) {
	h := makeSummaryHistory(t)
	d := NewDedupStore(1 * time.Minute)
	// Trigger a suppression: record first, then attempt duplicate.
	d.ShouldSend("lease/x", SeverityWarning)
	d.ShouldSend("lease/x", SeverityWarning) // suppressed

	s := NewAlertSummary(h, d)
	if s.Suppressed != 1 {
		t.Errorf("expected 1 suppressed, got %d", s.Suppressed)
	}
}

func TestAlertSummary_FormatContainsKeyFields(t *testing.T) {
	h := makeSummaryHistory(t)
	s := NewAlertSummary(h, nil)
	out := s.Format()

	for _, want := range []string{"Alert Summary", "Total Alerts", "Critical", "Warning", "Recent Events"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}
}

func TestAlertSummary_NilDedupOK(t *testing.T) {
	h := NewAlertHistory(5)
	s := NewAlertSummary(h, nil)
	if s.Suppressed != 0 {
		t.Errorf("expected 0 suppressed with nil dedup, got %d", s.Suppressed)
	}
}
