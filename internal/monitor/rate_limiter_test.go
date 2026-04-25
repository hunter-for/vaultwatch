package monitor

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowsUpToBurst(t *testing.T) {
	rl := NewRateLimiter(3, 10*time.Minute)
	leaseID := "lease/test/1"

	for i := 0; i < 3; i++ {
		if !rl.Allow(leaseID) {
			t.Fatalf("expected Allow to return true on call %d", i+1)
		}
	}

	if rl.Allow(leaseID) {
		t.Error("expected Allow to return false after burst exhausted")
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)
	leaseID := "lease/test/reset"

	rl.Allow(leaseID)
	rl.Allow(leaseID)

	if rl.Allow(leaseID) {
		t.Error("expected rate limit to be exhausted before window reset")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow(leaseID) {
		t.Error("expected Allow to return true after window reset")
	}
}

func TestRateLimiter_IndependentPerLease(t *testing.T) {
	rl := NewRateLimiter(1, 10*time.Minute)

	if !rl.Allow("lease/a") {
		t.Error("expected first allow for lease/a")
	}
	if !rl.Allow("lease/b") {
		t.Error("expected first allow for lease/b")
	}
	if rl.Allow("lease/a") {
		t.Error("expected lease/a to be rate-limited")
	}
	if rl.Allow("lease/b") {
		t.Error("expected lease/b to be rate-limited")
	}
}

func TestRateLimiter_Remaining(t *testing.T) {
	rl := NewRateLimiter(3, 10*time.Minute)
	leaseID := "lease/remaining"

	if got := rl.Remaining(leaseID); got != 3 {
		t.Errorf("expected 3 remaining for unseen lease, got %d", got)
	}

	rl.Allow(leaseID)
	if got := rl.Remaining(leaseID); got != 2 {
		t.Errorf("expected 2 remaining after one allow, got %d", got)
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(1, 10*time.Minute)
	leaseID := "lease/reset-test"

	rl.Allow(leaseID)
	if rl.Allow(leaseID) {
		t.Error("expected rate limit to be exhausted")
	}

	rl.Reset(leaseID)
	if !rl.Allow(leaseID) {
		t.Error("expected Allow to succeed after Reset")
	}
}

func TestRateLimiter_DefaultsOnZeroValues(t *testing.T) {
	rl := NewRateLimiter(0, 0)
	leaseID := "lease/defaults"

	count := 0
	for i := 0; i < 10; i++ {
		if rl.Allow(leaseID) {
			count++
		}
	}
	if count != 3 {
		t.Errorf("expected default burst of 3, got %d", count)
	}
}
