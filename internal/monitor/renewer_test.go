package monitor

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRenewClient records calls and optionally returns an error.
type mockRenewClient struct {
	called  []string
	failFor map[string]bool
}

func (m *mockRenewClient) RenewLease(_ context.Context, leaseID string, _ int) error {
	m.called = append(m.called, leaseID)
	if m.failFor[leaseID] {
		return errors.New("simulated renew failure")
	}
	return nil
}

func renewableLeaseWithTTL(id string, ttl time.Duration) LeaseStatus {
	return LeaseStatus{
		LeaseID:   id,
		Renewable: true,
		ExpireTime: time.Now().Add(ttl),
	}
}

func TestLeaseRenewer_ShouldRenew_True(t *testing.T) {
	r := NewLeaseRenewer(nil, 60*time.Second)
	ls := renewableLeaseWithTTL("lease/1", 30*time.Second)
	if !r.ShouldRenew(ls) {
		t.Error("expected ShouldRenew true for lease within leadway")
	}
}

func TestLeaseRenewer_ShouldRenew_False_HighTTL(t *testing.T) {
	r := NewLeaseRenewer(nil, 60*time.Second)
	ls := renewableLeaseWithTTL("lease/2", 5*time.Minute)
	if r.ShouldRenew(ls) {
		t.Error("expected ShouldRenew false for lease with high TTL")
	}
}

func TestLeaseRenewer_ShouldRenew_False_NotRenewable(t *testing.T) {
	r := NewLeaseRenewer(nil, 60*time.Second)
	ls := renewableLeaseWithTTL("lease/3", 10*time.Second)
	ls.Renewable = false
	if r.ShouldRenew(ls) {
		t.Error("expected ShouldRenew false for non-renewable lease")
	}
}

func TestLeaseRenewer_Renew_CallsClient(t *testing.T) {
	client := &mockRenewClient{failFor: map[string]bool{}}
	r := NewLeaseRenewer(client, 60*time.Second)
	ls := renewableLeaseWithTTL("lease/4", 20*time.Second)

	if err := r.Renew(context.Background(), ls, 3600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.called) != 1 || client.called[0] != "lease/4" {
		t.Errorf("expected client called with lease/4, got %v", client.called)
	}
}

func TestLeaseRenewer_RenewAll_CollectsErrors(t *testing.T) {
	client := &mockRenewClient{failFor: map[string]bool{"lease/bad": true}}
	r := NewLeaseRenewer(client, 60*time.Second)
	leases := []LeaseStatus{
		renewableLeaseWithTTL("lease/ok", 10*time.Second),
		renewableLeaseWithTTL("lease/bad", 10*time.Second),
	}
	errs := r.RenewAll(context.Background(), leases, 3600)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
	if len(client.called) != 2 {
		t.Errorf("expected both leases attempted, got %v", client.called)
	}
}

func TestNewLeaseRenewer_DefaultLeadway(t *testing.T) {
	r := NewLeaseRenewer(nil, 0)
	if r.leadway != 30*time.Second {
		t.Errorf("expected default leadway 30s, got %s", r.leadway)
	}
}
