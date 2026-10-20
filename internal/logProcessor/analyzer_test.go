package logprocessor

import (
	"testing"
	"time"

	"github.com/manelyhamedani/GhostOps/internal/config"
	"github.com/manelyhamedani/GhostOps/internal/model"
)

func TestRecordLog(t *testing.T) {
	cfg, err := config.NewConfig("../config/testConfig.json")
	if err != nil {
		t.Error(err)
	}
	a, err := NewAnalyzer(*cfg)
	if err != nil {
		t.Error(err)
	}
	ts, err := time.Parse(time.RFC3339, "2026-10-01T00:26:06Z")
	if err != nil {
		t.Error(err)
	}

	le := model.LogEntry{
		Timestamp: ts,
		Severity:  model.INFO,
		Service:   "payment-service",
		Message:   "User successfully logged in",
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
	if a.countsBySeverityLevels[model.INFO] != 1 {
		t.Errorf("severity")
	}
}
