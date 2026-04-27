package monitor

import (
	"fmt"
	"time"

	"github.com/your-org/vaultwatch/internal/alert"
)

// RetrySender wraps an alert.Sender and retries on failure using a RetryTracker.
type RetrySender struct {
	inner   alert.Sender
	tracker *RetryTracker
	sleep   func(time.Duration)
}

// NewRetrySender creates a RetrySender wrapping the given sender with the provided policy.
func NewRetrySender(inner alert.Sender, policy RetryPolicy) *RetrySender {
	return &RetrySender{
		inner:   inner,
		tracker: NewRetryTracker(policy),
		sleep:   time.Sleep,
	}
}

// Send attempts to send the alert, retrying on failure according to the policy.
func (r *RetrySender) Send(a alert.Alert) error {
	leaseID := a.LeaseID
	var lastErr error

	for r.tracker.ShouldRetry(leaseID) {
		err := r.inner.Send(a)
		if err == nil {
			r.tracker.Reset(leaseID)
			return nil
		}
		lastErr = err
		delay := r.tracker.NextDelay(leaseID)
		r.sleep(delay)
	}

	return fmt.Errorf("%w: lease %s after %d attempts: %v",
		ErrMaxRetriesExceeded, leaseID, r.tracker.Attempts(leaseID), lastErr)
}
