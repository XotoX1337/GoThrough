package config

import (
	"strings"
	"testing"
)

// flatLegacy is a v0.6-era config: flat steps, no schema/sections. It must keep
// loading unchanged.
const flatLegacy = `
game: Gothic 2
title: "Chapter 1"
steps:
  - id: 1
    title: "First"
    description: "Do the thing."
    trigger: { type: manual }
  - id: 2
    title: "Second"
`

const sectioned = `
schema: 3
game: Gothic 2
title: "Chapter 1 - Sections & Choice"
next: "chapter2.yaml"
sections:
  - title: "Intro"
    steps:
      - id: 1
        title: "Shared start"
        optional: true
        tasks:
          - "Plain task"
          - text: "Task with a callout"
            warning: "be careful"
            info: "good to know"
        quests:
          - { name: "Main Quest", status: received }
        hints: ["a hint"]
        warnings: ["careful"]
        infos: ["heads up"]
  - title: "Fork"
    steps:
      - choice:
          key: guild
          prompt: "Pick a guild"
          options:
            - { value: militia, label: "Militia" }
            - { value: merc, label: "Mercenary" }
      - id: 10
        title: "Join militia"
        when: { guild: militia }
      - id: 20
        title: "Join mercenaries"
        when: { guild: [merc, militia] }
      - id: 2
        title: "Shared end"
`

func TestLoadsLegacyFlatConfig(t *testing.T) {
	wt, err := LoadBytes([]byte(flatLegacy))
	if err != nil {
		t.Fatalf("legacy config must still load: %v", err)
	}
	out := wt.Outline()
	if len(out) != 2 || out[0].Step == nil || out[0].Step.ID != 1 {
		t.Fatalf("legacy outline = %+v", out)
	}
	if out[0].Section != "" {
		t.Fatalf("flat config should have empty section, got %q", out[0].Section)
	}
}

func TestLoadsSectionedChoiceConfig(t *testing.T) {
	wt, err := LoadBytes([]byte(sectioned))
	if err != nil {
		t.Fatalf("sectioned config: %v", err)
	}
	if wt.Schema != 3 || wt.Next != "chapter2.yaml" {
		t.Fatalf("metadata: schema=%d next=%q", wt.Schema, wt.Next)
	}
	out := wt.Outline()
	// Intro(1) + choice + militia(10) + merc(20) + Shared end(2) = 5 nodes.
	if len(out) != 5 {
		t.Fatalf("outline length = %d, want 5", len(out))
	}
	if out[0].Section != "Intro" || out[1].Section != "Fork" {
		t.Fatalf("section tags wrong: %q / %q", out[0].Section, out[1].Section)
	}
	if !out[0].Step.Optional {
		t.Fatal("step 1 should be optional")
	}
	if out[0].Step.Quests[0].Status != "received" {
		t.Fatalf("quest status = %q", out[0].Step.Quests[0].Status)
	}
	if len(out[0].Step.Infos) != 1 || out[0].Step.Infos[0] != "heads up" {
		t.Fatalf("step infos = %+v", out[0].Step.Infos)
	}
	// Tasks: a bare string and a mapping with callouts.
	tasks := out[0].Step.Tasks
	if len(tasks) != 2 || tasks[0].Text != "Plain task" {
		t.Fatalf("tasks[0] = %+v", tasks)
	}
	if tasks[1].Text != "Task with a callout" || tasks[1].Warning != "be careful" || tasks[1].Info != "good to know" {
		t.Fatalf("tasks[1] = %+v", tasks[1])
	}
	// Choice.
	if out[1].Choice == nil || out[1].Choice.Key != "guild" || len(out[1].Choice.Options) != 2 {
		t.Fatalf("choice parse wrong: %+v", out[1].Choice)
	}
	// when: scalar and list forms.
	if got := out[2].Step.When["guild"]; len(got) != 1 || got[0] != "militia" {
		t.Fatalf("scalar when = %+v", got)
	}
	if got := out[3].Step.When["guild"]; len(got) != 2 || got[0] != "merc" || got[1] != "militia" {
		t.Fatalf("list when = %+v", got)
	}
}

func TestRejectsSectionsAndStepsTogether(t *testing.T) {
	const both = `
game: g
title: t
steps: [{id: 1, title: a}]
sections: [{title: s, steps: [{id: 2, title: b}]}]
`
	if _, err := LoadBytes([]byte(both)); err == nil || !strings.Contains(err.Error(), "not both") {
		t.Fatalf("expected 'not both' error, got %v", err)
	}
}

func TestRejectsDuplicateID(t *testing.T) {
	const dup = `
game: g
title: t
steps:
  - id: 1
    title: a
  - id: 1
    title: b
`
	if _, err := LoadBytes([]byte(dup)); err == nil || !strings.Contains(err.Error(), "duplicate id") {
		t.Fatalf("expected duplicate id error, got %v", err)
	}
}

func TestRejectsChoiceWithOneOption(t *testing.T) {
	const one = `
game: g
title: t
steps:
  - choice:
      key: k
      prompt: pick
      options:
        - { value: a, label: A }
  - id: 1
    title: x
`
	if _, err := LoadBytes([]byte(one)); err == nil || !strings.Contains(err.Error(), "at least 2 options") {
		t.Fatalf("expected >=2 options error, got %v", err)
	}
}

func TestRejectsDuplicateChoiceKey(t *testing.T) {
	const dup = `
game: g
title: t
steps:
  - choice: { key: k, prompt: a, options: [{value: x, label: X}, {value: y, label: Y}] }
  - choice: { key: k, prompt: b, options: [{value: x, label: X}, {value: y, label: Y}] }
  - id: 1
    title: s
`
	if _, err := LoadBytes([]byte(dup)); err == nil || !strings.Contains(err.Error(), "duplicate choice key") {
		t.Fatalf("expected duplicate choice key error, got %v", err)
	}
}

func TestRejectsWhenReferencingUnknownChoice(t *testing.T) {
	const bad = `
game: g
title: t
steps:
  - id: 1
    title: a
    when: { ghost: yes }
`
	if _, err := LoadBytes([]byte(bad)); err == nil || !strings.Contains(err.Error(), "unknown choice") {
		t.Fatalf("expected unknown choice error, got %v", err)
	}
}

func TestRejectsWhenReferencingUnknownValue(t *testing.T) {
	const bad = `
game: g
title: t
steps:
  - choice: { key: k, prompt: p, options: [{value: a, label: A}, {value: b, label: B}] }
  - id: 1
    title: s
    when: { k: c }
`
	if _, err := LoadBytes([]byte(bad)); err == nil || !strings.Contains(err.Error(), "unknown value") {
		t.Fatalf("expected unknown value error, got %v", err)
	}
}

func TestRejectsTaskWithoutText(t *testing.T) {
	const bad = `
game: g
title: t
steps:
  - id: 1
    title: a
    tasks:
      - { warning: "no text here" }
`
	if _, err := LoadBytes([]byte(bad)); err == nil || !strings.Contains(err.Error(), "task 1 missing text") {
		t.Fatalf("expected task text error, got %v", err)
	}
}

func TestRejectsUnknownQuestStatus(t *testing.T) {
	const bad = `
game: g
title: t
steps:
  - id: 1
    title: a
    quests: [{name: q, status: bogus}]
`
	if _, err := LoadBytes([]byte(bad)); err == nil || !strings.Contains(err.Error(), "unknown status") {
		t.Fatalf("expected quest status error, got %v", err)
	}
}
