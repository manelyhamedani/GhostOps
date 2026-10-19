package logprocessor

import "time"

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
