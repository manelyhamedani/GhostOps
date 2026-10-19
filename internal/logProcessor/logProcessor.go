package logprocessor

import (
	"bufio"
	"os"
	"slices"
	"strings"
	"time"
)

type SeverityLevel int

const (
	INFO SeverityLevel = iota
	WARN
	ERROR
	FATAL
)

var severityLevels = []string{"INFO", "WARN", "ERROR", "FATAL"}

type LogEntry struct {
	timestamp time.Time
	severity  SeverityLevel
	service   string
	message   string
}

// Read file line by line and call lineHandler on each line
func ReadLine(file *os.File, lineHandler func(string) error) error {
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if err := lineHandler(line); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// Parse each line to log entry with format: timestamp severityLevel serviceName message
// timestamp is RFC3339
// severityLevel is in {INFO, WARN, ERROR, FATAL}
func ParseLine(line string) (LogEntry, error) {
	parts := strings.SplitN(line, " ", 4)
	if len(parts) != 4 {
		return LogEntry{}, ParseErr
	}

	ts, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return LogEntry{}, ParseErr
	}

	sv := slices.Index(severityLevels, parts[1])
	if sv == -1 {
		return LogEntry{}, ParseErr
	}

	logEntry := LogEntry{
		timestamp: ts,
		severity:  SeverityLevel(sv),
		service:   parts[2],
		message:   parts[3],
	}

	return logEntry, nil
}
