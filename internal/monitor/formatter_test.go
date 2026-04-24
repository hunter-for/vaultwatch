package monitor

import (
	"strings"
	"testing"
	"time"
)

func TestFormat_Critical(t *testing.T) {
	alert := LeaseAlert{
		LeaseID:   "secret/db/creds",
		TTL:       2 * time.Minute,
		ExpiresAt: time.Now().Add(2 * time.Minute),
	}
	out := Format(alert)
	if !strings.Contains(out, "[CRITICAL]") {
		t.Errorf("expected CRITICAL severity, got: %s", out)
	}
	if !strings.Contains(out, "secret/db/creds") {
		t.Errorf("expected lease ID in output, got: %s", out)
	}
}

func TestFormat_Warning(t *testing.T) {
	alert := LeaseAlert{
		LeaseID: "secret/api/key",
		TTL:     20 * time.Minute,
	}
	out := Format(alert)
	if !strings.Contains(out, "[WARNING]") {
		t.Errorf("expected WARNING severity, got: %s", out)
	}
}

func TestFormat_Info(t *testing.T) {
	alert := LeaseAlert{
		LeaseID: "secret/svc/token",
		TTL:     45 * time.Minute,
	}
	out := Format(alert)
	if !strings.Contains(out, "[INFO]") {
		t.Errorf("expected INFO severity, got: %s", out)
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{90*time.Second, "1m 30s"},
		{3600 * time.Second, "1h"},
		{3661 * time.Second, "1h 1m 1s"},
		{0, "0s"},
	}
	for _, tc := range cases {
		got := formatDuration(tc.d)
		if got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}
