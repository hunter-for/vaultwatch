package alert

import (
	"errors"
	"fmt"

	"github.com/user/vaultwatch/internal/monitor"
)

// Sender is the interface for anything that can dispatch an alert.
type Sender interface {
	Send(a monitor.Alert) error
}

// MultiSender fans out a single alert to multiple Sender implementations.
// All senders are attempted; errors are combined and returned together.
type MultiSender struct {
	senders []Sender
}

// NewMultiSender wraps the provided senders into a MultiSender.
func NewMultiSender(senders ...Sender) *MultiSender {
	return &MultiSender{senders: senders}
}

// Send delivers the alert to every registered sender.
// It continues on individual failures and returns a combined error if any occurred.
func (m *MultiSender) Send(a monitor.Alert) error {
	var errs []error
	for _, s := range m.senders {
		if err := s.Send(a); err != nil {
			errs = append(errs, fmt.Errorf("%T: %w", s, err))
		}
	}
	return errors.Join(errs...)
}
