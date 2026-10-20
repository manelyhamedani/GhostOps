package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/manelyhamedani/GhostOps/internal/model"
)

type Config struct {
	TimeSlice string
	Rules     []model.IncidentRule
}

func NewConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := Config{
		TimeSlice: "1m",
	}

	if err = json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	ts, err := time.ParseDuration(c.TimeSlice)
	if err != nil {
		return fmt.Errorf("Invalid time slice: %q, %w", ts, err)
	}

	if ts <= 0 {
		return fmt.Errorf("Time slice must be a positive number")
	}

	seen := make(map[string]bool)

	for i := range len(c.Rules) {
		if c.Rules[i].Name == "" {
			return fmt.Errorf("rules[%d]: name required", i)
		}
		if seen[c.Rules[i].Name] {
			return fmt.Errorf("rules[%d]: duplicate name", i)
		}
		seen[c.Rules[i].Name] = true
		if c.Rules[i].Threshold <= 0 {
			return fmt.Errorf("rules[%d]: threshold must be a positive number", i)
		}
	}

	return nil
}
