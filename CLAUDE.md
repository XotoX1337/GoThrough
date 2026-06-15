# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoThrough is a game-agnostic walkthrough overlay platform written in Go. Inspired by Zygor Guides (WoW addon) — but open source and for any game. Users load a YAML walkthrough config and follow steps via an always-on-top overlay without alt-tabbing.

**GitHub:** https://github.com/XotoX1337/GoThrough  
**Status:** v0.2 done — Wails overlay window working.

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.21+ |
| CLI | [Cobra](https://github.com/spf13/cobra) |
| Overlay UI | [Wails v2](https://wails.io) |
| Config format | YAML |
| Screen capture | Win32 GDI or `ztrue/screenshot` |
| OCR (v2) | gosseract (Tesseract C-binding) |
| Image processing (v2) | gocv (OpenCV C-binding) |
| Target platform | Windows (primary), Linux (secondary) |

CGo (gosseract, gocv) is a first encounter for this project — handle with care.

## Planned Package Structure

```
GoThrough/
├── cmd/               # Cobra CLI commands (entry points)
├── config/            # YAML config loader & validator
├── engine/            # Step management & progress tracking
├── overlay/           # UI overlay window (always-on-top)
├── capture/           # Screen capture & OCR (v2 only)
├── configs/           # Community walkthrough YAML files
│   └── gothic2/
│       └── chapter1.yaml
```

## Walkthrough Config Format (YAML)

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

Trigger types: `manual` (v1), `ocr` (v2), `memory` (future).

## Build

```bash
wails build -s        # production build → build/bin/GoThrough.exe
wails dev             # dev mode with hot reload (requires frontend served separately)
```

Note: `-s` skips Wails' npm frontend pipeline — we embed assets directly via `//go:embed`.

## Roadmap

- [x] v0.1 — Config loader + step engine (no UI)
- [x] v0.2 — Basic overlay window (manual progression)
- [ ] v0.3 — Always-on-top + hotkey support
- [ ] v0.4 — Progress persistence
- [ ] v0.5 — OCR trigger support
- [ ] v1.0 — First full Gothic 2 walkthrough config

## Open Design Decisions

- **Always-on-top on Windows:** Win32 API directly or via UI framework?
- **Screenshot approach for DirectX games (Gothic 2):** GDI, DWM, or third-party lib?
- **Progress persistence:** Simple JSON file or SQLite?

## Developer Context

- Author has solid Go experience (see `tinymail` SMTP library, `dogo` Docker CLI helper)
- Familiar with Cobra, interfaces, struct patterns, Go module system
- First encounter with CGo (gosseract/gocv) — keep CGo isolated to `capture/` package
