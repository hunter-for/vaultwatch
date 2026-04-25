package monitor_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// TestRateLimiter_WithClassify verifies that the RateLimiter correctly
// throttles alerts when integrated with Classify-derived severities.
func TestRateLimiter_WithClassify(t *testing.T) {
	thresholds := monitor.DefaultThresholds()
	rl := monitor.NewRateLimiter(2, 5*time.Minute)

	leases := []struct {
		id  string
		ttl time.Duration
	}{
		{"lease/critical/1", 3 * time.Minute},
		{"lease/warning/1", 20 * time.Minute},
		{"lease/info/1", 2 * time.Hour},
	}

	for _, l := range leases {
		sev := monitor.Classify(l.ttl, thresholds)
		if sev == monitor.SeverityInfo {
			// info leases: we still rate-limit them
			if !rl.Allow(l.id) {
				t.Errorf("expected first alert allowed for %s", l.id)
			}
			continue
		}
		// critical and warning: allow up to burst
		if !rl.Allow(l.id) {
			t.Errorf("expected first alert allowed for %s (sev=%s)", l.id, sev)
		}
		if !rl.Allow(l.id) {
			t.Errorf("expected second alert allowed for %s (sev=%s)", l.id, sev)
		}
		if rl.Allow(l.id) {
			t.Errorf("expected third alert denied for %s (sev=%s)", l.id, sev)
		}
	}
}

// TestRateLimiter_DoesNotBlockDedupEscalation ensures that resetting
// a rate-limit bucket (e.g., on severity escalation) restores capacity.
func TestRateLimiter_DoesNotBlockDedupEscalation(t *testing.T) {
	rl := monitor.NewRateLimiter(1, 10*time.Minute)
	leaseID := "lease/escalation"

	// exhaust the bucket
	rl.Allow(leaseID)
	if rl.Allow(leaseID) {
		t.Fatal("expected rate limit to be exhausted")
	}

	// simulate escalation: caller resets the bucket
	rl.Reset(leaseID)

	if !rl.Allow(leaseID) {
		t.Error("expected alert to be allowed after escalation reset")
	}
}
