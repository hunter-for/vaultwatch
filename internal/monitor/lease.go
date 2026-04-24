package monitor

import (
	"fmt"
	"time"
)

// LeaseStatus represents the current state of a Vault lease.
type LeaseStatus struct {
	LeaseID   string
	Path      string
	TTL       time.Duration
	Renewable bool
	ExpireAt  time.Time
}

// IsExpired returns true if the lease has already expired.
func (l LeaseStatus) IsExpired() bool {
	return time.Now().After(l.ExpireAt)
}

// TimeRemaining returns the duration until the lease expires.
// Returns 0 if the lease is already expired.
func (l LeaseStatus) TimeRemaining() time.Duration {
	remaining := time.Until(l.ExpireAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// String returns a human-readable summary of the lease.
func (l LeaseStatus) String() string {
	return fmt.Sprintf("LeaseStatus{ID: %q, Path: %q, TTL: %s, Renewable: %v, ExpiresAt: %s}",
		l.LeaseID,
		l.Path,
		l.TTL,
		l.Renewable,
		l.ExpireAt.Format(time.RFC3339),
	)
}

// NewLeaseStatus constructs a LeaseStatus from a lease ID, path, TTL, and renewable flag.
func NewLeaseStatus(leaseID, path string, ttl time.Duration, renewable bool) LeaseStatus {
	return LeaseStatus{
		LeaseID:   leaseID,
		Path:      path,
		TTL:       ttl,
		Renewable: renewable,
		ExpireAt:  time.Now().Add(ttl),
	}
}
