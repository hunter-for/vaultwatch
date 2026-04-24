// Package alert defines the Alert type and Sender interface used by
// vaultwatch to dispatch lease-expiry notifications.
//
// Backends
//
// StdoutSender writes human-readable alerts to standard output and is the
// default backend when no external integration is configured.
//
// Additional backends (e.g. Slack, PagerDuty) can be added by implementing
// the Sender interface:
//
//	type Sender interface {
//		Send(a Alert) error
//	}
//
// Usage
//
//	sender := alert.NewStdoutSender()
//	err := sender.Send(alert.Alert{
//		LeaseID:  "database/creds/my-role/abc123",
//		TTL:      5 * time.Minute,
//		Severity: alert.SeverityCritical,
//		Timestamp: time.Now(),
//	})
package alert
