package monitor_test

import (
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// baseTokenStatus returns a TokenStatus-like structure via CheckToken's expected inputs.
// We test CheckToken and FormatTokenAlert using realistic TTL scenarios.

func tokenLeaseWithTTL(ttl time.Duration, renewable bool) monitor.LeaseStatus {
	return monitor.NewLeaseStatus(
		"auth/token/self",
		ttl,
		renewable,
	)
}

func TestCheckToken_CriticalTTL(t *testing.T) {
	lease := tokenLeaseWithTTL(3*time.Minute, true)
	thresholds := monitor.DefaultThresholds()

	severity := monitor.Classify(lease.TimeRemaining(), thresholds)
	if severity != monitor.Critical {
		t.Errorf("expected Critical severity for 3m TTL, got %s", severity)
	}
}

func TestCheckToken_WarningTTL(t *testing.T) {
	lease := tokenLeaseWithTTL(20*time.Minute, true)
	thresholds := monitor.DefaultThresholds()

	severity := monitor.Classify(lease.TimeRemaining(), thresholds)
	if severity != monitor.Warning {
		t.Errorf("expected Warning severity for 20m TTL, got %s", severity)
	}
}

func TestCheckToken_InfoTTL(t *testing.T) {
	lease := tokenLeaseWithTTL(2*time.Hour, true)
	thresholds := monitor.DefaultThresholds()

	severity := monitor.Classify(lease.TimeRemaining(), thresholds)
	if severity != monitor.Info {
		t.Errorf("expected Info severity for 2h TTL, got %s", severity)
	}
}

func TestFormatTokenAlert_ContainsLeaseID(t *testing.T) {
	lease := tokenLeaseWithTTL(5*time.Minute, true)
	msg := monitor.FormatTokenAlert(lease)

	if !strings.Contains(msg, "auth/token/self") {
		t.Errorf("expected message to contain lease ID, got: %s", msg)
	}
}

func TestFormatTokenAlert_ContainsTTL(t *testing.T) {
	lease := tokenLeaseWithTTL(10*time.Minute, true)
	msg := monitor.FormatTokenAlert(lease)

	if !strings.Contains(msg, "10m") {
		t.Errorf("expected message to contain TTL duration, got: %s", msg)
	}
}

func TestFormatTokenAlert_ContainsRenewableStatus(t *testing.T) {
	renewable := tokenLeaseWithTTL(5*time.Minute, true)
	nonRenewable := tokenLeaseWithTTL(5*time.Minute, false)

	msgR := monitor.FormatTokenAlert(renewable)
	if !strings.Contains(msgR, "renewable") {
		t.Errorf("expected renewable message to contain 'renewable', got: %s", msgR)
	}

	msgNR := monitor.FormatTokenAlert(nonRenewable)
	if !strings.Contains(msgNR, "not renewable") {
		t.Errorf("expected non-renewable message to contain 'not renewable', got: %s", msgNR)
	}
}

func TestCheckToken_ExpiredLease(t *testing.T) {
	lease := tokenLeaseWithTTL(0, false)

	if !lease.IsExpired() {
		t.Error("expected lease with 0 TTL to be expired")
	}
}

func TestFormatTokenAlert_ExpiredContainsExpiredText(t *testing.T) {
	lease := tokenLeaseWithTTL(0, false)
	msg := monitor.FormatTokenAlert(lease)

	if !strings.Contains(msg, "expired") {
		t.Errorf("expected expired token message to contain 'expired', got: %s", msg)
	}
}
