package model

type IncidentRule struct {
	Name      string
	Severity  []SeverityLevel
	Threshold int
	Keywords  []string
}

type Incidents struct {
	rules              map[string]IncidentRule
	incidentStatistics map[string]int
}
