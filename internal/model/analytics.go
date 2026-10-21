package model

import "time"

type MatchedIncidentToCount map[string]int

type TimeSliceToIncident map[time.Time]MatchedIncidentToCount

type ServiceToTimeSlice map[string]TimeSliceToIncident

type Analytics struct {
	ValidEntries           int
	MalformedEntries       int
	CountsBySeverityLevels map[SeverityLevel]int
	CountsByServices       map[string]int
	TimeSlice              time.Duration
	Incidents              Incidents
	ServiceAnalytics       ServiceToTimeSlice
}
