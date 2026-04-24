package monitor

import (
	"context"
	"fmt"
	"log"
	"time"
)

// LeaseRenewer attempts to renew Vault leases before they expire.
type LeaseRenewer struct {
	client  LeaseRenewClient
	leadway time.Duration
}

// LeaseRenewClient describes the Vault client methods needed for renewal.
type LeaseRenewClient interface {
	RenewLease(ctx context.Context, leaseID string, increment int) error
}

// NewLeaseRenewer creates a LeaseRenewer with the given client and leadway.
// leadway is how far before expiry to attempt renewal.
func NewLeaseRenewer(client LeaseRenewClient, leadway time.Duration) *LeaseRenewer {
	if leadway <= 0 {
		leadway = 30 * time.Second
	}
	return &LeaseRenewer{
		client:  client,
		leadway: leadway,
	}
}

// ShouldRenew reports whether a lease should be renewed given its current status.
func (r *LeaseRenewer) ShouldRenew(ls LeaseStatus) bool {
	if !ls.Renewable {
		return false
	}
	return ls.TimeRemaining() <= r.leadway && !ls.IsExpired()
}

// Renew attempts to renew the given lease and returns an error on failure.
func (r *LeaseRenewer) Renew(ctx context.Context, ls LeaseStatus, incrementSeconds int) error {
	if !r.ShouldRenew(ls) {
		return nil
	}
	log.Printf("[renewer] renewing lease %s (TTL remaining: %s)", ls.LeaseID, formatDuration(ls.TimeRemaining()))
	if err := r.client.RenewLease(ctx, ls.LeaseID, incrementSeconds); err != nil {
		return fmt.Errorf("renew lease %s: %w", ls.LeaseID, err)
	}
	return nil
}

// RenewAll attempts renewal for all leases in the slice, collecting errors.
func (r *LeaseRenewer) RenewAll(ctx context.Context, leases []LeaseStatus, incrementSeconds int) []error {
	var errs []error
	for _, ls := range leases {
		if err := r.Renew(ctx, ls, incrementSeconds); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
