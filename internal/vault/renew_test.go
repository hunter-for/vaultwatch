package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRenewLease_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/leases/renew" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["lease_id"] != "secret/data/myapp/token" {
			t.Errorf("unexpected lease_id: %v", req["lease_id"])
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"lease_id":       "secret/data/myapp/token",
			"lease_duration": 3600,
			"renewable":      true,
		})
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.RenewLease(context.Background(), "secret/data/myapp/token", 3600); err != nil {
		t.Fatalf("RenewLease: %v", err)
	}
}

func TestRenewLease_EmptyLeaseID(t *testing.T) {
	client, err := NewClient("http://127.0.0.1:8200", "token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.RenewLease(context.Background(), "", 3600); err == nil {
		t.Error("expected error for empty leaseID")
	}
}

func TestRenewLease_Non200Status(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "bad-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.RenewLease(context.Background(), "lease/1", 3600); err == nil {
		t.Error("expected error on 403 response")
	}
}

func TestRenewLease_BadURL(t *testing.T) {
	client, err := NewClient("http://127.0.0.1:1", "token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.RenewLease(context.Background(), "lease/unreachable", 0); err == nil {
		t.Error("expected error for unreachable server")
	}
}
