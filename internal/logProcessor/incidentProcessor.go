package logprocessor

import (
	"fmt"
	"slices"
	"strings"

	"github.com/manelyhamedani/GhostOps/internal/config"
	"github.com/manelyhamedani/GhostOps/internal/model"
)

type Incidents struct {
	rules map[string]model.IncidentRule
}

// NewIncidents loads rules from the config
func NewIncidents(cfg config.Config) (*Incidents, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	incidents := Incidents{
		rules: make(map[string]model.IncidentRule),
	}

	for _, r := range cfg.Rules {
		incidents.rules[r.Name] = r
	}

	return &incidents, nil
}

func (inc *Incidents) MatchRules(logEntry model.LogEntry, analytics map[string]int) error {
	if analytics == nil {
		return fmt.Errorf("Cannot append to nil map analytics")
	}

	severity := logEntry.Severity
	message := logEntry.Message

	for _, r := range inc.rules {
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

		analytics[r.Name]++
	}

	return nil
}
