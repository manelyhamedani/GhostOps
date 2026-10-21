package model

import (
	"encoding/json"
	"fmt"
)

type IncidentSeverity int

const (
	LOW IncidentSeverity = iota
	MEDIUM
	HIGH
)

var IncidentSeverityByName = map[string]IncidentSeverity{
	"LOW":    LOW,
	"MEDIUM": MEDIUM,
	"HIGH":   HIGH,
}

var NameOfIncidentSeverity = map[IncidentSeverity]string{
	LOW:    "LOW",
	MEDIUM: "MEDIUM",
	HIGH:   "HIGH",
}

func (is IncidentSeverity) String() string {
	return NameOfIncidentSeverity[is]
}

type IncidentRule struct {
	Name             string
	Severity         []SeverityLevel
	Threshold        int
	Keywords         []string
	IncidentSeverity IncidentSeverity
}

// Satisfy Unmarshler interface
func (s *IncidentSeverity) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Severity must be a string: %w", err)
	}

	val, ok := IncidentSeverityByName[str]
	if !ok {
		return fmt.Errorf("Unknown Severity %q", str)
	}

	*s = val
	return nil
}
