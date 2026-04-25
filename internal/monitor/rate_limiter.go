package monitor

import (
	"sync"
	"time"
)

// RateLimiter controls how frequently alerts can be sent per lease,
// regardless of severity changes. It complements the cooldown/dedup
// logic by enforcing a hard cap on alert throughput.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*rateBucket
	maxBurst int
	window   time.Duration
}

type rateBucket struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a RateLimiter that allows up to maxBurst alerts
// per leaseID within the given window duration.
func NewRateLimiter(maxBurst int, window time.Duration) *RateLimiter {
	if maxBurst <= 0 {
		maxBurst = 3
	}
	if window <= 0 {
		window = 10 * time.Minute
	}
	return &RateLimiter{
		buckets:  make(map[string]*rateBucket),
		maxBurst: maxBurst,
		window:   window,
	}
}

// Allow returns true if the alert for the given leaseID is within rate
// limits. It consumes one token from the bucket for that lease.
func (r *RateLimiter) Allow(leaseID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[leaseID]
	if !ok || now.Sub(b.lastReset) >= r.window {
		r.buckets[leaseID] = &rateBucket{
			tokens:    r.maxBurst - 1,
			lastReset: now,
		}
		return true
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

// Remaining returns the number of tokens left in the window for a lease.
func (r *RateLimiter) Remaining(leaseID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.buckets[leaseID]
	if !ok {
		return r.maxBurst
	}
	if time.Since(b.lastReset) >= r.window {
		return r.maxBurst
	}
	return b.tokens
}

// Reset clears the rate limit state for a specific leaseID.
func (r *RateLimiter) Reset(leaseID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.buckets, leaseID)
}
