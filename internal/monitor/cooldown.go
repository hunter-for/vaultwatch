package monitor

import (
	"sync"
	"time"
)

// CooldownPolicy defines per-severity cooldown durations.
type CooldownPolicy struct {
	Critical time.Duration
	Warning  time.Duration
	Info     time.Duration
}

// DefaultCooldownPolicy returns sensible defaults for alert cooldowns.
func DefaultCooldownPolicy() CooldownPolicy {
	return CooldownPolicy{
		Critical: 5 * time.Minute,
		Warning:  15 * time.Minute,
		Info:     30 * time.Minute,
	}
}

// CooldownTracker tracks the last alert time per lease ID and severity.
type CooldownTracker struct {
	mu     sync.Mutex
	policy CooldownPolicy
	last   map[string]time.Time
}

// NewCooldownTracker creates a CooldownTracker with the given policy.
func NewCooldownTracker(policy CooldownPolicy) *CooldownTracker {
	return &CooldownTracker{
		policy: policy,
		last:   make(map[string]time.Time),
	}
}

// key builds a unique key from lease ID and severity.
func (c *CooldownTracker) key(leaseID string, sev Severity) string {
	return leaseID + ":" + sev.String()
}

// durationFor returns the cooldown duration for a given severity.
func (c *CooldownTracker) durationFor(sev Severity) time.Duration {
	switch sev {
	case SeverityCritical:
		return c.policy.Critical
	case SeverityWarning:
		return c.policy.Warning
	default:
		return c.policy.Info
	}
}

// Allow returns true if enough time has passed since the last alert
// for the given lease ID and severity combination.
func (c *CooldownTracker) Allow(leaseID string, sev Severity) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	k := c.key(leaseID, sev)
	lastSent, seen := c.last[k]
	if !seen {
		return true
	}
	return time.Since(lastSent) >= c.durationFor(sev)
}

// Record marks the current time as the last alert sent for the given
// lease ID and severity.
func (c *CooldownTracker) Record(leaseID string, sev Severity) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.last[c.key(leaseID, sev)] = time.Now()
}

// Reset clears the cooldown state for a given lease ID across all severities.
func (c *CooldownTracker) Reset(leaseID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, sev := range []Severity{SeverityCritical, SeverityWarning, SeverityInfo} {
		delete(c.last, c.key(leaseID, sev))
	}
}
