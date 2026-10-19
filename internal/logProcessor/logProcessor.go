package logprocessor

import (
	"bufio"
	"errors"
	"os"
	"slices"
	"strings"
	"time"
)

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

func ProcessLogs(filePath string, analyzer *Analyzer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return ReadLine(file, func(line string) error {
		logEntry, err := ParseLine(line)
		isMalformed := false
		if errors.Is(err, ParseErr) {
			isMalformed = true
		} else if err != nil {
			return err
		}

		return analyzer.RecordLog(logEntry, isMalformed)
	})
}
