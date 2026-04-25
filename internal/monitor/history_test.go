package monitor

import (
	"testing"
	"time"
)

func makeEvent(leaseID string, sev Severity, ttl time.Duration) AlertEvent {
	return AlertEvent{
		LeaseID:  leaseID,
		Severity: sev,
		SentAt:   time.Now(),
		TTL:      ttl,
	}
}

func TestAlertHistory_RecordAndLen(t *testing.T) {
	h := NewAlertHistory(10)
	if h.Len() != 0 {
		t.Fatalf("expected 0, got %d", h.Len())
	}
	h.Record(makeEvent("lease-1", SeverityCritical, 5*time.Minute))
	h.Record(makeEvent("lease-2", SeverityWarning, 20*time.Minute))
	if h.Len() != 2 {
		t.Fatalf("expected 2, got %d", h.Len())
	}
}

func TestAlertHistory_EvictsOldestWhenFull(t *testing.T) {
	h := NewAlertHistory(3)
	for i := 0; i < 4; i++ {
		h.Record(makeEvent("lease", SeverityInfo, time.Duration(i)*time.Minute))
	}
	if h.Len() != 3 {
		t.Fatalf("expected 3 after eviction, got %d", h.Len())
	}
	// Oldest (TTL=0) should have been evicted; newest TTL should be 3m.
	recent := h.Recent(1)
	if recent[0].TTL != 3*time.Minute {
		t.Errorf("expected newest TTL 3m, got %v", recent[0].TTL)
	}
}

func TestAlertHistory_RecentNewestFirst(t *testing.T) {
	h := NewAlertHistory(10)
	h.Record(makeEvent("a", SeverityInfo, 1*time.Minute))
	h.Record(makeEvent("b", SeverityWarning, 2*time.Minute))
	h.Record(makeEvent("c", SeverityCritical, 3*time.Minute))

	recent := h.Recent(2)
	if len(recent) != 2 {
		t.Fatalf("expected 2 results, got %d", len(recent))
	}
	if recent[0].LeaseID != "c" {
		t.Errorf("expected newest first (c), got %s", recent[0].LeaseID)
	}
	if recent[1].LeaseID != "b" {
		t.Errorf("expected second newest (b), got %s", recent[1].LeaseID)
	}
}

func TestAlertHistory_CountBySeverity(t *testing.T) {
	h := NewAlertHistory(20)
	h.Record(makeEvent("l1", SeverityCritical, time.Minute))
	h.Record(makeEvent("l2", SeverityCritical, time.Minute))
	h.Record(makeEvent("l3", SeverityWarning, 10*time.Minute))
	h.Record(makeEvent("l4", SeverityInfo, 30*time.Minute))

	counts := h.CountBySeverity()
	if counts[SeverityCritical] != 2 {
		t.Errorf("expected 2 critical, got %d", counts[SeverityCritical])
	}
	if counts[SeverityWarning] != 1 {
		t.Errorf("expected 1 warning, got %d", counts[SeverityWarning])
	}
	if counts[SeverityInfo] != 1 {
		t.Errorf("expected 1 info, got %d", counts[SeverityInfo])
	}
}

func TestAlertHistory_DefaultMaxLen(t *testing.T) {
	h := NewAlertHistory(0) // should default to 100
	for i := 0; i < 105; i++ {
		h.Record(makeEvent("lease", SeverityInfo, time.Minute))
	}
	if h.Len() != 100 {
		t.Errorf("expected max 100 events, got %d", h.Len())
	}
}
