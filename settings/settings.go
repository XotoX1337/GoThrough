// Package settings persists user-configurable application settings.
//
// State lives in a single JSON file under the user config dir
// (e.g. %AppData%\GoThrough\settings.json on Windows), a sibling of the
// progress file. It mirrors the progress package's mechanics — atomic
// temp-file+rename writes, missing file means defaults — rather than inventing a
// new store. JSON keeps the data CGo-free and human-inspectable.
//
// The package is deliberately ignorant of the hotkey library: a Binding is a
// pair of plain strings (modifier names + key name). Translating those into
// golang.design/x/hotkey constants is the overlay layer's job, which keeps this
// package pure Go and unit-testable without the Wails/CGo toolchain.
package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// fileVersion tags the on-disk schema so future formats can be migrated.
const fileVersion = 1

// Binding is a global hotkey: zero or more modifier names plus exactly one
// trigger — either a keyboard key (Key) or a mouse button (Button), lower-case
// (e.g. {Mods: ["ctrl","alt"], Key: "right"} or {Mods:["ctrl","alt"],
// Button:"middle"}). The names are validated against the hotkey backends by the
// overlay layer, not here. Key and Button are mutually exclusive; a binding with
// Button set is a mouse binding (see IsMouse).
type Binding struct {
	Mods   []string `json:"mods"`
	Key    string   `json:"key,omitempty"`
	Button string   `json:"button,omitempty"`
}

// IsMouse reports whether this binding triggers on a mouse button rather than a
// keyboard key. Mouse bindings are handled by the separate mousehook backend.
func (b Binding) IsMouse() bool { return b.Button != "" }

// Hotkeys holds the binding for each global action the overlay supports.
type Hotkeys struct {
	Next       Binding `json:"next"`
	Prev       Binding `json:"prev"`
	ToggleHide Binding `json:"toggleHide"`
	Quit       Binding `json:"quit"`
}

// LastConfig records the walkthrough that was loaded most recently so the app
// can reopen it on the next launch. An empty Path means "none" — start in the
// config picker. Embedded distinguishes a bundled config (configstore key) from
// a file on disk, mirroring the App.LoadConfig arguments.
type LastConfig struct {
	Path     string `json:"path"`
	Embedded bool   `json:"embedded"`
}

// Settings is the root document written to disk.
type Settings struct {
	Version    int        `json:"version"`
	Hotkeys    Hotkeys    `json:"hotkeys"`
	Opacity    float64    `json:"opacity"`
	LastConfig LastConfig `json:"lastConfig"`
}

// Defaults returns the built-in settings, used when no file exists yet. The
// hotkey defaults equal the combinations the overlay used before rebinding
// existed (Ctrl+Alt+arrows / H / Q).
func Defaults() Settings {
	ctrlAlt := func() []string { return []string{"ctrl", "alt"} }
	return Settings{
		Version: fileVersion,
		Hotkeys: Hotkeys{
			Next:       Binding{Mods: ctrlAlt(), Key: "right"},
			Prev:       Binding{Mods: ctrlAlt(), Key: "left"},
			ToggleHide: Binding{Mods: ctrlAlt(), Key: "h"},
			Quit:       Binding{Mods: ctrlAlt(), Key: "q"},
		},
		Opacity: 1.0,
	}
}

// Store is a file-backed settings document. It is safe for concurrent use:
// settings are read by the hotkey re-registration path and written from the
// HUD, which run on different goroutines.
type Store struct {
	mu   sync.Mutex
	path string
	cur  Settings
}

// DefaultPath returns the standard location of the settings file.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating user config dir: %w", err)
	}
	return filepath.Join(dir, "GoThrough", "settings.json"), nil
}

// Open loads the settings file at path. A missing file is not an error — it
// yields a store seeded with Defaults that will create the file on first save.
// An existing file is layered over the defaults so that fields absent from an
// older or partial file keep their default binding rather than going empty.
func Open(path string) (*Store, error) {
	s := &Store{path: path, cur: Defaults()}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading settings file: %w", err)
	}

	if err := json.Unmarshal(data, &s.cur); err != nil {
		return nil, fmt.Errorf("parsing settings file %s: %w", path, err)
	}
	s.cur.Version = fileVersion
	return s, nil
}

// Get returns a copy of the current settings, safe to read without holding the
// store lock. The Hotkeys value is copied by assignment; the Mods slices are
// shared, so callers must not mutate them in place (they don't).
func (s *Store) Get() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cur
}

// Save replaces the stored settings and atomically rewrites the file (temp file
// + rename) so a crash mid-write can't corrupt the existing settings.
func (s *Store) Save(ns Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ns.Version = fileVersion
	s.cur = ns
	return s.writeLocked()
}

// writeLocked serializes and atomically writes the current document. Caller
// holds mu.
func (s *Store) writeLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("creating settings dir: %w", err)
	}

	data, err := json.MarshalIndent(s.cur, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding settings: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".settings-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp settings file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once the rename succeeds

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp settings file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp settings file: %w", err)
	}

	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("replacing settings file: %w", err)
	}
	return nil
}
