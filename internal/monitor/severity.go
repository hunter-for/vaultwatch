package monitor

import "time"

// Severity represents the urgency level of a lease expiration alert.
type Severity int

const (
	// SeverityInfo indicates the lease has ample time remaining.
	SeverityInfo Severity = iota
	// SeverityWarning indicates the lease is approaching expiration.
	SeverityWarning
	// SeverityCritical indicates the lease is critically close to expiration.
	SeverityCritical
)

// String returns the human-readable name of the severity level.
func (s Severity) String() string {
	switch s {
	case SeverityCritical:
		return "CRITICAL"
	case SeverityWarning:
		return "WARNING"
	default:
		return "INFO"
	}
}

// Alert holds all information about a lease expiration event.
type Alert struct {
	LeaseID   string
	TTL       time.Duration
	Severity  Severity
	Timestamp time.Time
}

// Thresholds defines the TTL boundaries used to classify alert severity.
type Thresholds struct {
	Critical time.Duration
	Warning  time.Duration
}

// DefaultThresholds returns the standard thresholds used by the monitor.
func DefaultThresholds() Thresholds {
	return Thresholds{
		Critical: 1 * time.Hour,
		Warning:  6 * time.Hour,
	}
}

// Classify returns the appropriate Severity for a given TTL and thresholds.
func Classify(ttl time.Duration, t Thresholds) Severity {
	switch {
	case ttl <= t.Critical:
		return SeverityCritical
	case ttl <= t.Warning:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}
