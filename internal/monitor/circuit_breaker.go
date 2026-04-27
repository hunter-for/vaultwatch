package monitor

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerPolicy configures thresholds for the circuit breaker.
type CircuitBreakerPolicy struct {
	MaxFailures  int
	ResetTimeout time.Duration
}

// DefaultCircuitBreakerPolicy returns sensible defaults.
func DefaultCircuitBreakerPolicy() CircuitBreakerPolicy {
	return CircuitBreakerPolicy{
		MaxFailures:  5,
		ResetTimeout: 30 * time.Second,
	}
}

// CircuitBreaker tracks consecutive failures per lease and opens the circuit
// when the failure threshold is exceeded.
type CircuitBreaker struct {
	mu       sync.Mutex
	policy   CircuitBreakerPolicy
	states   map[string]*cbEntry
}

type cbEntry struct {
	state      CircuitState
	failures   int
	openedAt   time.Time
}

// NewCircuitBreaker creates a CircuitBreaker with the given policy.
func NewCircuitBreaker(policy CircuitBreakerPolicy) *CircuitBreaker {
	return &CircuitBreaker{
		policy: policy,
		states: make(map[string]*cbEntry),
	}
}

// Allow returns true if the circuit is closed or half-open for the given leaseID.
func (cb *CircuitBreaker) Allow(leaseID string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	e := cb.entryFor(leaseID)
	switch e.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(e.openedAt) >= cb.policy.ResetTimeout {
			e.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return false
}

// RecordSuccess resets the circuit to closed for the given leaseID.
func (cb *CircuitBreaker) RecordSuccess(leaseID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e := cb.entryFor(leaseID)
	e.failures = 0
	e.state = CircuitClosed
}

// RecordFailure increments the failure count and may open the circuit.
func (cb *CircuitBreaker) RecordFailure(leaseID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e := cb.entryFor(leaseID)
	e.failures++
	if e.failures >= cb.policy.MaxFailures {
		e.state = CircuitOpen
		e.openedAt = time.Now()
	}
}

// State returns the current CircuitState for a leaseID.
func (cb *CircuitBreaker) State(leaseID string) CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.entryFor(leaseID).state
}

// Reset clears all state for a leaseID.
func (cb *CircuitBreaker) Reset(leaseID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	delete(cb.states, leaseID)
}

func (cb *CircuitBreaker) entryFor(leaseID string) *cbEntry {
	if _, ok := cb.states[leaseID]; !ok {
		cb.states[leaseID] = &cbEntry{state: CircuitClosed}
	}
	return cb.states[leaseID]
}

// ErrCircuitOpen is returned when a circuit is open.
type ErrCircuitOpen struct{ LeaseID string }

func (e *ErrCircuitOpen) Error() string {
	return fmt.Sprintf("circuit open for lease %s", e.LeaseID)
}
