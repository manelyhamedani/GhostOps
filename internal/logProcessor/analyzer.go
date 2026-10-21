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
	analytics model.Analytics
}

func NewAnalyzer(cfg config.Config) (*Analyzer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	ts, err := time.ParseDuration(cfg.TimeSlice)
	if err != nil {
		return nil, err
	}

	Incidents, err := model.NewIncidents(cfg.Rules)
	if err != nil {
		return nil, err
	}

	a := &Analyzer{
		analytics: model.Analytics{
			CountsBySeverityLevels: make(map[model.SeverityLevel]int),
			CountsByServices:       make(map[string]int),
			TimeSlice:              ts,
			Incidents:              *Incidents,
			ServiceAnalytics:       make(model.ServiceToTimeSlice),
		},
	}

	return a, nil
}

func (a *Analyzer) RecordLog(logEntry model.LogEntry, isMalformed bool) error {
	if isMalformed {
		a.analytics.MalformedEntries++
		return nil
	}
	severity := logEntry.Severity
	service := logEntry.Service
	a.analytics.CountsBySeverityLevels[severity]++
	a.analytics.CountsByServices[service]++
	a.analytics.ValidEntries++

	sa, ok := a.analytics.ServiceAnalytics[service]
	if !ok {
		sa = make(model.TimeSliceToIncident)
		a.analytics.ServiceAnalytics[service] = sa
	}

	startTime := logEntry.Timestamp.UTC().Truncate(a.analytics.TimeSlice)
	ts, ok := sa[startTime]
	if !ok {
		ts = make(map[string]int)
		sa[startTime] = ts
	}

	//TODO: Refactor naming
	err := a.analytics.Incidents.MatchRules(logEntry, ts)
	if err != nil {
		return err
	}

	return nil
}

func (a *Analyzer) ReportText(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	fmt.Fprintf(tw, "SUMMARY\n=======\n")
	fmt.Fprintf(tw, "Valid entries:\t%d\n", a.analytics.ValidEntries)
	fmt.Fprintf(tw, "Malformed entries:\t%d\n", a.analytics.MalformedEntries)
	fmt.Fprintf(tw, "Time slice:\t%s\n", a.analytics.TimeSlice)

	fmt.Fprintf(tw, "BY SEVERITY\n-----------\n")
	sevKeys := slices.Sorted(maps.Keys(a.analytics.CountsBySeverityLevels))
	for _, sk := range sevKeys {
		fmt.Fprintf(tw, "%s\t%d\n", sk, a.analytics.CountsBySeverityLevels[sk])
	}

	fmt.Fprintf(tw, "BY SERVICE\n----------\n")
	svcKeys := slices.Sorted(maps.Keys(a.analytics.CountsByServices))
	for _, sk := range svcKeys {
		fmt.Fprintf(tw, "%s\t%d\n", sk, a.analytics.CountsByServices[sk])
	}

	fmt.Fprintf(tw, "INCIDENTS\n---------\n")
	for svc, ts := range a.analytics.ServiceAnalytics {
		for start, inc := range ts {
			for incName, freq := range inc {
				rule, ok := a.analytics.Incidents.Rules[incName]
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
					start.Add(a.analytics.TimeSlice).Format(time.RFC3339),
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

	jr.Summary.ValidEntries = a.analytics.ValidEntries
	jr.Summary.MalformedEntries = a.analytics.MalformedEntries
	jr.Summary.TimeSlice = a.analytics.TimeSlice.String()
	jr.Summary.CountsByServices = a.analytics.CountsByServices
	jr.Summary.CountsBySeverity = make(map[string]int, len(a.analytics.CountsBySeverityLevels))
	for sev, n := range a.analytics.CountsBySeverityLevels {
		jr.Summary.CountsBySeverity[sev.String()] = n
	}

	for svc, ts := range a.analytics.ServiceAnalytics {
		for start, inc := range ts {
			for incName, freq := range inc {
				rule, ok := a.analytics.Incidents.Rules[incName]
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
					EndTime:   start.Add(a.analytics.TimeSlice).Format(time.RFC3339),
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
