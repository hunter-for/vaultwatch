package monitor

import (
	"testing"
	"time"
)

func TestCooldownTracker_AllowsFirstAlert(t *testing.T) {
	tracker := NewCooldownTracker(DefaultCooldownPolicy())
	if !tracker.Allow("lease/abc", SeverityWarning) {
		t.Error("expected first alert to be allowed")
	}
}

func TestCooldownTracker_SuppressesWithinCooldown(t *testing.T) {
	policy := CooldownPolicy{
		Critical: 10 * time.Minute,
		Warning:  10 * time.Minute,
		Info:     10 * time.Minute,
	}
	tracker := NewCooldownTracker(policy)
	tracker.Record("lease/abc", SeverityWarning)

	if tracker.Allow("lease/abc", SeverityWarning) {
		t.Error("expected alert to be suppressed within cooldown")
	}
}

func TestCooldownTracker_AllowsAfterCooldownExpires(t *testing.T) {
	policy := CooldownPolicy{
		Critical: 1 * time.Millisecond,
		Warning:  1 * time.Millisecond,
		Info:     1 * time.Millisecond,
	}
	tracker := NewCooldownTracker(policy)
	tracker.Record("lease/abc", SeverityInfo)

	time.Sleep(5 * time.Millisecond)

	if !tracker.Allow("lease/abc", SeverityInfo) {
		t.Error("expected alert to be allowed after cooldown expires")
	}
}

func TestCooldownTracker_DifferentSeveritiesAreIndependent(t *testing.T) {
	policy := CooldownPolicy{
		Critical: 10 * time.Minute,
		Warning:  10 * time.Minute,
		Info:     10 * time.Minute,
	}
	tracker := NewCooldownTracker(policy)
	tracker.Record("lease/abc", SeverityWarning)

	if !tracker.Allow("lease/abc", SeverityCritical) {
		t.Error("expected critical to be allowed independently of warning")
	}
	if tracker.Allow("lease/abc", SeverityWarning) {
		t.Error("expected warning to still be suppressed")
	}
}

func TestCooldownTracker_ResetClearsSeverities(t *testing.T) {
	policy := CooldownPolicy{
		Critical: 10 * time.Minute,
		Warning:  10 * time.Minute,
		Info:     10 * time.Minute,
	}
	tracker := NewCooldownTracker(policy)
	tracker.Record("lease/abc", SeverityCritical)
	tracker.Record("lease/abc", SeverityWarning)
	tracker.Reset("lease/abc")

	if !tracker.Allow("lease/abc", SeverityCritical) {
		t.Error("expected critical to be allowed after reset")
	}
	if !tracker.Allow("lease/abc", SeverityWarning) {
		t.Error("expected warning to be allowed after reset")
	}
}

func TestDefaultCooldownPolicy_Values(t *testing.T) {
	p := DefaultCooldownPolicy()
	if p.Critical >= p.Warning {
		t.Errorf("expected critical (%v) < warning (%v)", p.Critical, p.Warning)
	}
	if p.Warning >= p.Info {
		t.Errorf("expected warning (%v) < info (%v)", p.Warning, p.Info)
	}
}
