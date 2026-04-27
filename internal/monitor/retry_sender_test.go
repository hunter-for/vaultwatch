package monitor

import (
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

type mockSender struct {
	callCount int
	failUntil int
	err       error
}

func (m *mockSender) Send(_ alert.Alert) error {
	m.callCount++
	if m.callCount <= m.failUntil {
		return m.err
	}
	return nil
}

func noSleep(_ time.Duration) {}

func baseRetryAlert() alert.Alert {
	return alert.Alert{
		LeaseID:  "lease-abc",
		Severity: alert.Critical,
		TTL:      30 * time.Second,
	}
}

func TestRetrySender_SucceedsFirstTry(t *testing.T) {
	inner := &mockSender{}
	rs := NewRetrySender(inner, DefaultRetryPolicy())
	rs.sleep = noSleep
	if err := rs.Send(baseRetryAlert()); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if inner.callCount != 1 {
		t.Errorf("expected 1 call, got %d", inner.callCount)
	}
}

func TestRetrySender_RetriesOnFailure(t *testing.T) {
	inner := &mockSender{failUntil: 2, err: errors.New("send error")}
	policy := RetryPolicy{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: time.Second, Multiplier: 2.0}
	rs := NewRetrySender(inner, policy)
	rs.sleep = noSleep
	if err := rs.Send(baseRetryAlert()); err != nil {
		t.Errorf("expected success after retries, got %v", err)
	}
	if inner.callCount != 3 {
		t.Errorf("expected 3 calls, got %d", inner.callCount)
	}
}

func TestRetrySender_ReturnsErrAfterExhaustion(t *testing.T) {
	inner := &mockSender{failUntil: 10, err: errors.New("persistent error")}
	policy := RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Second, Multiplier: 2.0}
	rs := NewRetrySender(inner, policy)
	rs.sleep = noSleep
	err := rs.Send(baseRetryAlert())
	if err == nil {
		t.Fatal("expected error after exhaustion")
	}
	if !errors.Is(err, ErrMaxRetriesExceeded) {
		t.Errorf("expected ErrMaxRetriesExceeded, got %v", err)
	}
}

func TestRetrySender_ResetsAfterSuccess(t *testing.T) {
	inner := &mockSender{failUntil: 1, err: errors.New("transient")}
	policy := RetryPolicy{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: time.Second, Multiplier: 2.0}
	rs := NewRetrySender(inner, policy)
	rs.sleep = noSleep
	_ = rs.Send(baseRetryAlert())
	if rs.tracker.Attempts("lease-abc") != 0 {
		t.Error("expected tracker to reset after success")
	}
}
