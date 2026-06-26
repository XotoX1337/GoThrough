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

// Task is one actionable sub-step within a Step. In YAML it is either a bare
// string (the instruction) or a mapping that attaches an info/warning/hint
// callout to that single sub-step. Text is Markdown (rendered inline in the HUD).
type Task struct {
	Text    string `yaml:"text"`
	Info    string `yaml:"info,omitempty"`
	Warning string `yaml:"warning,omitempty"`
	Hint    string `yaml:"hint,omitempty"`
}

// UnmarshalYAML accepts either a scalar (→ Text) or a mapping with explicit
// fields.
func (t *Task) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		return value.Decode(&t.Text)
	}
	type raw Task // avoid recursion into this method
	var r raw
	if err := value.Decode(&r); err != nil {
		return err
	}
	*t = Task(r)
	return nil
}

// StringList accepts either a single scalar or a sequence of scalars, so a
// condition value can be written as `key: value` or `key: [a, b]`.
type StringList []string

// UnmarshalYAML decodes a scalar into a one-element list, or a sequence as-is.
func (sl *StringList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		var s string
		if err := value.Decode(&s); err != nil {
			return err
		}
		*sl = StringList{s}
		return nil
	}
	var ss []string
	if err := value.Decode(&ss); err != nil {
		return err
	}
	*sl = ss
	return nil
}

// Condition gates a step on recorded choice answers: choiceKey → accepted
// values. A step is shown only when, for every key, the recorded answer is one
// of the accepted values (AND across keys, OR within a key's value list). An
// unanswered choice never matches, so dependent steps stay hidden until the
// user decides.
type Condition map[string]StringList

// Step is a single instruction in the walkthrough sequence. Description is an
// optional Markdown intro; Tasks is the actionable checklist within the step
// (each task may carry its own info/warning/hint). Step-level Hints/Warnings/
// Infos apply to the whole step. When, if set, gates the step on choice answers.
type Step struct {
	ID          int       `yaml:"id"`
	Title       string    `yaml:"title"`
	Description string    `yaml:"description,omitempty"`
	Tasks       []Task    `yaml:"tasks,omitempty"`
	When        Condition `yaml:"when,omitempty"`
	Optional    bool      `yaml:"optional,omitempty"`
	Quests      []Quest   `yaml:"quests,omitempty"`
	Hints       []string  `yaml:"hints,omitempty"`
	Warnings    []string  `yaml:"warnings,omitempty"`
	Infos       []string  `yaml:"infos,omitempty"`
	Trigger     Trigger   `yaml:"trigger"`
}

// ChoiceOption is one selectable answer of a Choice. Value is the stable key
// recorded in progress and referenced by step `when` conditions; Label is the
// human-readable text shown in the HUD.
type ChoiceOption struct {
	Value       string `yaml:"value"`
	Label       string `yaml:"label"`
	Description string `yaml:"description,omitempty"`
}

// Choice is a flat decision point that appears inline in the step list: it asks
// a question (Prompt) and records the answer under Key (persisted across
// restarts). It carries no nested steps — ordinary steps elsewhere in the file
// opt in via their `when` condition.
type Choice struct {
	Key     string         `yaml:"key"`
	Prompt  string         `yaml:"prompt"`
	Options []ChoiceOption `yaml:"options"`
}

// Node is one entry in a steps list: either a plain Step or a Choice. The two
// are mutually exclusive; UnmarshalYAML decides by the presence of a `choice:`
// key.
type Node struct {
	Step   *Step
	Choice *Choice
}

// UnmarshalYAML distinguishes a choice node (`- choice: {...}`) from a plain
// step node by probing for the `choice` key first.
func (n *Node) UnmarshalYAML(value *yaml.Node) error {
	var probe struct {
		Choice *Choice `yaml:"choice"`
	}
	if err := value.Decode(&probe); err != nil {
		return err
	}
	if probe.Choice != nil {
		n.Choice = probe.Choice
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
	Choice  *Choice
}

// Outline returns every top-level node in document order, tagged with its
// section.
func (w *Walkthrough) Outline() []OutlineNode {
	var out []OutlineNode
	if len(w.Sections) > 0 {
		for _, sec := range w.Sections {
			for _, n := range sec.Steps {
				out = append(out, OutlineNode{Section: sec.Title, Step: n.Step, Choice: n.Choice})
			}
		}
		return out
	}
	for _, n := range w.Steps {
		out = append(out, OutlineNode{Step: n.Step, Choice: n.Choice})
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

	// Pass 1: collect choices so step `when` conditions can be checked against
	// real keys/values.
	choices := map[string]map[string]bool{}
	for _, n := range outline {
		if n.Choice == nil {
			continue
		}
		c := n.Choice
		if c.Key == "" {
			return fmt.Errorf("choice %q: missing key", c.Prompt)
		}
		if _, dup := choices[c.Key]; dup {
			return fmt.Errorf("duplicate choice key %q", c.Key)
		}
		if len(c.Options) < 2 {
			return fmt.Errorf("choice %q: needs at least 2 options", c.Key)
		}
		values := map[string]bool{}
		for i, o := range c.Options {
			if o.Value == "" {
				return fmt.Errorf("choice %q: option %d missing value", c.Key, i+1)
			}
			if o.Label == "" {
				return fmt.Errorf("choice %q: option %q missing label", c.Key, o.Value)
			}
			if values[o.Value] {
				return fmt.Errorf("choice %q: duplicate option value %q", c.Key, o.Value)
			}
			values[o.Value] = true
		}
		choices[c.Key] = values
	}

	// Pass 2: steps (IDs, titles, triggers, quests, tasks, when references).
	seen := make(map[int]bool)
	count := 0
	for _, n := range outline {
		if n.Choice != nil {
			continue
		}
		if err := validateStep(n.Step, seen, &count, choices); err != nil {
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
// non-empty. choices is the set of declared choice keys→values, used to validate
// `when` conditions.
func validateStep(s *Step, seen map[int]bool, count *int, choices map[string]map[string]bool) error {
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
	for i, t := range s.Tasks {
		if t.Text == "" {
			return fmt.Errorf("step %d (id=%d): task %d missing text", *count, s.ID, i+1)
		}
	}
	for key, vals := range s.When {
		accepted, ok := choices[key]
		if !ok {
			return fmt.Errorf("step %d (id=%d): when references unknown choice %q", *count, s.ID, key)
		}
		for _, v := range vals {
			if !accepted[v] {
				return fmt.Errorf("step %d (id=%d): when choice %q has unknown value %q", *count, s.ID, key, v)
			}
		}
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
