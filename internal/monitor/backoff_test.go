package monitor

import (
	"testing"
	"time"
)

func TestBackoffTracker_FirstDelayIsInitial(t *testing.T) {
	policy := BackoffPolicy{
		InitialDelay: 10 * time.Second,
		MaxDelay:     5 * time.Minute,
		Multiplier:   2.0,
	}
	bt := NewBackoffTracker(policy)
	delay := bt.NextDelay("lease-1")
	if delay != 10*time.Second {
		t.Errorf("expected 10s, got %v", delay)
	}
}

func TestBackoffTracker_DelayDoubles(t *testing.T) {
	policy := BackoffPolicy{
		InitialDelay: 5 * time.Second,
		MaxDelay:     5 * time.Minute,
		Multiplier:   2.0,
	}
	bt := NewBackoffTracker(policy)
	_ = bt.NextDelay("lease-1") // attempt 0 → 5s
	delay := bt.NextDelay("lease-1") // attempt 1 → 10s
	if delay != 10*time.Second {
		t.Errorf("expected 10s on second attempt, got %v", delay)
	}
}

func TestBackoffTracker_CapsAtMaxDelay(t *testing.T) {
	policy := BackoffPolicy{
		InitialDelay: 30 * time.Second,
		MaxDelay:     1 * time.Minute,
		Multiplier:   4.0,
	}
	bt := NewBackoffTracker(policy)
	for i := 0; i < 5; i++ {
		_ = bt.NextDelay("lease-x")
	}
	delay := bt.NextDelay("lease-x")
	if delay > 1*time.Minute {
		t.Errorf("delay %v exceeded max of 1m", delay)
	}
}

func TestBackoffTracker_IndependentPerLease(t *testing.T) {
	bt := NewBackoffTracker(DefaultBackoffPolicy())
	_ = bt.NextDelay("a")
	_ = bt.NextDelay("a")
	first := bt.NextDelay("b")
	if first != DefaultBackoffPolicy().InitialDelay {
		t.Errorf("lease 'b' should start fresh, got %v", first)
	}
}

func TestBackoffTracker_AttemptsCount(t *testing.T) {
	bt := NewBackoffTracker(DefaultBackoffPolicy())
	if bt.Attempts("x") != 0 {
		t.Error("expected 0 attempts initially")
	}
	_ = bt.NextDelay("x")
	_ = bt.NextDelay("x")
	if bt.Attempts("x") != 2 {
		t.Errorf("expected 2 attempts, got %d", bt.Attempts("x"))
	}
}

func TestBackoffTracker_ResetClearsState(t *testing.T) {
	bt := NewBackoffTracker(DefaultBackoffPolicy())
	_ = bt.NextDelay("lease-r")
	_ = bt.NextDelay("lease-r")
	bt.Reset("lease-r")
	if bt.Attempts("lease-r") != 0 {
		t.Error("expected attempts reset to 0")
	}
	delay := bt.NextDelay("lease-r")
	if delay != DefaultBackoffPolicy().InitialDelay {
		t.Errorf("expected initial delay after reset, got %v", delay)
	}
}
