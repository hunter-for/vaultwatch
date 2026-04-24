// Package monitor provides lease monitoring and expiration alerting.
package monitor

import (
	"context"
	"log"
	"time"

	"github.com/vaultwatch/internal/vault"
)

// LeaseAlert represents an expiring lease that requires attention.
type LeaseAlert struct {
	LeaseID   string
	ExpiresAt time.Time
	TTL       time.Duration
}

// Monitor watches Vault leases and emits alerts before expiration.
type Monitor struct {
	client    *vault.Client
	threshold time.Duration
	interval  time.Duration
	alerts    chan LeaseAlert
}

// New creates a Monitor with the given Vault client and configuration.
func New(client *vault.Client, threshold, interval time.Duration) *Monitor {
	return &Monitor{
		client:    client,
		threshold: threshold,
		interval:  interval,
		alerts:    make(chan LeaseAlert, 32),
	}
}

// Alerts returns the read-only channel of lease alerts.
func (m *Monitor) Alerts() <-chan LeaseAlert {
	return m.alerts
}

// Run starts the monitoring loop, blocking until ctx is cancelled.
func (m *Monitor) Run(ctx context.Context, leaseIDs []string) error {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			m.checkLeases(ctx, leaseIDs)
		}
	}
}

func (m *Monitor) checkLeases(ctx context.Context, leaseIDs []string) {
	for _, id := range leaseIDs {
		lease, err := m.client.LookupLease(ctx, id)
		if err != nil {
			log.Printf("monitor: failed to lookup lease %q: %v", id, err)
			continue
		}

		ttl := time.Duration(lease.Data["ttl"].(float64)) * time.Second
		if ttl <= m.threshold {
			m.alerts <- LeaseAlert{
				LeaseID:   id,
				ExpiresAt: time.Now().Add(ttl),
				TTL:       ttl,
			}
		}
	}
}
