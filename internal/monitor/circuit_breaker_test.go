package monitor

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerPolicy())
	if !cb.Allow("lease-1") {
		t.Fatal("expected circuit to be closed initially")
	}
	if cb.State("lease-1") != CircuitClosed {
		t.Fatalf("expected closed, got %s", cb.State("lease-1"))
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	policy := CircuitBreakerPolicy{MaxFailures: 3, ResetTimeout: 10 * time.Second}
	cb := NewCircuitBreaker(policy)

	for i := 0; i < 3; i++ {
		cb.RecordFailure("lease-1")
	}

	if cb.State("lease-1") != CircuitOpen {
		t.Fatalf("expected open after max failures, got %s", cb.State("lease-1"))
	}
	if cb.Allow("lease-1") {
		t.Fatal("expected circuit to block when open")
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	policy := CircuitBreakerPolicy{MaxFailures: 2, ResetTimeout: 10 * time.Millisecond}
	cb := NewCircuitBreaker(policy)

	cb.RecordFailure("lease-2")
	cb.RecordFailure("lease-2")

	time.Sleep(20 * time.Millisecond)

	if !cb.Allow("lease-2") {
		t.Fatal("expected circuit to allow in half-open after timeout")
	}
	if cb.State("lease-2") != CircuitHalfOpen {
		t.Fatalf("expected half-open, got %s", cb.State("lease-2"))
	}
}

func TestCircuitBreaker_ClosesOnSuccess(t *testing.T) {
	policy := CircuitBreakerPolicy{MaxFailures: 2, ResetTimeout: 10 * time.Millisecond}
	cb := NewCircuitBreaker(policy)

	cb.RecordFailure("lease-3")
	cb.RecordFailure("lease-3")
	time.Sleep(20 * time.Millisecond)
	cb.Allow("lease-3") // transitions to half-open
	cb.RecordSuccess("lease-3")

	if cb.State("lease-3") != CircuitClosed {
		t.Fatalf("expected closed after success, got %s", cb.State("lease-3"))
	}
}

func TestCircuitBreaker_IndependentPerLease(t *testing.T) {
	policy := CircuitBreakerPolicy{MaxFailures: 2, ResetTimeout: 10 * time.Second}
	cb := NewCircuitBreaker(policy)

	cb.RecordFailure("lease-a")
	cb.RecordFailure("lease-a")

	if !cb.Allow("lease-b") {
		t.Fatal("lease-b should be unaffected by lease-a failures")
	}
}

func TestCircuitBreaker_ResetClearsState(t *testing.T) {
	policy := CircuitBreakerPolicy{MaxFailures: 1, ResetTimeout: 10 * time.Second}
	cb := NewCircuitBreaker(policy)

	cb.RecordFailure("lease-x")
	if cb.State("lease-x") != CircuitOpen {
		t.Fatal("expected open")
	}

	cb.Reset("lease-x")
	if cb.State("lease-x") != CircuitClosed {
		t.Fatalf("expected closed after reset, got %s", cb.State("lease-x"))
	}
}

func TestCircuitStateString(t *testing.T) {
	cases := []struct {
		state CircuitState
		want  string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}
	for _, tc := range cases {
		if tc.state.String() != tc.want {
			t.Errorf("expected %q, got %q", tc.want, tc.state.String())
		}
	}
}
