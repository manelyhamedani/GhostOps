package logprocessor

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
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
	want := LogEntry{
		timestamp: ts,
		severity:  INFO,
		service:   "payment-service",
		message:   "User successfully logged in",
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
	analyzer, err := NewAnalyzer(time.Minute)
	if err != nil {
		t.Error(err)
	}

	err = ProcessLogs(filePath, analyzer)
	if err != nil {
		t.Error(err)
	}

	want := Analyzer{
		validEntries:     2,
		malformedEntries: 1,
		countsBySeverityLevels: map[SeverityLevel]int{
			INFO:  1,
			FATAL: 1,
		},
		countsByServices: map[string]int{
			"payment-service": 1,
			"user-service":    1,
		},
		timeSlices: time.Minute,
	}

	got := *analyzer

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ProcessLogs() = %+v, want = %+v", got, want)
	}
}
