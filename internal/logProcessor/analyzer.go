package logprocessor

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"text/tabwriter"
	"time"

	"github.com/manelyhamedani/GhostOps/internal/config"
	"github.com/manelyhamedani/GhostOps/internal/model"
)

type Analyzer struct {
	validEntries           int
	malformedEntries       int
	countsBySeverityLevels map[model.SeverityLevel]int
	countsByServices       map[string]int
	timeSlice              time.Duration
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
		timeSlice:              ts,
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

	startTime := logEntry.Timestamp.UTC().Truncate(a.timeSlice)
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

func (a *Analyzer) ReportText(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	fmt.Fprintf(tw, "SUMMARY\n=======\n")
	fmt.Fprintf(tw, "Valid entries:\t%d\n", a.validEntries)
	fmt.Fprintf(tw, "Malformed entries:\t%d\n", a.malformedEntries)
	fmt.Fprintf(tw, "Time slice:\t%s\n", a.timeSlice)

	fmt.Fprintf(tw, "BY SEVERITY\n-----------\n")
	sevKeys := slices.Sorted(maps.Keys(a.countsBySeverityLevels))
	for _, sk := range sevKeys {
		fmt.Fprintf(tw, "%s\t%d\n", sk, a.countsBySeverityLevels[sk])
	}

	fmt.Fprintf(tw, "BY SERVICE\n----------\n")
	svcKeys := slices.Sorted(maps.Keys(a.countsByServices))
	for _, sk := range svcKeys {
		fmt.Fprintf(tw, "%s\t%d\n", sk, a.countsByServices[sk])
	}

	fmt.Fprintf(tw, "INCIDENTS\n---------\n")
	for svc, ts := range a.serviceAnalytics {
		for start, inc := range ts {
			for incName, freq := range inc {
				rule, ok := a.incidents.rules[incName]
				if !ok {
					continue
				}

				if rule.Threshold > freq {
					continue
				}

				fmt.Fprintf(tw, "%s:\n- Affected service: %s\n- Time range: %s -> %s\n- Frequency: %d (Threshold = %d)\n\n",
					incName,
					svc,
					start.Format(time.RFC3339),
					start.Add(a.timeSlice).Format(time.RFC3339),
					freq,
					rule.Threshold,
				)
			}
		}
	}

	return tw.Flush()
}

type jsonIncident struct {
	Incident  string
	Service   string
	StartTime string
	EndTime   string
	Frequency int
	Threshold int
}

type jsonReport struct {
	Summary struct {
		ValidEntries     int
		MalformedEntries int
		TimeSlice        string
		CountsBySeverity map[string]int
		CountsByServices map[string]int
	}
	Incidents []jsonIncident
}

func (a *Analyzer) ReportJSON(w io.Writer) error {
	var jr jsonReport

	jr.Summary.ValidEntries = a.validEntries
	jr.Summary.MalformedEntries = a.malformedEntries
	jr.Summary.TimeSlice = a.timeSlice.String()
	jr.Summary.CountsByServices = a.countsByServices
	jr.Summary.CountsBySeverity = make(map[string]int, len(a.countsBySeverityLevels))
	for sev, n := range a.countsBySeverityLevels {
		jr.Summary.CountsBySeverity[sev.String()] = n
	}

	for svc, ts := range a.serviceAnalytics {
		for start, inc := range ts {
			for incName, freq := range inc {
				rule, ok := a.incidents.rules[incName]
				if !ok {
					continue
				}

				if rule.Threshold > freq {
					continue
				}

				jr.Incidents = append(jr.Incidents, jsonIncident{
					Incident:  incName,
					Service:   svc,
					StartTime: start.Format(time.RFC3339),
					EndTime:   start.Add(a.timeSlice).Format(time.RFC3339),
					Frequency: freq,
					Threshold: rule.Threshold,
				})
			}
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jr)
}
