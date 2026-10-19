package logprocessor

import (
	"bufio"
	"os"
	"time"
)

type SeverityLevel int

const (
	INFO SeverityLevel = iota
	WARN
	ERROR
	FATAL
)

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

func ParseLine(line string) (LogEntry, error) {
	return LogEntry{}, nil
}
