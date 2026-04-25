package monitor

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestStatusReporter_WriteNoSnapshot(t *testing.T) {
	store := NewSnapshotStore()
	reporter := NewStatusReporter(store)

	var buf bytes.Buffer
	if err := reporter.Write(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No snapshot available") {
		t.Errorf("expected no-snapshot message, got: %s", buf.String())
	}
}

func TestStatusReporter_WriteWithLeases(t *testing.T) {
	store := NewSnapshotStore()
	store.Set([]LeaseStatus{
		NewLeaseStatus("secret/data/db", 5*time.Minute, true),
		NewLeaseStatus("secret/data/api", 2*time.Hour, false),
	})

	reporter := NewStatusReporter(store)
	var buf bytes.Buffer
	if err := reporter.Write(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "secret/data/db") {
		t.Error("expected lease ID in output")
	}
	if !strings.Contains(out, "secret/data/api") {
		t.Error("expected second lease ID in output")
	}
	if !strings.Contains(out, "LEASE ID") {
		t.Error("expected table header in output")
	}
	if !strings.Contains(out, "Total leases: 2") {
		t.Error("expected total count in output")
	}
}

func TestStatusReporter_WriteShowsRenewable(t *testing.T) {
	store := NewSnapshotStore()
	store.Set([]LeaseStatus{
		NewLeaseStatus("secret/renewable", 30*time.Minute, true),
		NewLeaseStatus("secret/not-renewable", 30*time.Minute, false),
	})

	reporter := NewStatusReporter(store)
	var buf bytes.Buffer
	_ = reporter.Write(&buf)
	out := buf.String()

	if !strings.Contains(out, "yes") {
		t.Error("expected 'yes' for renewable lease")
	}
	if !strings.Contains(out, "no") {
		t.Error("expected 'no' for non-renewable lease")
	}
}

func TestStatusReporter_Summary(t *testing.T) {
	store := NewSnapshotStore()
	store.Set([]LeaseStatus{
		NewLeaseStatus("a", 3*time.Minute, true),
		NewLeaseStatus("b", 20*time.Minute, true),
		NewLeaseStatus("c", 3*time.Hour, true),
	})

	reporter := NewStatusReporter(store)
	summary := reporter.Summary()

	if !strings.Contains(summary, "critical=1") {
		t.Errorf("expected critical=1 in summary: %s", summary)
	}
	if !strings.Contains(summary, "warning=1") {
		t.Errorf("expected warning=1 in summary: %s", summary)
	}
	if !strings.Contains(summary, "info=1") {
		t.Errorf("expected info=1 in summary: %s", summary)
	}
}
