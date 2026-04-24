package alert

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/vaultwatch/internal/monitor"
)

// EmailConfig holds SMTP configuration for email alerts.
type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
	To       []string
}

// EmailSender sends alert notifications via email.
type EmailSender struct {
	cfg  EmailConfig
	auth smtp.Auth
}

// NewEmailSender creates a new EmailSender with the given SMTP configuration.
func NewEmailSender(cfg EmailConfig) *EmailSender {
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
	return &EmailSender{cfg: cfg, auth: auth}
}

// Send delivers an alert via email.
func (e *EmailSender) Send(a monitor.Alert) error {
	subject := fmt.Sprintf("[VaultWatch][%s] Lease expiring: %s", strings.ToUpper(a.Severity), a.LeaseID)
	body := fmt.Sprintf(
		"Vault lease alert\r\n\r\nLease ID : %s\r\nSeverity : %s\r\nTTL      : %s\r\nTime     : %s\r\n",
		a.LeaseID,
		a.Severity,
		a.TTL.Round(time.Second),
		a.Timestamp.Format(time.RFC3339),
	)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		e.cfg.From,
		strings.Join(e.cfg.To, ", "),
		subject,
		body,
	)

	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPHost, e.cfg.SMTPPort)
	return smtp.SendMail(addr, e.auth, e.cfg.From, e.cfg.To, []byte(msg))
}
