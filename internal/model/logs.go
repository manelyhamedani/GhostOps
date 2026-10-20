package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type SeverityLevel int

const (
	INFO SeverityLevel = iota
	WARN
	ERROR
	FATAL
)

var SeverityLevelsByName = map[string]SeverityLevel{
	"INFO":  INFO,
	"WARN":  WARN,
	"ERROR": ERROR,
	"FATAL": FATAL,
}

// TODO: Add constructor
type LogEntry struct {
	Timestamp time.Time
	Severity  SeverityLevel
	Service   string
	Message   string
}

// Satisfy Unmarshler interface
func (s *SeverityLevel) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Severity must be a string: %w", err)
	}

	val, ok := SeverityLevelsByName[str]
	if !ok {
		return fmt.Errorf("Unknown Severity %q", str)
	}

	*s = val
	return nil
}
