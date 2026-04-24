package alert_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/alert"
)

func baseAlert(sev alert.Severity) alert.Alert {
	return alert.Alert{
		LeaseID:   "database/creds/my-role/abc123",
		TTL:       5 * time.Minute,
		Severity:  sev,
		Message:   "lease expiring soon",
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
}

func TestStdoutSender_Critical(t *testing.T) {
	var buf bytes.Buffer
	s := &alert.StdoutSender{Out: &buf}

	if err := s.Send(baseAlert(alert.SeverityCritical)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "CRITICAL") {
		t.Errorf("expected CRITICAL in output, got: %s", out)
	}
	if !strings.Contains(out, "database/creds/my-role/abc123") {
		t.Errorf("expected lease ID in output, got: %s", out)
	}
}

func TestStdoutSender_Warning(t *testing.T) {
	var buf bytes.Buffer
	s := &alert.StdoutSender{Out: &buf}

	if err := s.Send(baseAlert(alert.SeverityWarning)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "WARNING") {
		t.Errorf("expected WARNING in output, got: %s", buf.String())
	}
}

func TestStdoutSender_IncludesTTL(t *testing.T) {
	var buf bytes.Buffer
	s := &alert.StdoutSender{Out: &buf}

	a := baseAlert(alert.SeverityInfo)
	a.TTL = 2*time.Minute + 30*time.Second

	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "2m30s") {
		t.Errorf("expected TTL '2m30s' in output, got: %s", buf.String())
	}
}

func TestStdoutSender_Timestamp(t *testing.T) {
	var buf bytes.Buffer
	s := &alert.StdoutSender{Out: &buf}

	if err := s.Send(baseAlert(alert.SeverityInfo)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "2024-01-15T10:00:00Z") {
		t.Errorf("expected RFC3339 timestamp in output, got: %s", buf.String())
	}
}
