package logprocessor

import (
	"time"

	"github.com/manelyhamedani/GhostOps/internal/config"
	"github.com/manelyhamedani/GhostOps/internal/model"
)

type Analyzer struct {
	validEntries           int
	malformedEntries       int
	countsBySeverityLevels map[model.SeverityLevel]int
	countsByServices       map[string]int
	timeSlices             time.Duration
	incidents              Incidents
	serviceAnalytics       map[string]map[time.Time]map[string]int
}

func NewAnalyzer(cfg config.Config) (*Analyzer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	ts, err := time.ParseDuration(cfg.TimeSlice)
	if err != nil {
		return nil, err
	}

	Incidents, err := NewIncidents(cfg)
	if err != nil {
		return nil, err
	}

	a := &Analyzer{
		countsBySeverityLevels: make(map[model.SeverityLevel]int),
		countsByServices:       make(map[string]int),
		timeSlices:             ts,
		incidents:              *Incidents,
		serviceAnalytics:       make(map[string]map[time.Time]map[string]int),
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

	sa, ok := a.serviceAnalytics[service]
	if !ok {
		sa = make(map[time.Time]map[string]int)
		a.serviceAnalytics[service] = sa
	}

	startTime := logEntry.Timestamp.UTC().Truncate(a.timeSlices)
	ts, ok := sa[startTime]
	if !ok {
		ts = make(map[string]int)
		sa[startTime] = ts
	}

	//TODO: Refactor naming
	err := a.incidents.MatchRules(logEntry, ts)
	if err != nil {
		return err
	}

	return nil
}

func (a *Analyzer) Report() error {
	return nil
}
