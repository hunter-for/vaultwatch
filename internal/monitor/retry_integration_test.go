package monitor

import (
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// countingSender counts sends and fails the first N times.
type countingSender struct {
	attempts  int
	failFirst int
}

func (c *countingSender) Send(_ alert.Alert) error {
	c.attempts++
	if c.attempts <= c.failFirst {
		return errors.New("simulated failure")
	}
	return nil
}

func TestRetrySender_WithClassify_EventualSuccess(t *testing.T) {
	lease := renewableLeaseWithTTL("lease-retry-1", 4*time.Minute)
	severity := Classify(lease, DefaultThresholds())
	if severity == Info {
		t.Skip("lease not in alert range for this test")
	}

	a := alert.Alert{
		LeaseID:  lease.LeaseID,
		Severity: alert.Severity(severity.String()),
		TTL:      lease.TTL,
	}

	inner := &countingSender{failFirst: 2}
	policy := RetryPolicy{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond, Multiplier: 2.0}
	rs := NewRetrySender(inner, policy)
	rs.sleep = noSleep

	if err := rs.Send(a); err != nil {
		t.Errorf("expected eventual success, got %v", err)
	}
	if inner.attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", inner.attempts)
	}
}

func TestRetrySender_WithClassify_ExhaustsAndErrors(t *testing.T) {
	lease := renewableLeaseWithTTL("lease-retry-2", 4*time.Minute)
	a := alert.Alert{
		LeaseID:  lease.LeaseID,
		Severity: alert.Critical,
		TTL:      lease.TTL,
	}

	inner := &countingSender{failFirst: 99}
	policy := RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond, Multiplier: 2.0}
	rs := NewRetrySender(inner, policy)
	rs.sleep = noSleep

	err := rs.Send(a)
	if !errors.Is(err, ErrMaxRetriesExceeded) {
		t.Errorf("expected ErrMaxRetriesExceeded, got %v", err)
	}
}
