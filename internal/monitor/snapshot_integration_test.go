package monitor_test

import (
	"strings"
	"bytes"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// TestSnapshotStoreAndReporter_Integration verifies that the SnapshotStore
// and StatusReporter work together end-to-end.
func TestSnapshotStoreAndReporter_Integration(t *testing.T) {
	store := monitor.NewSnapshotStore()
	reporter := monitor.NewStatusReporter(store)

	// Before any data: summary should show all zeros.
	summary := reporter.Summary()
	if !strings.Contains(summary, "critical=0") {
		t.Errorf("expected critical=0 before set, got: %s", summary)
	}

	// Populate store.
	leases := []monitor.LeaseStatus{
		monitor.NewLeaseStatus("auth/token/create/abc", 4*time.Minute, true),
		monitor.NewLeaseStatus("database/creds/readonly", 25*time.Minute, true),
		monitor.NewLeaseStatus("pki/issue/web", 6*time.Hour, false),
	}
	store.Set(leases)

	// Summary should reflect new data.
	summary = reporter.Summary()
	if !strings.Contains(summary, "critical=1") {
		t.Errorf("expected critical=1, got: %s", summary)
	}
	if !strings.Contains(summary, "warning=1") {
		t.Errorf("expected warning=1, got: %s", summary)
	}
	if !strings.Contains(summary, "info=1") {
		t.Errorf("expected info=1, got: %s", summary)
	}

	// Written report should include all lease IDs.
	var buf bytes.Buffer
	if err := reporter.Write(&buf); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	out := buf.String()
	for _, id := range []string{"auth/token/create/abc", "database/creds/readonly", "pki/issue/web"} {
		if !strings.Contains(out, id) {
			t.Errorf("expected lease ID %q in report output", id)
		}
	}
}
