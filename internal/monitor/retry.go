package monitor

import (
	"errors"
	"time"
)

// RetryPolicy defines how retries are attempted for a failed alert send.
type RetryPolicy struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}
}

// RetryTracker tracks retry state for a single lease alert send.
type RetryTracker struct {
	policy   RetryPolicy
	attempts map[string]int
	delays   map[string]time.Duration
}

// NewRetryTracker creates a RetryTracker with the given policy.
func NewRetryTracker(policy RetryPolicy) *RetryTracker {
	return &RetryTracker{
		policy:   policy,
		attempts: make(map[string]int),
		delays:   make(map[string]time.Duration),
	}
}

// ShouldRetry returns true if another attempt is allowed for the given lease.
func (r *RetryTracker) ShouldRetry(leaseID string) bool {
	return r.attempts[leaseID] < r.policy.MaxAttempts
}

// NextDelay returns the delay before the next retry attempt and records the attempt.
func (r *RetryTracker) NextDelay(leaseID string) time.Duration {
	delay, ok := r.delays[leaseID]
	if !ok {
		delay = r.policy.InitialDelay
	} else {
		delay = time.Duration(float64(delay) * r.policy.Multiplier)
		if delay > r.policy.MaxDelay {
			delay = r.policy.MaxDelay
		}
	}
	r.delays[leaseID] = delay
	r.attempts[leaseID]++
	return delay
}

// Reset clears retry state for a lease after a successful send.
func (r *RetryTracker) Reset(leaseID string) {
	delete(r.attempts, leaseID)
	delete(r.delays, leaseID)
}

// Attempts returns the number of attempts recorded for a lease.
func (r *RetryTracker) Attempts(leaseID string) int {
	return r.attempts[leaseID]
}

// ErrMaxRetriesExceeded is returned when retries are exhausted.
var ErrMaxRetriesExceeded = errors.New("max retry attempts exceeded")
