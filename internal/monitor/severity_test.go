package monitor

import (
	"testing"
	"time"
)

func TestSeverityString(t *testing.T) {
	cases := []struct {
		sev  Severity
		want string
	}{
		{SeverityCritical, "CRITICAL"},
		{SeverityWarning, "WARNING"},
		{SeverityInfo, "INFO"},
	}
	for _, tc := range cases {
		if got := tc.sev.String(); got != tc.want {
			t.Errorf("Severity(%d).String() = %q, want %q", tc.sev, got, tc.want)
		}
	}
}

func TestClassify_Critical(t *testing.T) {
	th := DefaultThresholds()
	got := Classify(30*time.Minute, th)
	if got != SeverityCritical {
		t.Errorf("expected SeverityCritical, got %v", got)
	}
}

func TestClassify_CriticalBoundary(t *testing.T) {
	th := DefaultThresholds()
	got := Classify(th.Critical, th)
	if got != SeverityCritical {
		t.Errorf("expected SeverityCritical at boundary, got %v", got)
	}
}

func TestClassify_Warning(t *testing.T) {
	th := DefaultThresholds()
	got := Classify(3*time.Hour, th)
	if got != SeverityWarning {
		t.Errorf("expected SeverityWarning, got %v", got)
	}
}

func TestClassify_WarningBoundary(t *testing.T) {
	th := DefaultThresholds()
	got := Classify(th.Warning, th)
	if got != SeverityWarning {
		t.Errorf("expected SeverityWarning at boundary, got %v", got)
	}
}

func TestClassify_Info(t *testing.T) {
	th := DefaultThresholds()
	got := Classify(24*time.Hour, th)
	if got != SeverityInfo {
		t.Errorf("expected SeverityInfo, got %v", got)
	}
}

func TestDefaultThresholds(t *testing.T) {
	th := DefaultThresholds()
	if th.Critical != 1*time.Hour {
		t.Errorf("expected Critical=1h, got %v", th.Critical)
	}
	if th.Warning != 6*time.Hour {
		t.Errorf("expected Warning=6h, got %v", th.Warning)
	}
}
