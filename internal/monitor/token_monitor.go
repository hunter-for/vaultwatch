package monitor

import (
	"fmt"
	"time"
)

// TokenStatus represents the current state of a Vault token.
type TokenStatus struct {
	Accessor    string
	DisplayName string
	TTL         time.Duration
	Renewable   bool
	ExpireTime  time.Time
	Severity    Severity
}

// TokenLookup is the interface for retrieving token metadata from Vault.
type TokenLookup interface {
	LookupSelfToken() (accessor, displayName string, ttl time.Duration, renewable bool, expireTime time.Time, err error)
}

// CheckToken evaluates the current token's TTL against the provided thresholds
// and returns a TokenStatus with a computed severity level.
func CheckToken(ttl time.Duration, renewable bool, accessor, displayName string, expireTime time.Time, thresholds Thresholds) TokenStatus {
	severity := Classify(ttl, thresholds)
	return TokenStatus{
		Accessor:    accessor,
		DisplayName: displayName,
		TTL:         ttl,
		Renewable:   renewable,
		ExpireTime:  expireTime,
		Severity:    severity,
	}
}

// FormatTokenAlert returns a human-readable alert string for a token status.
func FormatTokenAlert(ts TokenStatus) string {
	if ts.TTL <= 0 {
		return fmt.Sprintf("[%s] token %s (%s) has EXPIRED",
			ts.Severity, ts.Accessor, ts.DisplayName)
	}
	return fmt.Sprintf("[%s] token %s (%s) expires in %s (renewable: %v)",
		ts.Severity, ts.Accessor, ts.DisplayName, formatDuration(ts.TTL), ts.Renewable)
}
