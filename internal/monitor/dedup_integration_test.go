package monitor_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// TestDedupStore_WithClassify verifies that DedupStore integrates correctly
// with the Classify function: as a lease TTL drops through severity thresholds
// the dedup store allows each escalation through.
func TestDedupStore_WithClassify(t *testing.T) {
	thresholds := monitor.DefaultThresholds()
	dedup := monitor.NewDedupStore()
	cooldown := time.Hour
	leaseID := "secret/data/myapp/db"

	// Simulate Info TTL — first alert should always go through.
	sev := monitor.Classify(45*time.Minute, thresholds)
	if !dedup.ShouldAlert(leaseID, sev, cooldown) {
		t.Fatal("expected Info alert to be allowed on first check")
	}

	// Same severity again within cooldown — should be suppressed.
	if dedup.ShouldAlert(leaseID, sev, cooldown) {
		t.Fatal("expected repeated Info alert to be suppressed")
	}

	// TTL drops to Warning threshold — escalation should bypass cooldown.
	sev = monitor.Classify(10*time.Minute, thresholds)
	if !dedup.ShouldAlert(leaseID, sev, cooldown) {
		t.Fatal("expected Warning escalation to bypass cooldown")
	}

	// TTL drops to Critical threshold — escalation should bypass cooldown again.
	sev = monitor.Classify(2*time.Minute, thresholds)
	if !dedup.ShouldAlert(leaseID, sev, cooldown) {
		t.Fatal("expected Critical escalation to bypass cooldown")
	}

	// Lease expires — evict from store.
	dedup.Evict(leaseID)
	if dedup.Len() != 0 {
		t.Fatalf("expected store to be empty after eviction, got %d", dedup.Len())
	}
}
