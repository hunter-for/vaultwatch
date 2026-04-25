package monitor

import (
	"sync"
	"time"
)

// Snapshot holds a point-in-time view of all monitored lease statuses.
type Snapshot struct {
	CapturedAt time.Time
	Leases     []LeaseStatus
}

// SnapshotStore stores the most recent monitor snapshot in memory.
type SnapshotStore struct {
	mu       sync.RWMutex
	current  *Snapshot
}

// NewSnapshotStore creates an empty SnapshotStore.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{}
}

// Set replaces the current snapshot with a new one.
func (s *SnapshotStore) Set(leases []LeaseStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = &Snapshot{
		CapturedAt: time.Now(),
		Leases:     leases,
	}
}

// Get returns the current snapshot, or nil if none has been stored yet.
func (s *SnapshotStore) Get() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Summary returns counts of leases by severity in the latest snapshot.
func (s *SnapshotStore) Summary() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := map[string]int{
		"critical": 0,
		"warning":  0,
		"info":     0,
	}

	if s.current == nil {
		return counts
	}

	thresholds := DefaultThresholds()
	for _, lease := range s.current.Leases {
		sev := Classify(lease.TTL, thresholds)
		counts[sev.String()]++
	}
	return counts
}
