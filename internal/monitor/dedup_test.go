package monitor

import (
	"testing"
	"time"
)

func TestDedupStore_FirstAlertAlwaysSent(t *testing.T) {
	d := NewDedupStore()
	if !d.ShouldAlert("lease/1", SeverityWarning, time.Minute) {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestDedupStore_DuplicateWithinCooldownSuppressed(t *testing.T) {
	d := NewDedupStore()
	d.ShouldAlert("lease/1", SeverityWarning, time.Hour)
	if d.ShouldAlert("lease/1", SeverityWarning, time.Hour) {
		t.Fatal("expected duplicate alert within cooldown to be suppressed")
	}
}

func TestDedupStore_AlertAllowedAfterCooldown(t *testing.T) {
	d := NewDedupStore()
	// Use a zero cooldown so the window has already passed.
	d.ShouldAlert("lease/1", SeverityWarning, 0)
	if !d.ShouldAlert("lease/1", SeverityWarning, 0) {
		t.Fatal("expected alert to be allowed after cooldown expires")
	}
}

func TestDedupStore_EscalationBypassesCooldown(t *testing.T) {
	d := NewDedupStore()
	d.ShouldAlert("lease/1", SeverityWarning, time.Hour)
	// Escalate to Critical — should bypass the cooldown.
	if !d.ShouldAlert("lease/1", SeverityCritical, time.Hour) {
		t.Fatal("expected escalated severity to bypass cooldown")
	}
}

func TestDedupStore_SameSeverityNotEscalation(t *testing.T) {
	d := NewDedupStore()
	d.ShouldAlert("lease/1", SeverityCritical, time.Hour)
	if d.ShouldAlert("lease/1", SeverityCritical, time.Hour) {
		t.Fatal("expected same severity within cooldown to be suppressed")
	}
}

func TestDedupStore_EvictRemovesEntry(t *testing.T) {
	d := NewDedupStore()
	d.ShouldAlert("lease/1", SeverityWarning, time.Hour)
	if d.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", d.Len())
	}
	d.Evict("lease/1")
	if d.Len() != 0 {
		t.Fatalf("expected 0 entries after evict, got %d", d.Len())
	}
}

func TestDedupStore_MultipleLeases(t *testing.T) {
	d := NewDedupStore()
	d.ShouldAlert("lease/1", SeverityInfo, time.Hour)
	d.ShouldAlert("lease/2", SeverityWarning, time.Hour)
	d.ShouldAlert("lease/3", SeverityCritical, time.Hour)
	if d.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", d.Len())
	}
}

func TestDedupStore_EvictNonExistentIsNoop(t *testing.T) {
	d := NewDedupStore()
	// Should not panic.
	d.Evict("lease/nonexistent")
	if d.Len() != 0 {
		t.Fatalf("expected 0 entries, got %d", d.Len())
	}
}
