package monitor

import (
	"fmt"
	"strings"
	"time"
)

// Severity indicates the urgency level of a lease alert.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityWarning  Severity = "WARNING"
	SeverityInfo     Severity = "INFO"
)

// Format returns a human-readable string for the given alert.
func Format(a LeaseAlert) string {
	sev := severityFor(a.TTL)
	return fmt.Sprintf("[%s] Lease %q expires in %s (at %s)",
		sev,
		a.LeaseID,
		formatDuration(a.TTL),
		a.ExpiresAt.UTC().Format(time.RFC3339),
	)
}

// severityFor maps a remaining TTL to a severity level.
func severityFor(ttl time.Duration) Severity {
	switch {
	case ttl <= 5*time.Minute:
		return SeverityCritical
	case ttl <= 30*time.Minute:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// formatDuration returns a concise human-readable duration string.
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	parts := []string{}
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}
	if m > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
	}
	if s > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", s))
	}
	return strings.Join(parts, " ")
}
