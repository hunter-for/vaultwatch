package monitor_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vaultwatch/internal/monitor"
	"github.com/vaultwatch/internal/vault"
)

func newTestVaultServer(ttl float64) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/sys/leases/lookup", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"ttl": ttl},
		})
	})
	return httptest.NewServer(mux)
}

func TestMonitor_AlertsOnLowTTL(t *testing.T) {
	srv := newTestVaultServer(30)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	m := monitor.New(client, 60*time.Second, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go m.Run(ctx, []string{"lease/test/123"})

	select {
	case alert := <-m.Alerts():
		if alert.LeaseID != "lease/test/123" {
			t.Errorf("expected leaseID %q, got %q", "lease/test/123", alert.LeaseID)
		}
		if alert.TTL != 30*time.Second {
			t.Errorf("expected TTL 30s, got %v", alert.TTL)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for alert")
	}
}

func TestMonitor_NoAlertOnHighTTL(t *testing.T) {
	srv := newTestVaultServer(3600)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	m := monitor.New(client, 60*time.Second, 20*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	go m.Run(ctx, []string{"lease/healthy/456"})

	select {
	case alert := <-m.Alerts():
		t.Errorf("unexpected alert for high-TTL lease: %+v", alert)
	case <-ctx.Done():
		// expected: no alert fired
	}
}
