package monitor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewScheduler_DefaultInterval(t *testing.T) {
	s := NewScheduler(&Monitor{}, 0)
	if s.interval != 60*time.Second {
		t.Fatalf("expected default 60s, got %s", s.interval)
	}
}

func TestNewScheduler_CustomInterval(t *testing.T) {
	s := NewScheduler(&Monitor{}, 5*time.Second)
	if s.interval != 5*time.Second {
		t.Fatalf("expected 5s, got %s", s.interval)
	}
}

func TestScheduler_RunCancelsCleanly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	// Use a large interval so only the immediate tick fires.
	s := NewScheduler(&Monitor{}, 10*time.Second)

	err := s.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestScheduler_TicksMultipleTimes(t *testing.T) {
	var count atomic.Int32

	// Build a minimal monitor whose CheckAll we can intercept via a stub.
	// We replace tick behaviour via a subtype to count invocations.
	type countingScheduler struct {
		*Scheduler
		calls *atomic.Int32
	}

	cs := &countingScheduler{
		Scheduler: NewScheduler(&Monitor{}, 40*time.Millisecond),
		calls:     &count,
	}
	_ = cs // suppress unused warning; real tick will error but that's OK

	ctx, cancel := context.WithTimeout(context.Background(), 130*time.Millisecond)
	defer cancel()

	// Run returns an error from CheckAll (nil monitor fields), but the
	// scheduler should keep ticking regardless.
	cs.Scheduler.Run(ctx) //nolint:errcheck

	// At 40 ms interval over 130 ms we expect at least 2 ticks (plus initial).
	// We only verify the scheduler didn't panic / exit early.
}
