# GoThrough

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Status](https://img.shields.io/badge/status-WIP-yellow)

A game-agnostic walkthrough overlay platform written in Go. Load a YAML config for any game and follow step-by-step guides without alt-tabbing.

> Community-driven and open source — for any game.

## Features

- Game-agnostic YAML config format
- Always-on-top overlay (HUD-style), anchored top-right
- Config picker on launch — double-click the binary, pick a walkthrough, go
- Reopens your last walkthrough automatically on the next launch
- Quest-checklist sidebar: click any step to jump to it
- Global hotkeys that work while the game has focus (next/prev/hide/quit)
- Movable, lockable window (drag only when unlocked, clamped to screen)
- Adjustable overlay opacity (solid by default, dial it down to a glassy HUD)
- Progress is saved automatically and resumes where you left off (per walkthrough)
- Manual step progression (Next/Prev) — no fragile auto-detection to lead you astray
- In-HUD settings panel (gear icon) with rebindable global hotkeys

### Hotkeys

Global hotkeys are **rebindable** from the in-HUD settings panel (click the gear).
The defaults are:

| Hotkey | Action |
|---|---|
| `Ctrl+Alt+→` | Next step |
| `Ctrl+Alt+←` | Previous step |
| `Ctrl+Alt+H` | Toggle overlay visibility |
| `Ctrl+Alt+Q` | Quit |

> Mouse buttons can be bound too (e.g. `Ctrl+Alt+MiddleClick`, side buttons). A bare
> click still passes straight through to the game. Mouse-button hotkeys are unsupported
> under Wayland and on macOS.

## Stack

| Component | Technology |
|---|---|
| Language | Go 1.21+ |
| CLI | [Cobra](https://github.com/spf13/cobra) |
| Overlay UI | [Wails v2](https://wails.io) |
| Config format | YAML |
| Global hotkeys | [`golang.design/x/hotkey`](https://github.com/golang-design/hotkey) |

## Requirements

- Go 1.21+
- [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- GCC (Windows: [MinGW via Scoop](https://scoop.sh) — `scoop install mingw`)
- WebView2 Runtime (pre-installed on Windows 10/11)
- Run your game in **windowed** or **borderless windowed** mode (the supported display modes for the overlay; true exclusive fullscreen may occlude it)

## Build & Run

```bash
wails build -s                                              # build → build/bin/GoThrough.exe
./build/bin/GoThrough.exe                                   # open the config picker
./build/bin/GoThrough.exe run path/to/walkthrough.yaml      # run a walkthrough directly
./build/bin/GoThrough.exe run path/to/walkthrough.yaml --fresh # ignore saved progress, start at step 1
make run                                                    # shortcut: build + open the picker
```

> Progress and settings are saved automatically under your OS user-config dir
> (`%AppData%\GoThrough\` on Windows, `~/.config/GoThrough/` on Linux). Delete that
> folder to reset, or use `--fresh` to ignore saved progress for one run.

## Config Format

Walkthroughs are defined as YAML files. The minimal form is a flat list of steps:

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

Progression is user-driven: steps advance on a Next click or hotkey. `trigger`
defaults to `manual` and can be omitted. Automatic triggers (OCR, memory-reading)
were intentionally left out — a game-agnostic tool has no reliable way to read quest state.

### Sections, choices & richer steps (schema 3)

Larger walkthroughs can group steps into **sections**, branch on flat
**`choices`** (a decision is recorded under a key; steps opt in via a `when`
condition), and chain across files with **`next`**. Steps carry a checklist of
**`tasks`** — each task can attach its own `info`/`warning`/`hint` — plus
step-level `hints`, `warnings`, `infos`, `quests`, an optional Markdown
`description`, and an `optional` flag. Everything is additive — flat
`steps`-only configs keep loading unchanged.

```yaml
schema: 3
game: Gothic 2
version: Die Nacht des Raben
variant: Drachenjäger          # whole-file path label
author: yourname
source: "Forum walkthrough X"  # attribution
chapter: 1
day: 1
title: "Chapter 1 - Day 1"
next: "day2.yaml"              # hand-off to the next file when finished

sections:                      # use EITHER sections OR a flat `steps:` list
  - title: "Khorinis"
    steps:
      - id: 1
        title: "Reach the city"
        optional: true
        tasks:                 # actionable sub-steps (Markdown: **bold**, *italic*)
          - Follow the path **south**.
          - text: Avoid the *field raiders*.
            warning: They hit hard at low level.   # per-task callout
          - text: Grab the chest by the gate.
            info: Holds **50 gold**.
        quests:
          - { name: "Into the city", status: received }   # received | completed
        infos:    ["Town opens up after this."]           # step-level (blue)
        hints:    ["Talk to the gate guard first."]
        warnings: ["Don't enter the back room — an orc waits there."]

  - title: "Guild choice"
    steps:
      - choice:                # decision recorded under `key`, saved in progress.json
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

## Roadmap

- [x] v0.1 — Config loader + step engine (no UI)
- [x] v0.2 — Basic overlay window (manual progression)
- [x] v0.3 — Always-on-top + global hotkeys; HUD wired to the engine *(verified in-game over Gothic 2)*
- [x] v0.4 — Progress persistence (auto-saved per walkthrough, resumes on launch)
- [x] v0.5 — Settings: persistent user config + rebindable hotkeys (in-HUD panel)
- [x] v0.6 — UX & distribution polish: config picker, double-click launch, opacity, reopen last walkthrough, binary signing
- [x] v0.7 — Branching & sections in the config format (Gothic 2's guild choice), Markdown steps, hints/warnings/quests, `next`-file chaining
- [x] v0.8 — HUD UI pass: collapsible sections, auto-scroll, themes (dark/light/contrast), focus-overlay hotkey
- [x] v0.9 — Config schema v3: flat `choices` + `when` (replacing nested branches), per-task `tasks` with info/warning/hint callouts, two-level picker (game → chapter)
- [ ] v1.0 — First full Gothic 2 walkthrough config

## Community Configs

Walkthroughs bundled with the binary live under `configstore/configs/`, meant to grow into a community-maintained library. If you write one, open a PR.
