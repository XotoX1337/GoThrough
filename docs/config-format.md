# Walkthrough Config Format

Walkthroughs are plain YAML files. This page is the full reference; the
[README](../README.md#write-a-walkthrough) has the quick version.

## Minimal walkthrough

The smallest valid config is a flat list of steps:

```yaml
game: Gothic 2
version: vanilla
author: yourname
chapter: 1
title: "Chapter 1 - Arrival in Khorinis"

steps:
  - id: 1
    title: "Leave Xardas' Tower"
    description: "Go down the stairs and exit the tower."

  - id: 2
    title: "Head to Khorinis"
    description: "Follow the southern path to the city gate."
```

Progression is user-driven: steps advance on a **Next** click or hotkey.
`trigger` defaults to `manual` and can be omitted. Automatic triggers
(OCR, memory-reading) were intentionally left out — a game-agnostic tool has no
reliable way to read quest state.

## Top-level fields

| Field | Required | Description |
|---|---|---|
| `game` | yes | Game name. Groups chapters in the picker. |
| `title` | yes | Walkthrough/chapter title shown in the HUD header. |
| `version` | no | Game version/edition label (e.g. `vanilla`, `Die Nacht des Raben`). |
| `variant` | no | Whole-file path label (e.g. a class or guild route). |
| `author` | no | Author credit. |
| `source` | no | Attribution for the source material. |
| `chapter` | no | Chapter number (orders chapters in the picker). |
| `day` | no | Optional finer ordering within a chapter. |
| `schema` | no | Schema version. Current is `3`; older flat configs still load. |
| `next` | no | Relative path to the next file; the HUD shows a hand-off button at the end. |
| `steps` | one of | A flat list of steps. |
| `sections` | one of | Steps grouped into named sections. Use **either** `steps` **or** `sections`. |

## Sections

Group steps into named sections for larger walkthroughs:

```yaml
sections:
  - title: "Khorinis"
    steps:
      - id: 1
        title: "Reach the city"
```

## Steps

| Field | Description |
|---|---|
| `id` | Stable step identifier (used to restore progress across edits). |
| `title` | Short step title. |
| `description` | Optional Markdown prose (`**bold**`, `*italic*`). |
| `optional` | `true` marks the step as skippable. |
| `tasks` | Actionable sub-steps (a checklist). See below. |
| `quests` | Quest state entries: `{ name, status, note }`, `status` ∈ `received` \| `completed`. |
| `hints` | List of hint callouts. |
| `warnings` | List of warning callouts (red). |
| `infos` | List of info callouts (blue). |
| `when` | Conditions gating visibility based on recorded choices. See [Choices](#choices). |

### Tasks

A task is either a plain string or a mapping that attaches its own callout:

```yaml
tasks:
  - Follow the path **south**.
  - text: Avoid the *field raiders*.
    warning: They hit hard at low level.   # per-task callout
  - text: Grab the chest by the gate.
    info: Holds **50 gold**.
    hint: Lockpick from the left.
```

A task mapping supports `text` plus optional `info`, `warning`, and `hint`.

> **YAML gotcha:** a bare task string containing `": "` parses as a mapping.
> Quote it: `- "Talk to Diego: he points you on."`

## Choices

A choice is its own node in the steps list. It records the player's answer under
a `key`; steps then opt in to an answer via a `when` condition. Choices
re-converge — steps without a `when` are always shown.

```yaml
sections:
  - title: "Guild choice"
    steps:
      - choice:                # recorded under `key`, persisted in progress
          key: guild
          prompt: "Which guild do you join?"
          options:
            - { value: militia,   label: "Militia",   description: "City & order." }
            - { value: mercenary, label: "Mercenary" }
      - id: 100
        title: "Join the militia"
        when: { guild: militia }        # shown only for this answer
      - id: 200
        title: "Join the mercenaries"
        when: { guild: mercenary }
      - id: 2
        title: "Shared ending"          # no `when` → always shown (re-converges)
```

`when` matches recorded choices: a single value (`when: { guild: militia }`) or a
list for OR (`when: { guild: [militia, paladin] }`). Multiple keys are ANDed.
An unanswered choice hides its dependent steps until the player picks.

## Chaining files with `next`

Each file is its own walkthrough with its own progress. `next` points at the
file to hand off to when the current one is finished, resolved relative to the
current file's location:

```yaml
next: "day2.yaml"
```

When the player reaches the end, the HUD shows a hand-off button that loads the
next file.

## Full example

```yaml
schema: 3
game: Gothic 2
version: Die Nacht des Raben
variant: Drachenjäger
author: yourname
source: "Forum walkthrough X"
chapter: 1
day: 1
title: "Chapter 1 - Day 1"
next: "day2.yaml"

sections:
  - title: "Khorinis"
    steps:
      - id: 1
        title: "Reach the city"
        optional: true
        tasks:
          - Follow the path **south**.
          - text: Avoid the *field raiders*.
            warning: They hit hard at low level.
          - text: Grab the chest by the gate.
            info: Holds **50 gold**.
        quests:
          - { name: "Into the city", status: received }
        infos:    ["Town opens up after this."]
        hints:    ["Talk to the gate guard first."]
        warnings: ["Don't enter the back room — an orc waits there."]

  - title: "Guild choice"
    steps:
      - choice:
          key: guild
          prompt: "Which guild do you join?"
          options:
            - { value: militia,   label: "Militia",   description: "City & order." }
            - { value: mercenary, label: "Mercenary" }
      - id: 100
        title: "Join the militia"
        when: { guild: militia }
      - id: 200
        title: "Join the mercenaries"
        when: { guild: mercenary }
      - id: 2
        title: "Shared ending"
```
