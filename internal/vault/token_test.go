package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func tokenLookupHandler(t *testing.T, payload map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/token/lookup-self" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(payload)
	}
}

func TestLookupSelfToken_Success(t *testing.T) {
	expireTime := time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339)
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"accessor":     "abc123",
			"display_name": "token-test",
			"policies":     []string{"default", "read-only"},
			"ttl":          3600,
			"renewable":    true,
			"expire_time":  expireTime,
		},
	}
	ts := httptest.NewServer(tokenLookupHandler(t, payload))
	defer ts.Close()

	client, err := NewClient(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := client.LookupSelfToken()
	if err != nil {
		t.Fatalf("LookupSelfToken: %v", err)
	}

	if info.Accessor != "abc123" {
		t.Errorf("expected accessor abc123, got %s", info.Accessor)
	}
	if info.DisplayName != "token-test" {
		t.Errorf("expected display_name token-test, got %s", info.DisplayName)
	}
	if !info.Renewable {
		t.Error("expected renewable to be true")
	}
	if info.TTL != 3600*time.Second {
		t.Errorf("expected TTL 3600s, got %v", info.TTL)
	}
	if info.ExpireTime.IsZero() {
		t.Error("expected ExpireTime to be set")
	}
	if len(info.Policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(info.Policies))
	}
}

func TestLookupSelfToken_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	client, _ := NewClient(ts.URL, "bad-token")
	_, err := client.LookupSelfToken()
	if err == nil {
		t.Fatal("expected error on non-200 status")
	}
}

func TestLookupSelfToken_BadURL(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:0", "token")
	_, err := client.LookupSelfToken()
	if err == nil {
		t.Fatal("expected error on unreachable server")
	}
}
