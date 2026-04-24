package monitor

import (
	"testing"
	"time"
)

func TestNewLeaseStatus_Fields(t *testing.T) {
	ttl := 30 * time.Minute
	lease := NewLeaseStatus("lease/abc123", "secret/data/myapp", ttl, true)

	if lease.LeaseID != "lease/abc123" {
		t.Errorf("expected LeaseID %q, got %q", "lease/abc123", lease.LeaseID)
	}
	if lease.Path != "secret/data/myapp" {
		t.Errorf("expected Path %q, got %q", "secret/data/myapp", lease.Path)
	}
	if lease.TTL != ttl {
		t.Errorf("expected TTL %s, got %s", ttl, lease.TTL)
	}
	if !lease.Renewable {
		t.Error("expected Renewable to be true")
	}
}

func TestLeaseStatus_IsExpired_False(t *testing.T) {
	lease := NewLeaseStatus("lease/x", "secret/x", 10*time.Minute, false)
	if lease.IsExpired() {
		t.Error("expected lease to not be expired")
	}
}

func TestLeaseStatus_IsExpired_True(t *testing.T) {
	lease := LeaseStatus{
		LeaseID:  "lease/old",
		Path:     "secret/old",
		TTL:      1 * time.Second,
		ExpireAt: time.Now().Add(-1 * time.Minute),
	}
	if !lease.IsExpired() {
		t.Error("expected lease to be expired")
	}
}

func TestLeaseStatus_TimeRemaining_Positive(t *testing.T) {
	ttl := 15 * time.Minute
	lease := NewLeaseStatus("lease/y", "secret/y", ttl, true)
	remaining := lease.TimeRemaining()
	if remaining <= 0 {
		t.Errorf("expected positive time remaining, got %s", remaining)
	}
	if remaining > ttl {
		t.Errorf("time remaining %s exceeds original TTL %s", remaining, ttl)
	}
}

func TestLeaseStatus_TimeRemaining_Zero_WhenExpired(t *testing.T) {
	lease := LeaseStatus{
		LeaseID:  "lease/expired",
		ExpireAt: time.Now().Add(-5 * time.Minute),
	}
	if lease.TimeRemaining() != 0 {
		t.Errorf("expected 0 remaining for expired lease, got %s", lease.TimeRemaining())
	}
}

func TestLeaseStatus_String_Contains_ID(t *testing.T) {
	lease := NewLeaseStatus("lease/abc", "secret/db", 5*time.Minute, false)
	s := lease.String()
	if s == "" {
		t.Error("expected non-empty string representation")
	}
	for _, substr := range []string{"lease/abc", "secret/db"} {
		if len(s) == 0 {
			t.Errorf("String() missing expected content %q", substr)
		}
	}
}
