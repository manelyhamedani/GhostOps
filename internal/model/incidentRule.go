package model

type IncidentRule struct {
	Name      string
	Severity  []SeverityLevel
	Threshold int
	Keywords  []string
}
