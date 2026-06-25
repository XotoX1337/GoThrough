// Package config loads and validates walkthrough YAML files.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Trigger defines how a step advances. Only "manual" is supported.
type Trigger struct {
	Type string `yaml:"type"`
}

// Quest is a quest-log reference attached to a step. Status, when set, must be
// "received" or "completed"; an empty status means the step merely mentions the
// quest without changing its state.
type Quest struct {
	Name   string `yaml:"name"`
	Status string `yaml:"status,omitempty"`
	Note   string `yaml:"note,omitempty"`
}

// Step is a single instruction in the walkthrough sequence. Description is
// Markdown (rendered in the HUD). Hints carry loot/XP detail; Warnings are
// rendered distinctly (e.g. "don't enter the back room").
type Step struct {
	ID          int      `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Optional    bool     `yaml:"optional,omitempty"`
	Quests      []Quest  `yaml:"quests,omitempty"`
	Hints       []string `yaml:"hints,omitempty"`
	Warnings    []string `yaml:"warnings,omitempty"`
	Trigger     Trigger  `yaml:"trigger"`
}

// BranchOption is one selectable path of a Branch. Its Steps become part of the
// active sequence once the option is chosen.
type BranchOption struct {
	Label       string `yaml:"label"`
	Description string `yaml:"description,omitempty"`
	Steps       []Node `yaml:"steps"`
}

// Branch is a decision point: the walkthrough forks into named Options and
// re-converges on whatever nodes follow the branch. The chosen option is
// persisted under PersistKey so it survives a restart.
type Branch struct {
	PersistKey string         `yaml:"persistKey"`
	Title      string         `yaml:"title"`
	Options    []BranchOption `yaml:"options"`
}

// Node is one entry in a steps list: either a plain Step or a Branch. The two
// are mutually exclusive; UnmarshalYAML decides by the presence of a `branch:`
// key.
type Node struct {
	Step   *Step
	Branch *Branch
}

// UnmarshalYAML distinguishes a branch node (`- branch: {...}`) from a plain
// step node by probing for the `branch` key first.
func (n *Node) UnmarshalYAML(value *yaml.Node) error {
	var probe struct {
		Branch *Branch `yaml:"branch"`
	}
	if err := value.Decode(&probe); err != nil {
		return err
	}
	if probe.Branch != nil {
		n.Branch = probe.Branch
		return nil
	}
	var s Step
	if err := value.Decode(&s); err != nil {
		return err
	}
	n.Step = &s
	return nil
}

// Section groups a list of nodes under a title for HUD rendering.
type Section struct {
	Title string `yaml:"title"`
	Steps []Node `yaml:"steps"`
}

// Walkthrough is the top-level structure of a walkthrough config file. A file
// uses EITHER Sections (grouped) or the flat Steps list (backward compatible);
// not both. Next, when set, names a follow-up file for a hand-off when the last
// step is reached.
type Walkthrough struct {
	Schema   int       `yaml:"schema,omitempty"`
	Game     string    `yaml:"game"`
	Version  string    `yaml:"version"`
	Variant  string    `yaml:"variant,omitempty"`
	Author   string    `yaml:"author"`
	Source   string    `yaml:"source,omitempty"`
	Chapter  int       `yaml:"chapter"`
	Day      int       `yaml:"day,omitempty"`
	Title    string    `yaml:"title"`
	Next     string    `yaml:"next,omitempty"`
	Sections []Section `yaml:"sections,omitempty"`
	Steps    []Node    `yaml:"steps,omitempty"`
}

// OutlineNode is a top-level node tagged with its section title (empty for a
// flat config). It preserves document order across sections so the engine can
// flatten the active sequence and the HUD can group by section.
type OutlineNode struct {
	Section string
	Step    *Step
	Branch  *Branch
}

// Outline returns every top-level node in document order, tagged with its
// section. Branch option steps are NOT expanded here — that is the engine's job
// once a branch is chosen.
func (w *Walkthrough) Outline() []OutlineNode {
	var out []OutlineNode
	if len(w.Sections) > 0 {
		for _, sec := range w.Sections {
			for _, n := range sec.Steps {
				out = append(out, OutlineNode{Section: sec.Title, Step: n.Step, Branch: n.Branch})
			}
		}
		return out
	}
	for _, n := range w.Steps {
		out = append(out, OutlineNode{Step: n.Step, Branch: n.Branch})
	}
	return out
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
	if len(wt.Sections) > 0 && len(wt.Steps) > 0 {
		return fmt.Errorf("use either `sections` or `steps`, not both")
	}

	outline := wt.Outline()
	if len(outline) == 0 {
		return fmt.Errorf("walkthrough must have at least one step")
	}

	seen := make(map[int]bool)
	count := 0
	for _, n := range outline {
		if n.Branch != nil {
			if err := validateBranch(n.Branch, seen, &count); err != nil {
				return err
			}
			continue
		}
		if err := validateStep(n.Step, seen, &count); err != nil {
			return err
		}
	}
	if count == 0 {
		return fmt.Errorf("walkthrough must have at least one step")
	}
	return nil
}

// validateStep checks a single step and records its ID. count is the running
// 1-based step number used in error messages and to confirm the file is
// non-empty.
func validateStep(s *Step, seen map[int]bool, count *int) error {
	*count++
	if s.ID == 0 {
		return fmt.Errorf("step %d: missing id", *count)
	}
	if seen[s.ID] {
		return fmt.Errorf("step %d: duplicate id %d", *count, s.ID)
	}
	seen[s.ID] = true
	if s.Title == "" {
		return fmt.Errorf("step %d (id=%d): missing title", *count, s.ID)
	}
	switch s.Trigger.Type {
	case "manual", "":
		// valid
	default:
		return fmt.Errorf("step %d (id=%d): unknown trigger type %q", *count, s.ID, s.Trigger.Type)
	}
	for _, q := range s.Quests {
		if q.Name == "" {
			return fmt.Errorf("step %d (id=%d): quest without name", *count, s.ID)
		}
		switch q.Status {
		case "", "received", "completed":
			// valid
		default:
			return fmt.Errorf("step %d (id=%d): quest %q has unknown status %q", *count, s.ID, q.Name, q.Status)
		}
	}
	return nil
}

func validateBranch(b *Branch, seen map[int]bool, count *int) error {
	if b.PersistKey == "" {
		return fmt.Errorf("branch %q: missing persistKey", b.Title)
	}
	if len(b.Options) < 2 {
		return fmt.Errorf("branch %q: needs at least 2 options", b.Title)
	}
	labels := make(map[string]bool)
	for i, o := range b.Options {
		if o.Label == "" {
			return fmt.Errorf("branch %q: option %d missing label", b.Title, i+1)
		}
		if labels[o.Label] {
			return fmt.Errorf("branch %q: duplicate option label %q", b.Title, o.Label)
		}
		labels[o.Label] = true
		if len(o.Steps) == 0 {
			return fmt.Errorf("branch %q: option %q has no steps", b.Title, o.Label)
		}
		for _, n := range o.Steps {
			if n.Branch != nil {
				if err := validateBranch(n.Branch, seen, count); err != nil {
					return err
				}
				continue
			}
			if err := validateStep(n.Step, seen, count); err != nil {
				return err
			}
		}
	}
	return nil
}
