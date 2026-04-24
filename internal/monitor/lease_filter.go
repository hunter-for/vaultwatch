package monitor

import "time"

// LeaseFilter defines criteria for selecting leases to alert on.
type LeaseFilter struct {
	// OnlyRenewable, when true, skips non-renewable leases.
	OnlyRenewable bool
	// MinTTL skips leases whose remaining TTL exceeds this threshold.
	// A zero value means no upper bound filter is applied.
	MinTTL time.Duration
}

// Matches returns true if the given LeaseStatus satisfies the filter criteria.
func (f LeaseFilter) Matches(lease LeaseStatus) bool {
	if f.OnlyRenewable && !lease.Renewable {
		return false
	}
	if f.MinTTL > 0 && lease.TimeRemaining() > f.MinTTL {
		return false
	}
	return true
}

// Apply returns only the leases from the input slice that match the filter.
func (f LeaseFilter) Apply(leases []LeaseStatus) []LeaseStatus {
	result := make([]LeaseStatus, 0, len(leases))
	for _, l := range leases {
		if f.Matches(l) {
			result = append(result, l)
		}
	}
	return result
}

// DefaultLeaseFilter returns a filter that matches all leases within the
// warning threshold window, regardless of renewability.
func DefaultLeaseFilter(thresholds Thresholds) LeaseFilter {
	return LeaseFilter{
		OnlyRenewable: false,
		MinTTL:        thresholds.WarningThreshold,
	}
}
