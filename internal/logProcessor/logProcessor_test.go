package logprocessor

import (
	"os"
	"reflect"
	"testing"
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
