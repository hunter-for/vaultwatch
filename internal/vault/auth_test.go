package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func tokenInfoHandler(t *testing.T, ttl int, renewable bool, expireTime string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/token/lookup-self" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"accessor":    "abc123",
				"policies":    []string{"default", "read-secrets"},
				"ttl":         ttl,
				"renewable":   renewable,
				"expire_time": expireTime,
			},
		})
	}
}

func TestGetTokenInfo_Success(t *testing.T) {
	expireStr := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	ts := httptest.NewServer(tokenInfoHandler(t, 7200, true, expireStr))
	defer ts.Close()

	client, err := NewClient(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := client.GetTokenInfo()
	if err != nil {
		t.Fatalf("GetTokenInfo: %v", err)
	}

	if info.Accessor != "abc123" {
		t.Errorf("expected accessor abc123, got %s", info.Accessor)
	}
	if len(info.Policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(info.Policies))
	}
	if info.TTL != 7200*time.Second {
		t.Errorf("expected TTL 7200s, got %v", info.TTL)
	}
	if !info.Renewable {
		t.Error("expected renewable to be true")
	}
	if info.ExpireTime.IsZero() {
		t.Error("expected non-zero ExpireTime")
	}
}

func TestGetTokenInfo_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	client, _ := NewClient(ts.URL, "bad-token")
	_, err := client.GetTokenInfo()
	if err == nil {
		t.Fatal("expected error on non-200 response")
	}
}

func TestGetTokenInfo_BadURL(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:1", "token")
	_, err := client.GetTokenInfo()
	if err == nil {
		t.Fatal("expected error on unreachable server")
	}
}

func TestGetTokenInfo_EmptyExpireTime(t *testing.T) {
	ts := httptest.NewServer(tokenInfoHandler(t, 3600, false, ""))
	defer ts.Close()

	client, _ := NewClient(ts.URL, "token")
	info, err := client.GetTokenInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.ExpireTime.IsZero() {
		t.Errorf("expected zero ExpireTime when not provided, got %v", info.ExpireTime)
	}
}
