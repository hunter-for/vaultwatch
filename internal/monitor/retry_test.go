package monitor

import (
	"testing"
	"time"
)

func TestRetryTracker_ShouldRetry_InitiallyTrue(t *testing.T) {
	rt := NewRetryTracker(DefaultRetryPolicy())
	if !rt.ShouldRetry("lease-1") {
		t.Error("expected ShouldRetry to be true initially")
	}
}

func TestRetryTracker_ShouldRetry_FalseAfterMax(t *testing.T) {
	policy := RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Second, Multiplier: 2.0}
	rt := NewRetryTracker(policy)
	rt.NextDelay("lease-1")
	rt.NextDelay("lease-1")
	if rt.ShouldRetry("lease-1") {
		t.Error("expected ShouldRetry to be false after max attempts")
	}
}

func TestRetryTracker_NextDelay_Doubles(t *testing.T) {
	policy := RetryPolicy{MaxAttempts: 5, InitialDelay: 100 * time.Millisecond, MaxDelay: 10 * time.Second, Multiplier: 2.0}
	rt := NewRetryTracker(policy)
	d1 := rt.NextDelay("lease-1")
	d2 := rt.NextDelay("lease-1")
	if d2 != d1*2 {
		t.Errorf("expected delay to double: got %v after %v", d2, d1)
	}
}

func TestRetryTracker_NextDelay_CapsAtMax(t *testing.T) {
	policy := RetryPolicy{MaxAttempts: 10, InitialDelay: 5 * time.Second, MaxDelay: 8 * time.Second, Multiplier: 2.0}
	rt := NewRetryTracker(policy)
	rt.NextDelay("lease-1")
	d := rt.NextDelay("lease-1")
	if d > policy.MaxDelay {
		t.Errorf("expected delay capped at %v, got %v", policy.MaxDelay, d)
	}
}

func TestRetryTracker_Reset_ClearsState(t *testing.T) {
	rt := NewRetryTracker(DefaultRetryPolicy())
	rt.NextDelay("lease-1")
	rt.NextDelay("lease-1")
	rt.Reset("lease-1")
	if rt.Attempts("lease-1") != 0 {
		t.Error("expected attempts to be 0 after reset")
	}
	if !rt.ShouldRetry("lease-1") {
		t.Error("expected ShouldRetry true after reset")
	}
}

func TestRetryTracker_IndependentPerLease(t *testing.T) {
	policy := RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Second, Multiplier: 2.0}
	rt := NewRetryTracker(policy)
	rt.NextDelay("lease-1")
	rt.NextDelay("lease-1")
	if !rt.ShouldRetry("lease-2") {
		t.Error("expected lease-2 to be unaffected by lease-1 exhaustion")
	}
}

func TestRetryTracker_Attempts_Counts(t *testing.T) {
	rt := NewRetryTracker(DefaultRetryPolicy())
	rt.NextDelay("lease-1")
	rt.NextDelay("lease-1")
	if rt.Attempts("lease-1") != 2 {
		t.Errorf("expected 2 attempts, got %d", rt.Attempts("lease-1"))
	}
}
