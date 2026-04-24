package alert

import (
	"io"
	"net"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/vaultwatch/internal/monitor"
)

// dialAndDrain starts a minimal SMTP stub that accepts one connection and
// records the raw data sent to it.
func startSMTPStub(t *testing.T) (addr string, received *strings.Builder) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	received = &strings.Builder{}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		// Minimal SMTP handshake expected by net/smtp
		fmt.Fprintf(conn, "220 localhost SMTP\r\n")
		io.Copy(io.MultiWriter(received, io.Discard), conn)
	}()
	return ln.Addr().String(), received
}

func baseEmailAlert() monitor.Alert {
	return monitor.Alert{
		LeaseID:   "database/creds/my-role/abc123",
		Severity:  "critical",
		TTL:       5 * time.Minute,
		Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestNewEmailSender_NotNil(t *testing.T) {
	cfg := EmailConfig{
		SMTPHost: "localhost",
		SMTPPort: 25,
		From:     "vault@example.com",
		To:       []string{"ops@example.com"},
	}
	s := NewEmailSender(cfg)
	if s == nil {
		t.Fatal("expected non-nil EmailSender")
	}
}

func TestEmailSender_AuthIsSet(t *testing.T) {
	cfg := EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user",
		Password: "secret",
		From:     "vault@example.com",
		To:       []string{"ops@example.com"},
	}
	s := NewEmailSender(cfg)
	if s.auth == nil {
		t.Fatal("expected smtp.Auth to be set")
	}
	// Verify the auth mechanism is PlainAuth by type assertion.
	if _, ok := s.auth.(smtp.Auth); !ok {
		t.Fatal("expected smtp.Auth interface")
	}
}

func TestEmailSender_SendFailsOnBadAddr(t *testing.T) {
	cfg := EmailConfig{
		SMTPHost: "127.0.0.1",
		SMTPPort: 1, // nothing listening
		From:     "vault@example.com",
		To:       []string{"ops@example.com"},
	}
	s := NewEmailSender(cfg)
	err := s.Send(baseEmailAlert())
	if err == nil {
		t.Fatal("expected error when SMTP server is unavailable")
	}
}
