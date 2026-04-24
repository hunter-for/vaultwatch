package alert_test

import (
	"errors"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/alert"
	"github.com/user/vaultwatch/internal/monitor"
)

type stubSender struct {
	called bool
	err    error
}

func (s *stubSender) Send(_ monitor.Alert) error {
	s.called = true
	return s.err
}

func baseMultiAlert() monitor.Alert {
	return monitor.Alert{
		LeaseID:   "secret/data/test#xyz",
		TTL:       30 * time.Minute,
		Severity:  monitor.SeverityCritical,
		Message:   "critical expiry",
		Timestamp: time.Now().UTC(),
	}
}

func TestMultiSender_CallsAllSenders(t *testing.T) {
	a := &stubSender{}
	b := &stubSender{}

	ms := alert.NewMultiSender(a, b)
	if err := ms.Send(baseMultiAlert()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !a.called {
		t.Error("expected first sender to be called")
	}
	if !b.called {
		t.Error("expected second sender to be called")
	}
}

func TestMultiSender_ReturnsErrorsFromAll(t *testing.T) {
	errA := errors.New("sender A failed")
	errB := errors.New("sender B failed")

	a := &stubSender{err: errA}
	b := &stubSender{err: errB}

	ms := alert.NewMultiSender(a, b)
	err := ms.Send(baseMultiAlert())
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}

	if !errors.Is(err, errA) {
		t.Errorf("expected error to contain errA")
	}
	if !errors.Is(err, errB) {
		t.Errorf("expected error to contain errB")
	}
}

func TestMultiSender_ContinuesAfterFailure(t *testing.T) {
	a := &stubSender{err: errors.New("fail")}
	b := &stubSender{}

	ms := alert.NewMultiSender(a, b)
	_ = ms.Send(baseMultiAlert())

	if !b.called {
		t.Error("expected second sender to be called even after first failure")
	}
}
