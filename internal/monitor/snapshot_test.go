package monitor

import (
	"testing"
	"time"
)

func makeLeaseStatus(ttl time.Duration, renewable bool) LeaseStatus {
	return NewLeaseStatus("lease/test/"+ttl.String(), ttl, renewable)
}

func TestSnapshotStore_GetReturnsNilInitially(t *testing.T) {
	store := NewSnapshotStore()
	if store.Get() != nil {
		t.Fatal("expected nil snapshot before any Set call")
	}
}

func TestSnapshotStore_SetAndGet(t *testing.T) {
	store := NewSnapshotStore()
	leases := []LeaseStatus{
		makeLeaseStatus(10*time.Minute, true),
		makeLeaseStatus(2*time.Hour, true),
	}
	before := time.Now()
	store.Set(leases)
	snap := store.Get()

	if snap == nil {
		t.Fatal("expected non-nil snapshot after Set")
	}
	if len(snap.Leases) != 2 {
		t.Fatalf("expected 2 leases, got %d", len(snap.Leases))
	}
	if snap.CapturedAt.Before(before) {
		t.Error("CapturedAt should be after test start")
	}
}

func TestSnapshotStore_SetOverwritesPrevious(t *testing.T) {
	store := NewSnapshotStore()
	store.Set([]LeaseStatus{makeLeaseStatus(5*time.Minute, true)})
	store.Set([]LeaseStatus{})

	snap := store.Get()
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if len(snap.Leases) != 0 {
		t.Fatalf("expected 0 leases after overwrite, got %d", len(snap.Leases))
	}
}

func TestSnapshotStore_SummaryEmpty(t *testing.T) {
	store := NewSnapshotStore()
	summary := store.Summary()
	for _, k := range []string{"critical", "warning", "info"} {
		if summary[k] != 0 {
			t.Errorf("expected 0 for %s, got %d", k, summary[k])
		}
	}
}

func TestSnapshotStore_SummaryCountsBySeverity(t *testing.T) {
	store := NewSnapshotStore()
	leases := []LeaseStatus{
		makeLeaseStatus(3*time.Minute, true),  // critical
		makeLeaseStatus(20*time.Minute, true), // warning
		makeLeaseStatus(2*time.Hour, true),    // info
	}
	store.Set(leases)
	summary := store.Summary()

	if summary["critical"] != 1 {
		t.Errorf("expected 1 critical, got %d", summary["critical"])
	}
	if summary["warning"] != 1 {
		t.Errorf("expected 1 warning, got %d", summary["warning"])
	}
	if summary["info"] != 1 {
		t.Errorf("expected 1 info, got %d", summary["info"])
	}
}
