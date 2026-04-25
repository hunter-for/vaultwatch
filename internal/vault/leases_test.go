package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListLeases_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "LIST" {
			t.Errorf("expected LIST method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"keys": []string{"aws/creds/my-role/abc123", "aws/creds/my-role/def456"},
			},
		})
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	keys, err := c.ListLeases("aws/creds/my-role")
	if err != nil {
		t.Fatalf("ListLeases: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestListLeases_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	c, _ := NewClient(ts.URL, "bad-token")
	_, err := c.ListLeases("aws/creds/my-role")
	if err == nil {
		t.Fatal("expected error on 403, got nil")
	}
}

func TestGetLease_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":        "aws/creds/my-role/abc123",
				"renewable": true,
				"ttl":       3600,
			},
		})
	}))
	defer ts.Close()

	c, _ := NewClient(ts.URL, "test-token")
	entry, err := c.GetLease("aws/creds/my-role/abc123")
	if err != nil {
		t.Fatalf("GetLease: %v", err)
	}
	if entry.LeaseID != "aws/creds/my-role/abc123" {
		t.Errorf("unexpected LeaseID: %s", entry.LeaseID)
	}
	if !entry.Renewable {
		t.Error("expected renewable=true")
	}
	if entry.TTL != 3600*time.Second {
		t.Errorf("expected TTL=3600s, got %v", entry.TTL)
	}
}

func TestGetLease_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c, _ := NewClient(ts.URL, "test-token")
	_, err := c.GetLease("nonexistent/lease")
	if err == nil {
		t.Fatal("expected error on 404, got nil")
	}
}
