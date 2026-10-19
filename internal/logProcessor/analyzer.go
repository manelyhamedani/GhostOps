package logprocessor

import "time"

type Analyzer struct {
	validEntries           int
	malformedEntries       int
	countsBySeverityLevels map[SeverityLevel]int
	countsByServices       map[string]int
	timeSlices             time.Duration
	//TODO: Add incident rules analytics
}

func NewAnalyzer(timeSlices time.Duration) (*Analyzer, error) {
	a := &Analyzer{
		countsBySeverityLevels: make(map[SeverityLevel]int),
		countsByServices:       make(map[string]int),
		timeSlices:             timeSlices,
	}
	return a, nil
}

func (a *Analyzer) RecordLog(logEntry LogEntry, isMalformed bool) error {
	if isMalformed {
		a.malformedEntries++
		return nil
	}
	severity := logEntry.severity
	service := logEntry.service
	a.countsBySeverityLevels[severity]++
	a.countsByServices[service]++
	a.validEntries++
	return nil
}
