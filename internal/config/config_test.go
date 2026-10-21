package config

import (
	"reflect"
	"testing"

	"github.com/manelyhamedani/GhostOps/internal/model"
)

func TestNewConfig(t *testing.T) {
	path := "testConfig.json"
	cfg, err := NewConfig(path)
	if err != nil {
		t.Error(err)
	}
	got := *cfg

	want := Config{
		TimeSlice: "2m",
		Rules: []model.IncidentRule{
			{
				Name:             "DATABASE_CONNECTION_FAILURE",
				Severity:         []model.SeverityLevel{model.ERROR, model.FATAL},
				Threshold:        1,
				Keywords:         []string{"database", "connection"},
				IncidentSeverity: model.HIGH,
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewConfig() = %+v, want = %+v", got, want)
	}
}
