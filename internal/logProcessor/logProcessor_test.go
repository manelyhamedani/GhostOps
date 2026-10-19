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
