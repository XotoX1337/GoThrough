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
schema: 2
game: Gothic 2
title: "Chapter 1 - Sections & Branch"
next: "chapter2.yaml"
sections:
  - title: "Intro"
    steps:
      - id: 1
        title: "Shared start"
        optional: true
        quests:
          - { name: "Main Quest", status: received }
        hints: ["a hint"]
        warnings: ["careful"]
  - title: "Fork"
    steps:
      - branch:
          persistKey: guild
          title: "Pick a guild"
          options:
            - label: "Militia"
              steps:
                - id: 10
                  title: "Join militia"
            - label: "Mercenary"
              steps:
                - id: 20
                  title: "Join mercenaries"
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

func TestLoadsSectionedBranchingConfig(t *testing.T) {
	wt, err := LoadBytes([]byte(sectioned))
	if err != nil {
		t.Fatalf("sectioned config: %v", err)
	}
	if wt.Schema != 2 || wt.Next != "chapter2.yaml" {
		t.Fatalf("metadata: schema=%d next=%q", wt.Schema, wt.Next)
	}
	out := wt.Outline()
	// Intro(1) + branch + Shared end(2) = 3 top-level nodes.
	if len(out) != 3 {
		t.Fatalf("outline length = %d, want 3", len(out))
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
	if out[1].Branch == nil || out[1].Branch.PersistKey != "guild" || len(out[1].Branch.Options) != 2 {
		t.Fatalf("branch parse wrong: %+v", out[1].Branch)
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

func TestRejectsDuplicateIDAcrossBranch(t *testing.T) {
	const dup = `
game: g
title: t
steps:
  - id: 1
    title: a
  - branch:
      persistKey: k
      title: pick
      options:
        - label: A
          steps: [{id: 1, title: x}]
        - label: B
          steps: [{id: 2, title: y}]
`
	if _, err := LoadBytes([]byte(dup)); err == nil || !strings.Contains(err.Error(), "duplicate id") {
		t.Fatalf("expected duplicate id error, got %v", err)
	}
}

func TestRejectsBranchWithOneOption(t *testing.T) {
	const one = `
game: g
title: t
steps:
  - branch:
      persistKey: k
      title: pick
      options:
        - label: A
          steps: [{id: 1, title: x}]
`
	if _, err := LoadBytes([]byte(one)); err == nil || !strings.Contains(err.Error(), "at least 2 options") {
		t.Fatalf("expected >=2 options error, got %v", err)
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
