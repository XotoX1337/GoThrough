// Package config loads and validates walkthrough YAML files.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Trigger defines how a step advances. Only "manual" is supported in v0.1.
type Trigger struct {
	Type string `yaml:"type"`
}

// Step is a single instruction in the walkthrough sequence.
type Step struct {
	ID          int     `yaml:"id"`
	Title       string  `yaml:"title"`
	Description string  `yaml:"description"`
	Trigger     Trigger `yaml:"trigger"`
}

// Walkthrough is the top-level structure of a walkthrough config file.
type Walkthrough struct {
	Game    string `yaml:"game"`
	Version string `yaml:"version"`
	Author  string `yaml:"author"`
	Chapter int    `yaml:"chapter"`
	Title   string `yaml:"title"`
	Steps   []Step `yaml:"steps"`
}

// Load reads and validates a walkthrough YAML file from disk.
func Load(path string) (*Walkthrough, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	return LoadBytes(data)
}

// LoadBytes parses and validates a walkthrough from raw YAML bytes.
// Used by the embedded config store and tests.
func LoadBytes(data []byte) (*Walkthrough, error) {
	var wt Walkthrough
	if err := yaml.Unmarshal(data, &wt); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := validate(&wt); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &wt, nil
}

func validate(wt *Walkthrough) error {
	if wt.Game == "" {
		return fmt.Errorf("missing required field: game")
	}
	if wt.Title == "" {
		return fmt.Errorf("missing required field: title")
	}
	if len(wt.Steps) == 0 {
		return fmt.Errorf("walkthrough must have at least one step")
	}

	seen := make(map[int]bool)
	for i, s := range wt.Steps {
		if s.ID == 0 {
			return fmt.Errorf("step %d: missing id", i+1)
		}
		if seen[s.ID] {
			return fmt.Errorf("step %d: duplicate id %d", i+1, s.ID)
		}
		seen[s.ID] = true
		if s.Title == "" {
			return fmt.Errorf("step %d (id=%d): missing title", i+1, s.ID)
		}
		switch s.Trigger.Type {
		case "manual", "":
			// valid
		default:
			return fmt.Errorf("step %d (id=%d): unknown trigger type %q", i+1, s.ID, s.Trigger.Type)
		}
	}

	return nil
}
