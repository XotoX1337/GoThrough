# GoThrough – Session Context

Paste this file at the start of every Claude session to restore project context.

---

## What is GoThrough?

A game-agnostic walkthrough overlay platform written in Go.
Inspired by Zygor Guides (WoW addon) — but for any game, open source, community-driven.
No navigation arrow needed — just step-by-step guidance with auto-progression.

GitHub: https://github.com/XotoX1337/GoThrough

---

## Tech Stack

- **Language:** Go 1.21+
- **CLI:** Cobra (already used in sibling project `dogo`)
- **Overlay UI:** Fyne or Wails (TBD)
- **Config format:** YAML
- **Screen capture:** Win32 GDI or `ztrue/screenshot`
- **OCR:** gosseract (Tesseract C-binding) — v2 feature
- **Image processing:** gocv (OpenCV C-binding) — v2 feature
- **Target platform:** Windows (primary), Linux (secondary)

---

## Project Structure

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

---

## Config Format (YAML)

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

### Trigger types (planned)
- `manual` — user clicks Done or presses hotkey (v1)
- `ocr` — screen capture + Tesseract text recognition (v2)
- `memory` — read game memory directly (future)

---

## Roadmap & Current Status

- [ ] v0.1 — Config loader + step engine (no UI) ← **start here**
- [ ] v0.2 — Basic overlay window (manual progression)
- [ ] v0.3 — Always-on-top + hotkey support
- [ ] v0.4 — Progress persistence
- [ ] v0.5 — OCR trigger support
- [ ] v1.0 — First full Gothic 2 walkthrough config

**Current phase:** Project setup — nothing implemented yet.

---

## Developer Background

- Go experience: solid (see `tinymail` SMTP library, `dogo` Docker CLI helper with Cobra)
- Familiar with: Cobra, interfaces, struct patterns, Go module system
- CGo experience: unknown — gosseract/gocv will be first encounter
- Using Claude Free tier — paste this file at session start to restore context

---

## Open Questions / Decisions Pending

- Fyne vs Wails for the overlay UI?
- Always-on-top on Windows: Win32 API directly or via UI framework?
- Gothic 2 runs in DirectX — screenshot approach: GDI, DWM, or third-party lib?
- YAML vs JSON for configs? (YAML preferred for human readability)
- Progress storage: simple JSON file or SQLite?

---

## Session Log

| Date | What was done |
|------|---------------|
| 2026-06-14 | Brainstorming session — stack decided, project named, README + CONTEXT created |

> **Update this table at the end of every session!**
