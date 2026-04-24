package monitor

import (
	"testing"
	"time"
)

func makeLeases() []LeaseStatus {
	return []LeaseStatus{
		NewLeaseStatus("lease/1", "secret/a", 5*time.Minute, true),
		NewLeaseStatus("lease/2", "secret/b", 30*time.Minute, false),
		NewLeaseStatus("lease/3", "secret/c", 2*time.Hour, true),
	}
}

func TestLeaseFilter_Matches_OnlyRenewable(t *testing.T) {
	f := LeaseFilter{OnlyRenewable: true}
	nonRenewable := NewLeaseStatus("lease/x", "secret/x", 10*time.Minute, false)
	if f.Matches(nonRenewable) {
		t.Error("expected non-renewable lease to not match OnlyRenewable filter")
	}
	renewable := NewLeaseStatus("lease/y", "secret/y", 10*time.Minute, true)
	if !f.Matches(renewable) {
		t.Error("expected renewable lease to match OnlyRenewable filter")
	}
}

func TestLeaseFilter_Matches_MinTTL(t *testing.T) {
	f := LeaseFilter{MinTTL: 15 * time.Minute}
	short := NewLeaseStatus("lease/short", "secret/s", 5*time.Minute, false)
	if !f.Matches(short) {
		t.Error("expected short-TTL lease to match MinTTL filter")
	}
	long := NewLeaseStatus("lease/long", "secret/l", 1*time.Hour, false)
	if f.Matches(long) {
		t.Error("expected long-TTL lease to not match MinTTL filter")
	}
}

func TestLeaseFilter_Apply_FiltersCorrectly(t *testing.T) {
	f := LeaseFilter{MinTTL: 1 * time.Hour}
	leases := makeLeases()
	result := f.Apply(leases)
	if len(result) != 2 {
		t.Errorf("expected 2 leases after filter, got %d", len(result))
	}
}

func TestLeaseFilter_Apply_EmptyInput(t *testing.T) {
	f := LeaseFilter{OnlyRenewable: true}
	result := f.Apply([]LeaseStatus{})
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestDefaultLeaseFilter_UsesWarningThreshold(t *testing.T) {
	thresholds := DefaultThresholds()
	f := DefaultLeaseFilter(thresholds)
	if f.MinTTL != thresholds.WarningThreshold {
		t.Errorf("expected MinTTL %s, got %s", thresholds.WarningThreshold, f.MinTTL)
	}
	if f.OnlyRenewable {
		t.Error("expected OnlyRenewable to be false for default filter")
	}
}
