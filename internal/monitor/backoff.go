package monitor

import (
	"math"
	"time"
)

// BackoffPolicy defines how retry delays are calculated for failed alert sends.
type BackoffPolicy struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultBackoffPolicy returns a sensible exponential backoff configuration.
func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{
		InitialDelay: 5 * time.Second,
		MaxDelay:     5 * time.Minute,
		Multiplier:   2.0,
	}
}

// BackoffTracker tracks retry attempts per lease and computes next delay.
type BackoffTracker struct {
	policy  BackoffPolicy
	attempts map[string]int
}

// NewBackoffTracker creates a BackoffTracker using the given policy.
func NewBackoffTracker(policy BackoffPolicy) *BackoffTracker {
	return &BackoffTracker{
		policy:   policy,
		attempts: make(map[string]int),
	}
}

// NextDelay returns the delay before the next retry for the given lease ID
// and increments the attempt counter.
func (b *BackoffTracker) NextDelay(leaseID string) time.Duration {
	attempt := b.attempts[leaseID]
	b.attempts[leaseID]++

	delay := float64(b.policy.InitialDelay) * math.Pow(b.policy.Multiplier, float64(attempt))
	if delay > float64(b.policy.MaxDelay) {
		delay = float64(b.policy.MaxDelay)
	}
	return time.Duration(delay)
}

// Attempts returns the number of retries recorded for the given lease ID.
func (b *BackoffTracker) Attempts(leaseID string) int {
	return b.attempts[leaseID]
}

// Reset clears the retry state for the given lease ID.
func (b *BackoffTracker) Reset(leaseID string) {
	delete(b.attempts, leaseID)
}
