package logprocessor

import (
	"testing"
	"time"
)

func TestRecordLog(t *testing.T) {
	a, err := NewAnalyzer(time.Minute)
	if err != nil {
		t.Error(err)
	}
	ts, err := time.Parse(time.RFC3339, "2026-10-01T00:26:06Z")
	if err != nil {
		t.Error(err)
	}
	le := LogEntry{
		timestamp: ts,
		severity:  INFO,
		service:   "payment-service",
		message:   "User successfully logged in",
	}
	err = a.RecordLog(le, false)
	if err != nil {
		t.Error(err)
	}
	if a.malformedEntries != 0 {
		t.Errorf("malformed")
	}
	if a.validEntries != 1 {
		t.Errorf("valid")
	}
	if a.countsByServices["payment-service"] != 1 {
		t.Errorf("service")
	}
	if a.countsBySeverityLevels[INFO] != 1 {
		t.Errorf("severity")
	}
}
