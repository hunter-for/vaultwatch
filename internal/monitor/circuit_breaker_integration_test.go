package monitor

import (
	"errors"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/alert"
)

type countingSender struct {
	calls  int
	failOn int // fail when calls == failOn (0 = never)
}

func (c *countingSender) Send(_ alert.Alert) error {
	c.calls++
	if c.failOn > 0 && c.calls <= c.failOn {
		return errors.New("sender error")
	}
	return nil
}

func cbAlert(leaseID string) alert.Alert {
	return alert.Alert{
		LeaseID:  leaseID,
		Severity: alert.SeverityWarning,
		Message:  "test alert",
	}
}

func TestCircuitBreakerSender_PassesThroughOnSuccess(t *testing.T) {
	inner := &countingSender{}
	s := NewCircuitBreakerSender(inner, DefaultCircuitBreakerPolicy())

	if err := s.Send(cbAlert("lease-1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
	if s.State("lease-1") != CircuitClosed {
		t.Fatalf("expected closed, got %s", s.State("lease-1"))
	}
}

func TestCircuitBreakerSender_OpensAfterFailures(t *testing.T) {
	policy := CircuitBreakerPolicy{MaxFailures: 3, ResetTimeout: 10 * time.Second}
	inner := &countingSender{failOn: 10} // always fail
	s := NewCircuitBreakerSender(inner, policy)

	for i := 0; i < 3; i++ {
		_ = s.Send(cbAlert("lease-2"))
	}

	if s.State("lease-2") != CircuitOpen {
		t.Fatalf("expected open, got %s", s.State("lease-2"))
	}

	err := s.Send(cbAlert("lease-2"))
	if err == nil {
		t.Fatal("expected ErrCircuitOpen")
	}
	var cbErr *ErrCircuitOpen
	if !errors.As(err, &cbErr) {
		t.Fatalf("expected ErrCircuitOpen, got %T: %v", err, err)
	}
}

func TestCircuitBreakerSender_RecoversThroughHalfOpen(t *testing.T) {
	policy := CircuitBreakerPolicy{MaxFailures: 2, ResetTimeout: 20 * time.Millisecond}
	inner := &countingSender{failOn: 2} // first 2 calls fail, then succeed
	s := NewCircuitBreakerSender(inner, policy)

	_ = s.Send(cbAlert("lease-3"))
	_ = s.Send(cbAlert("lease-3"))

	time.Sleep(30 * time.Millisecond)

	if err := s.Send(cbAlert("lease-3")); err != nil {
		t.Fatalf("expected success in half-open: %v", err)
	}
	if s.State("lease-3") != CircuitClosed {
		t.Fatalf("expected closed after recovery, got %s", s.State("lease-3"))
	}
}

func TestCircuitBreakerSender_WithClassify(t *testing.T) {
	policy := CircuitBreakerPolicy{MaxFailures: 2, ResetTimeout: 10 * time.Second}
	inner := &countingSender{failOn: 10}
	s := NewCircuitBreakerSender(inner, policy)

	lease := renewableLeaseWithTTL(5 * time.Minute)
	sev := Classify(lease, DefaultThresholds())

	a := alert.Alert{
		LeaseID:  lease.LeaseID,
		Severity: alert.Severity(sev.String()),
		Message:  "classified alert",
	}

	_ = s.Send(a)
	_ = s.Send(a)

	if s.State(lease.LeaseID) != CircuitOpen {
		t.Fatalf("expected circuit open after failures, got %s", s.State(lease.LeaseID))
	}
}
