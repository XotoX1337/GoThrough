# GoThrough

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Status](https://img.shields.io/badge/status-WIP-yellow)

A game-agnostic walkthrough overlay platform written in Go. Load a YAML config for any game and follow step-by-step guides without alt-tabbing.

> Inspired by Zygor Guides for WoW — but for any game, community-driven, and open source.

## Features (planned)

- 🗂 Game-agnostic YAML/JSON config format
- 🪟 Always-on-top overlay window
- ✅ Manual step progression (keyboard shortcut or click)
- 💾 Persistent progress (resume where you left off)
- 🔍 Optional OCR-based auto-detection (screen capture polling)
- 🧩 Extendable trigger system (`manual` → `ocr` → `memory`)

## Planned Stack

| Component | Technology |
|---|---|
| Language | Go 1.21+ |
| CLI | [Cobra](https://github.com/spf13/cobra) |
| Overlay UI | [Fyne](https://fyne.io) or [Wails](https://wails.io) |
| Config format | YAML |
| Screen capture | [mss](https://github.com/ztrue/screenshot) or Win32 GDI |
| OCR | [gosseract](https://github.com/otiai10/gosseract) (Tesseract binding) |
| Image processing | [gocv](https://gocv.io) (OpenCV binding) |

## Config Format

Walkthroughs are defined as YAML files. Example:

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

  - id: 3
    title: "Talk to the Gate Guard"
    description: "Speak to the guard at the city gate to enter."
    trigger:
      type: manual
```

### Trigger types

| Type | Description | Status |
|---|---|---|
| `manual` | User clicks "Done" or presses a hotkey | Planned v1 |
| `ocr` | Screen capture + text recognition | Planned v2 |
| `memory` | Read game memory (game-specific) | Future |

## Project Structure (planned)

```
GoThrough/
├── cmd/               # Cobra CLI commands
├── config/            # YAML config loader & validator
├── engine/            # Step management & progress tracking
├── overlay/           # UI overlay window
├── capture/           # Screen capture & OCR (v2)
├── configs/           # Community walkthrough YAML files
│   └── gothic2/
│       └── chapter1.yaml
├── CONTEXT.md
└── README.md
```

## Roadmap

- [ ] v0.1 — Config loader + step engine (no UI)
- [ ] v0.2 — Basic overlay window (manual progression)
- [ ] v0.3 — Always-on-top + hotkey support
- [ ] v0.4 — Progress persistence
- [ ] v0.5 — OCR trigger support
- [ ] v1.0 — First full Gothic 2 walkthrough config

## Community Configs

The `configs/` directory is meant to grow into a community-maintained library of walkthrough configs for any game. If you write one, open a PR.

## Credits

Built with Go. Name inspired by Gothic 2 + the concept of a walkthrough guide. Get it? **Go**Through. 😄
