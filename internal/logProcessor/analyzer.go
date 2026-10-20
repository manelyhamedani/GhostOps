package logprocessor

import (
	"time"

	"github.com/manelyhamedani/GhostOps/internal/model"
)

type Analyzer struct {
	validEntries           int
	malformedEntries       int
	countsBySeverityLevels map[model.SeverityLevel]int
	countsByServices       map[string]int
	timeSlices             time.Duration
	//TODO: Add incident rules analytics
}

func NewAnalyzer(timeSlices time.Duration) (*Analyzer, error) {
	a := &Analyzer{
		countsBySeverityLevels: make(map[model.SeverityLevel]int),
		countsByServices:       make(map[string]int),
		timeSlices:             timeSlices,
	}

	return a, nil
}

func (a *Analyzer) RecordLog(logEntry model.LogEntry, isMalformed bool) error {
	if isMalformed {
		a.malformedEntries++
		return nil
	}
	severity := logEntry.Severity
	service := logEntry.Service
	a.countsBySeverityLevels[severity]++
	a.countsByServices[service]++
	a.validEntries++
	return nil
}
