package monitor_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// TestAlertHistory_WithClassify verifies that AlertHistory integrates cleanly
// with the Classify function — events classified from real lease statuses are
// recorded and retrievable with correct severity counts.
func TestAlertHistory_WithClassify(t *testing.T) {
	h := monitor.NewAlertHistory(50)
	thresholds := monitor.DefaultThresholds()

	leases := []struct {
		id  string
		ttl time.Duration
	}{
		{"lease-critical-1", 3 * time.Minute},
		{"lease-critical-2", 1 * time.Minute},
		{"lease-warning-1", 15 * time.Minute},
		{"lease-info-1", 45 * time.Minute},
	}

	for _, l := range leases {
		sev := monitor.Classify(l.ttl, thresholds)
		h.Record(monitor.AlertEvent{
			LeaseID:  l.id,
			Severity: sev,
			SentAt:   time.Now(),
			TTL:      l.ttl,
		})
	}

	if h.Len() != 4 {
		t.Fatalf("expected 4 events, got %d", h.Len())
	}

	counts := h.CountBySeverity()
	if counts[monitor.SeverityCritical] != 2 {
		t.Errorf("expected 2 critical events, got %d", counts[monitor.SeverityCritical])
	}
	if counts[monitor.SeverityWarning] != 1 {
		t.Errorf("expected 1 warning event, got %d", counts[monitor.SeverityWarning])
	}
	if counts[monitor.SeverityInfo] != 1 {
		t.Errorf("expected 1 info event, got %d", counts[monitor.SeverityInfo])
	}

	// Most recent should be the info lease (last recorded).
	recent := h.Recent(1)
	if recent[0].LeaseID != "lease-info-1" {
		t.Errorf("expected most recent to be lease-info-1, got %s", recent[0].LeaseID)
	}
}
