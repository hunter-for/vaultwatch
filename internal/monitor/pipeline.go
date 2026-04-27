package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/vault"
)

// Pipeline coordinates the full lease monitoring cycle: list leases, filter,
// classify, deduplicate, rate-limit, and dispatch alerts. It is intended to be
// called on every scheduler tick.
type Pipeline struct {
	client    *vault.Client
	filter    *LeaseFilter
	dedup     *DedupStore
	rateLim   *RateLimiter
	snapshot  *SnapshotStore
	history   *AlertHistory
	sender    alert.Sender
	thresh    Thresholds
}

// PipelineConfig holds all dependencies required to construct a Pipeline.
type PipelineConfig struct {
	Client   *vault.Client
	Filter   *LeaseFilter
	Dedup    *DedupStore
	RateLim  *RateLimiter
	Snapshot *SnapshotStore
	History  *AlertHistory
	Sender   alert.Sender
	Thresh   Thresholds
}

// NewPipeline constructs a Pipeline from the provided configuration.
func NewPipeline(cfg PipelineConfig) *Pipeline {
	return &Pipeline{
		client:   cfg.Client,
		filter:   cfg.Filter,
		dedup:    cfg.Dedup,
		rateLim:  cfg.RateLim,
		snapshot: cfg.Snapshot,
		history:  cfg.History,
		sender:   cfg.Sender,
		thresh:   cfg.Thresh,
	}
}

// Run executes one full monitoring cycle. It lists all leases, applies filters,
// classifies severity, suppresses duplicates and rate-limited alerts, then
// sends any qualifying alerts. Errors from individual leases are logged but do
// not abort the cycle.
func (p *Pipeline) Run(ctx context.Context) error {
	leaseIDs, err := p.client.ListLeases(ctx)
	if err != nil {
		return fmt.Errorf("pipeline: list leases: %w", err)
	}

	var statuses []LeaseStatus
	for _, id := range leaseIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		lease, err := p.client.GetLease(ctx, id)
		if err != nil {
			log.Printf("pipeline: get lease %s: %v", id, err)
			continue
		}

		status := NewLeaseStatus(lease)
		statuses = append(statuses, status)
	}

	// Apply filter (renewable-only, min-TTL, etc.).
	filtered := p.filter.Apply(statuses)

	// Persist snapshot for status reporter.
	p.snapshot.Set(filtered)

	now := time.Now()

	for _, ls := range filtered {
		sev := Classify(ls.TTL, p.thresh)

		// Skip informational leases to reduce noise.
		if sev == SeverityInfo {
			continue
		}

		// Dedup: suppress if same severity was recently alerted.
		if !p.dedup.ShouldSend(ls.LeaseID, sev) {
			continue
		}

		// Rate limiter: cap alerts per lease per window.
		if !p.rateLim.Allow(ls.LeaseID) {
			continue
		}

		a := alert.Alert{
			LeaseID:   ls.LeaseID,
			TTL:       ls.TTL,
			Severity:  sev.String(),
			Timestamp: now,
		}

		if err := p.sender.Send(ctx, a); err != nil {
			log.Printf("pipeline: send alert for %s: %v", ls.LeaseID, err)
			continue
		}

		// Record successful dispatch in history and dedup store.
		p.history.Record(AlertEvent{
			LeaseID:   ls.LeaseID,
			Severity:  sev,
			Timestamp: now,
		})
		p.dedup.MarkSent(ls.LeaseID, sev)
	}

	return nil
}
