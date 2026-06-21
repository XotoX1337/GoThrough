# GoThrough

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Status](https://img.shields.io/badge/status-WIP-yellow)

A game-agnostic walkthrough overlay platform written in Go. Load a YAML config for any game and follow step-by-step guides without alt-tabbing.

> Inspired by Zygor Guides for WoW — but for any game, community-driven, and open source.

## Features

- Game-agnostic YAML config format
- Semi-transparent always-on-top overlay (HUD-style), anchored top-right
- Quest-checklist sidebar: click any step to jump to it
- Global hotkeys that work while the game has focus (next/prev/hide/quit)
- Movable, resizable, lockable window (drag only when unlocked, clamped to screen)
- Progress is saved automatically and resumes where you left off (per walkthrough)
- Extendable trigger system (`manual` → `ocr` → `memory`)

### Hotkeys

| Hotkey | Action |
|---|---|
| `Ctrl+Alt+→` | Next step |
| `Ctrl+Alt+←` | Previous step |
| `Ctrl+Alt+H` | Toggle overlay visibility |
| `Ctrl+Alt+Q` | Quit |

## Stack

| Component | Technology |
|---|---|
| Language | Go 1.21+ |
| CLI | [Cobra](https://github.com/spf13/cobra) |
| Overlay UI | [Wails v2](https://wails.io) |
| Config format | YAML |
| Screen capture | Win32 GDI or `ztrue/screenshot` *(v0.5)* |
| OCR | [gosseract](https://github.com/otiai10/gosseract) *(v0.5)* |
| Image processing | [gocv](https://gocv.io) *(v0.5)* |

## Requirements

- Go 1.21+
- [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- GCC (Windows: [MinGW via Scoop](https://scoop.sh) — `scoop install mingw`)
- WebView2 Runtime (pre-installed on Windows 10/11)

## Build & Run

```bash
wails build -s                                              # build → build/bin/GoThrough.exe
./build/bin/GoThrough.exe run configs/gothic2/chapter1.yaml # run a walkthrough
./build/bin/GoThrough.exe run config.yaml --fresh           # ignore saved progress, start at step 1
make run                                                    # shortcut: build + run
```

> Progress is stored as JSON under the OS user-config dir
> (`%AppData%\GoThrough\progress.json` on Windows) and restored on the next launch.

> `-s` skips Wails' npm pipeline — assets are embedded directly via `//go:embed`.

### Iterating on the UI (devui)

For fast HUD iteration without rebuilding the Wails app, `tools/devui` is a
pure-Go (stdlib only, no Node) dev server that serves the untouched
`overlay/frontend/index.html` with the Wails bindings mocked against real step
data and live-reloads on save:

```bash
go run ./tools/devui                       # → http://localhost:34116 (gothic2/chapter1)
go run ./tools/devui -config path/to.yaml  # preview any walkthrough
go run ./tools/devui -bg screenshot.png    # use a real game screenshot as the scene
```

> devui approximates the glassmorphism/blur look but cannot reproduce true window
> transparency — verify the final result with `wails build -s` over a running game.

## Config Format

Walkthroughs are defined as YAML files:

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
    trigger:
      type: manual

  - id: 2
    title: "Head to Khorinis"
    description: "Follow the southern path to the city gate."
    trigger:
      type: manual
```

### Trigger types

| Type | Description | Status |
|---|---|---|
| `manual` | User clicks Next or presses a hotkey | v0.1+ |
| `ocr` | Screen capture + text recognition | v0.5 |
| `memory` | Read game memory (game-specific) | Future |

## Project Structure

```
GoThrough/
├── cmd/               # Cobra CLI commands
├── config/            # YAML config loader & validator
├── engine/            # Step management & navigation
├── progress/          # JSON progress persistence (resume per walkthrough)
├── overlay/           # Wails UI window
│   ├── app.go         # Go backend (bound to frontend)
│   ├── overlay.go     # Wails app setup
│   └── frontend/      # HTML/CSS/JS HUD
├── configs/           # Community walkthrough YAML files
│   └── gothic2/
│       └── chapter1.yaml
├── tools/
│   └── devui/         # Pure-Go live-reload dev server for the HUD
├── scripts/           # Dev scripts
└── Makefile
```

## Roadmap

- [x] v0.1 — Config loader + step engine (no UI)
- [x] v0.2 — Basic overlay window (manual progression)
- [x] v0.3 — Always-on-top + global hotkeys; HUD wired to the engine *(in-game verification pending)*
- [x] v0.4 — Progress persistence (auto-saved per walkthrough, resumes on launch)
- [ ] v0.5 — OCR trigger support
- [ ] v1.0 — First full Gothic 2 walkthrough config

## Community Configs

The `configs/` directory is meant to grow into a community-maintained library of walkthrough configs. If you write one, open a PR.
