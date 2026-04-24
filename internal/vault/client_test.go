package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func TestNewClient_ValidAddress(t *testing.T) {
	c, err := NewClient("http://127.0.0.1:8200", "test-token")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestLookupLease_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/leases/lookup" {
			http.NotFound(w, r)
			return
		}
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"id":        "database/creds/my-role/abc123",
				"ttl":       float64(3600),
				"renewable": true,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	srv := newTestServer(t, handler)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	lease, err := c.LookupLease("database/creds/my-role/abc123")
	if err != nil {
		t.Fatalf("LookupLease: %v", err)
	}

	if lease.TTL != 3600*time.Second {
		t.Errorf("expected TTL 3600s, got %v", lease.TTL)
	}
	if !lease.Renewable {
		t.Error("expected lease to be renewable")
	}
	if lease.ExpireTime.Before(time.Now()) {
		t.Error("expected ExpireTime to be in the future")
	}
}

func TestIsHealthy_Sealed(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sealed":      true,
			"initialized": true,
		})
	}

	srv := newTestServer(t, handler)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	err = c.IsHealthy()
	if err == nil {
		t.Fatal("expected error for sealed vault, got nil")
	}
}
