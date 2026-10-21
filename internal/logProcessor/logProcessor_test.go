package logprocessor

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/manelyhamedani/GhostOps/internal/config"
	"github.com/manelyhamedani/GhostOps/internal/model"
)

func TestReadLine(t *testing.T) {
	filePath := "test.log"
	file, err := os.Open(filePath)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	var got []string

	err = ReadLine(file, func(line string) error {
		got = append(got, line)
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	want := []string{
		"2026-10-01T00:26:06Z INFO payment-service User successfully logged in",
		"2026-10-02T12:20:47Z FATAL user-service Database connection pool exhausted and crashed system",
		"2026-10-02T12:22:21Z ERROR order-service SQL syntax error while executing query: syntax error at end of input",
		"2026-10-02T12:15:09Z INFO notification-service Token generated for user session",
		"2026-10-02T12:17:44Z INFO auth-service User successfully logged in",
		"2026-10-02T12:18:17Z WARN order-service High memory usage detected: 85%",
		"2026-10-02T12:19:48Z WARN payment-service Deprecated API endpoint /api/old/users accessed",
		"--- LOG ENTRY START --- [incomplete]",
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("ReadLine() = %q, want = %q", got, want)
	}
}

func TestParseLine(t *testing.T) {
	validLine := "2026-10-01T00:26:06Z INFO payment-service User successfully logged in"
	invalidLine := "--- LOG ENTRY START --- [incomplete]"

	got, err := ParseLine(validLine)
	if err != nil {
		t.Error(err)
	}
	ts, err := time.Parse(time.RFC3339, "2026-10-01T00:26:06Z")
	if err != nil {
		t.Error(err)
	}

	want := model.LogEntry{
		Timestamp: ts,
		Severity:  model.INFO,
		Service:   "payment-service",
		Message:   "User successfully logged in",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseLine() = %+v, want = %+v", got, want)
	}

	got, err = ParseLine(invalidLine)
	if !errors.Is(err, ParseErr) {
		t.Errorf("ParseLine() err = %v, want = errors.Is(err, ParseErr)", err)
	}
}

func TestProcessLogs(t *testing.T) {
	filePath := "test.log"
	cfgPath := "../../config.json"
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		t.Error(err)
	}
	analyzer, err := NewAnalyzer(*cfg)
	if err != nil {
		t.Error(err)
	}

	err = ProcessLogs(filePath, analyzer)
	if err != nil {
		t.Error(err)
	}

	incidents, err := model.NewIncidents((*cfg).Rules)
	if err != nil {
		t.Error(err)
	}

	t1, _ := time.Parse(time.RFC3339, "2026-10-01T00:25:00Z")
	t2, _ := time.Parse(time.RFC3339, "2026-10-02T12:15:00Z")
	t3, _ := time.Parse(time.RFC3339, "2026-10-02T12:20:00Z")

	want := Analyzer{
		analytics: model.Analytics{
			ValidEntries:     7,
			MalformedEntries: 1,
			CountsBySeverityLevels: map[model.SeverityLevel]int{
				model.INFO:  3,
				model.FATAL: 1,
				model.ERROR: 1,
				model.WARN:  2,
			},
			CountsByServices: map[string]int{
				"payment-service":      2,
				"user-service":         1,
				"order-service":        2,
				"notification-service": 1,
				"auth-service":         1,
			},
			TimeSlice: 5 * time.Minute,
			Incidents: *incidents,
			ServiceAnalytics: model.ServiceToTimeSlice{
				"payment-service": {
					t1: map[string]int{},
					t2: map[string]int{
						"WARNING_SPIKE":        1,
						"DEPRECATED_API_USAGE": 1,
					},
				},
				"user-service": {
					t3: map[string]int{
						"FATAL_EVENT":                 1,
						"DATABASE_CONNECTION_FAILURE": 1,
					},
				},
				"order-service": {
					t3: map[string]int{
						"ERROR_SPIKE": 1,
						"SQL_ERROR":   1,
					},
					t2: map[string]int{
						"WARNING_SPIKE":     1,
						"HIGH_MEMORY_USAGE": 1,
					},
				},
				"notification-service": {
					t2: map[string]int{},
				},
				"auth-service": {
					t2: map[string]int{},
				},
			},
		},
	}

	got := *analyzer

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ProcessLogs() = %+v, want = %+v", got, want)
	}
}
