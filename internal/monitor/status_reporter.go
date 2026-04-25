package monitor

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// StatusReporter writes a human-readable status report to an io.Writer.
type StatusReporter struct {
	store *SnapshotStore
}

// NewStatusReporter creates a StatusReporter backed by the given SnapshotStore.
func NewStatusReporter(store *SnapshotStore) *StatusReporter {
	return &StatusReporter{store: store}
}

// Write outputs the current snapshot as a formatted table.
func (r *StatusReporter) Write(w io.Writer) error {
	snap := r.store.Get()
	if snap == nil {
		_, err := fmt.Fprintln(w, "No snapshot available yet.")
		return err
	}

	fmt.Fprintf(w, "Snapshot captured at: %s\n", snap.CapturedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Total leases: %d\n\n", len(snap.Leases))

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "LEASE ID\tTTL\tRENEWABLE\tSEVERITY")

	thresholds := DefaultThresholds()
	for _, lease := range snap.Leases {
		sev := Classify(lease.TTL, thresholds)
		renewable := "no"
		if lease.Renewable {
			renewable = "yes"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			lease.LeaseID,
			formatDuration(lease.TTL),
			renewable,
			sev.String(),
		)
	}
	return tw.Flush()
}

// Summary returns a brief one-line status string.
func (r *StatusReporter) Summary() string {
	counts := r.store.Summary()
	return fmt.Sprintf("critical=%d warning=%d info=%d",
		counts["critical"], counts["warning"], counts["info"])
}
