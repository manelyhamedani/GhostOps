package model

import (
	"fmt"
	"slices"
	"strings"
)

type Incidents struct {
	Rules map[string]IncidentRule
}

func NewIncidents(rules []IncidentRule) (*Incidents, error) {
	incidents := Incidents{
		Rules: make(map[string]IncidentRule),
	}

	for _, r := range rules {
		incidents.Rules[r.Name] = r
	}

	return &incidents, nil
}

func (inc *Incidents) MatchRules(logEntry LogEntry, incToCount map[string]int) error {
	if incToCount == nil {
		return fmt.Errorf("Cannot append to nil map analytics")
	}

	severity := logEntry.Severity
	message := logEntry.Message

	for _, r := range inc.Rules {
		if !slices.Contains(r.Severity, severity) {
			continue
		}

		keywordExists := false
		for _, k := range r.Keywords {
			if strings.Contains(strings.ToLower(message), strings.ToLower(k)) {
				keywordExists = true
				break
			}
		}

		if len(r.Keywords) > 0 && !keywordExists {
			continue
		}

		incToCount[r.Name]++
	}

	return nil
}
