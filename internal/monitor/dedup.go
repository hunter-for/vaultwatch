package monitor

import (
	"sync"
	"time"
)

// DedupStore tracks which lease alerts have already been sent to avoid
// flooding alert channels with repeated notifications for the same lease.
type DedupStore struct {
	mu      sync.Mutex
	entries map[string]dedupEntry
}

type dedupEntry struct {
	severity  Severity
	sentAt    time.Time
	cooldown  time.Duration
}

// NewDedupStore creates a new DedupStore ready for use.
func NewDedupStore() *DedupStore {
	return &DedupStore{
		entries: make(map[string]dedupEntry),
	}
}

// ShouldAlert returns true if an alert for the given lease ID and severity
// should be sent. It suppresses duplicate alerts within the cooldown window.
// If the severity has escalated, the alert is always allowed through.
func (d *DedupStore) ShouldAlert(leaseID string, sev Severity, cooldown time.Duration) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	entry, exists := d.entries[leaseID]
	if !exists {
		d.entries[leaseID] = dedupEntry{severity: sev, sentAt: now, cooldown: cooldown}
		return true
	}

	// Always alert on severity escalation.
	if sev > entry.severity {
		d.entries[leaseID] = dedupEntry{severity: sev, sentAt: now, cooldown: cooldown}
		return true
	}

	// Suppress if within cooldown window.
	if now.Before(entry.sentAt.Add(entry.cooldown)) {
		return false
	}

	d.entries[leaseID] = dedupEntry{severity: sev, sentAt: now, cooldown: cooldown}
	return true
}

// Evict removes a lease entry from the store, e.g. after a lease expires.
func (d *DedupStore) Evict(leaseID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, leaseID)
}

// Len returns the number of tracked lease entries.
func (d *DedupStore) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.entries)
}
