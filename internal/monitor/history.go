package monitor

import (
	"sync"
	"time"
)

// AlertEvent records a single alert that was sent for a lease.
type AlertEvent struct {
	LeaseID   string
	Severity  Severity
	SentAt    time.Time
	TTL       time.Duration
}

// AlertHistory maintains an in-memory ring buffer of recent alert events.
type AlertHistory struct {
	mu     sync.RWMutex
	events []AlertEvent
	maxLen int
}

// NewAlertHistory creates an AlertHistory that retains up to maxLen events.
func NewAlertHistory(maxLen int) *AlertHistory {
	if maxLen <= 0 {
		maxLen = 100
	}
	return &AlertHistory{
		events: make([]AlertEvent, 0, maxLen),
		maxLen: maxLen,
	}
}

// Record appends an alert event, evicting the oldest when at capacity.
func (h *AlertHistory) Record(e AlertEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.events) >= h.maxLen {
		h.events = h.events[1:]
	}
	h.events = append(h.events, e)
}

// Recent returns up to n most recent events, newest first.
func (h *AlertHistory) Recent(n int) []AlertEvent {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := len(h.events)
	if n <= 0 || n > total {
		n = total
	}
	result := make([]AlertEvent, n)
	for i := 0; i < n; i++ {
		result[i] = h.events[total-1-i]
	}
	return result
}

// CountBySeverity returns a map of severity → event count across all stored events.
func (h *AlertHistory) CountBySeverity() map[Severity]int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	counts := make(map[Severity]int)
	for _, e := range h.events {
		counts[e.Severity]++
	}
	return counts
}

// Len returns the number of stored events.
func (h *AlertHistory) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.events)
}
