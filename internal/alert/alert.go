// Package alert provides alerting backends for vaultwatch.
package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Severity represents the urgency level of an alert.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityWarning  Severity = "WARNING"
	SeverityInfo     Severity = "INFO"
)

// Alert holds the data for a single lease expiry notification.
type Alert struct {
	LeaseID   string
	TTL       time.Duration
	Severity  Severity
	Message   string
	Timestamp time.Time
}

// Sender is the interface implemented by all alert backends.
type Sender interface {
	Send(a Alert) error
}

// StdoutSender writes alerts to an io.Writer (defaults to os.Stdout).
type StdoutSender struct {
	Out io.Writer
}

// NewStdoutSender returns a StdoutSender that writes to os.Stdout.
func NewStdoutSender() *StdoutSender {
	return &StdoutSender{Out: os.Stdout}
}

// Send formats and writes the alert to the configured writer.
func (s *StdoutSender) Send(a Alert) error {
	_, err := fmt.Fprintf(
		s.Out,
		"[%s] %s | lease=%s ttl=%s\n",
		a.Timestamp.UTC().Format(time.RFC3339),
		a.Severity,
		a.LeaseID,
		a.TTL.Round(time.Second),
	)
	return err
}
