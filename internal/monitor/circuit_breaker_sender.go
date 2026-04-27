package monitor

import (
	"fmt"

	"github.com/user/vaultwatch/internal/alert"
)

// CircuitBreakerSender wraps an alert.Sender and gates sends through a
// CircuitBreaker keyed on the alert's LeaseID.
type CircuitBreakerSender struct {
	inner   alert.Sender
	circuit *CircuitBreaker
}

// NewCircuitBreakerSender wraps inner with circuit-breaker protection.
func NewCircuitBreakerSender(inner alert.Sender, policy CircuitBreakerPolicy) *CircuitBreakerSender {
	return &CircuitBreakerSender{
		inner:   inner,
		circuit: NewCircuitBreaker(policy),
	}
}

// Send checks the circuit state before delegating to the inner sender.
// Failures are recorded on the circuit; successes reset it.
func (s *CircuitBreakerSender) Send(a alert.Alert) error {
	if !s.circuit.Allow(a.LeaseID) {
		return &ErrCircuitOpen{LeaseID: a.LeaseID}
	}

	if err := s.inner.Send(a); err != nil {
		s.circuit.RecordFailure(a.LeaseID)
		return fmt.Errorf("circuit breaker sender: %w", err)
	}

	s.circuit.RecordSuccess(a.LeaseID)
	return nil
}

// State exposes the circuit state for a given leaseID (useful for status reporting).
func (s *CircuitBreakerSender) State(leaseID string) CircuitState {
	return s.circuit.State(leaseID)
}
