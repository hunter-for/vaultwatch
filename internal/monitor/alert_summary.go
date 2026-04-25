package monitor

import (
	"fmt"
	"strings"
	"time"
)

// AlertSummary aggregates alert counts and recent events for reporting.
type AlertSummary struct {
	GeneratedAt time.Time
	TotalAlerts int
	BySeverity  map[Severity]int
	RecentEvents []AlertEvent
	Suppressed  int
}

// NewAlertSummary builds an AlertSummary from an AlertHistory and DedupStore.
func NewAlertSummary(history *AlertHistory, dedup *DedupStore) AlertSummary {
	counts := history.CountBySeverity()
	recent := history.Recent(5)

	suppressed := 0
	if dedup != nil {
		suppressed = dedup.SuppressedCount()
	}

	total := 0
	for _, c := range counts {
		total += c
	}

	return AlertSummary{
		GeneratedAt:  time.Now().UTC(),
		TotalAlerts:  total,
		BySeverity:   counts,
		RecentEvents: recent,
		Suppressed:   suppressed,
	}
}

// Format returns a human-readable summary string.
func (s AlertSummary) Format() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Alert Summary [%s] ===\n",
		s.GeneratedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Total Alerts : %d\n", s.TotalAlerts))
	sb.WriteString(fmt.Sprintf("Suppressed   : %d\n", s.Suppressed))

	for _, sev := range []Severity{SeverityCritical, SeverityWarning, SeverityInfo} {
		if count, ok := s.BySeverity[sev]; ok && count > 0 {
			sb.WriteString(fmt.Sprintf("  %-10s: %d\n", sev.String(), count))
		}
	}

	if len(s.RecentEvents) > 0 {
		sb.WriteString("\nRecent Events:\n")
		for _, ev := range s.RecentEvents {
			sb.WriteString(fmt.Sprintf("  [%s] %s — %s\n",
				ev.Severity.String(),
				ev.LeaseID,
				ev.FiredAt.Format(time.RFC3339),
			))
		}
	}

	return sb.String()
}
