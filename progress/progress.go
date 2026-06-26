// Package progress persists how far the user has advanced in each walkthrough.
//
// State lives in a single JSON file under the user config dir
// (e.g. %AppData%\GoThrough\progress.json on Windows). JSON is chosen over
// SQLite to stay CGo-free — this project keeps its only CGo dependencies
// isolated to the capture package — and because the data is tiny and worth
// keeping human-inspectable.
package progress

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/XotoX1337/GoThrough/config"
)

// keySep separates the game, chapter and title segments of a progress key.
// It is a control character so it can't collide with text in those fields.
const keySep = "\x1f"

// fileVersion tags the on-disk schema so future formats can be migrated.
// v3 (v0.9) replaced the old per-walkthrough Branches map with Choices
// (choiceKey -> option value); v2's "branches" key is simply ignored on load.
const fileVersion = 3

// record is the saved position for one walkthrough. StepID is kept alongside
// the index so progress survives steps being inserted or removed from the
// config: on restore we prefer to re-find the step by ID and only fall back to
// the raw index. Choices records answered choices (choice key -> option value)
// so a walkthrough with conditional steps resumes on the same path.
type record struct {
	StepIndex int               `json:"stepIndex"`
	StepID    int               `json:"stepId"`
	Choices   map[string]string `json:"choices,omitempty"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// document is the root JSON structure written to disk.
type document struct {
	Version int               `json:"version"`
	Entries map[string]record `json:"entries"`
}

// Store is a file-backed collection of per-walkthrough progress records. It is
// safe for concurrent use.
type Store struct {
	mu   sync.Mutex
	path string
	doc  document
}

// DefaultPath returns the standard location of the progress file.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating user config dir: %w", err)
	}
	return filepath.Join(dir, "GoThrough", "progress.json"), nil
}

// Open loads the progress file at path. A missing file is not an error — it
// yields an empty store that will create the file on first save.
func Open(path string) (*Store, error) {
	s := &Store{path: path, doc: document{Version: fileVersion, Entries: map[string]record{}}}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading progress file: %w", err)
	}

	if err := json.Unmarshal(data, &s.doc); err != nil {
		return nil, fmt.Errorf("parsing progress file %s: %w", path, err)
	}
	if s.doc.Entries == nil {
		s.doc.Entries = map[string]record{}
	}
	return s, nil
}

// Key uniquely identifies a walkthrough for progress tracking. It is derived
// from the walkthrough's identity rather than its file path, so progress
// follows the walkthrough even if the config file is moved or renamed.
func Key(wt *config.Walkthrough) string {
	return fmt.Sprintf("%s%sch%d%s%s", wt.Game, keySep, wt.Chapter, keySep, wt.Title)
}

// gameOfKey extracts the game segment (everything before the first separator)
// from a progress key produced by Key.
func gameOfKey(key string) string {
	if i := strings.Index(key, keySep); i >= 0 {
		return key[:i]
	}
	return key
}

// Handle is a per-walkthrough view of a Store. It satisfies engine.Persister.
type Handle struct {
	store *Store
	key   string
}

// For returns a Handle bound to the given walkthrough's progress entry.
func (s *Store) For(wt *config.Walkthrough) *Handle {
	return &Handle{store: s, key: Key(wt)}
}

// Load returns the saved step index, ID, and choice answers for this
// walkthrough. ok is false when no progress has been recorded yet.
func (h *Handle) Load() (index, stepID int, choices map[string]string, ok bool) {
	h.store.mu.Lock()
	defer h.store.mu.Unlock()
	rec, ok := h.store.doc.Entries[h.key]
	return rec.StepIndex, rec.StepID, rec.Choices, ok
}

// Save records the current position and choice answers and atomically rewrites
// the progress file. It satisfies engine.Persister.
func (h *Handle) Save(index, stepID int, choices map[string]string) error {
	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	// Copy the map so later engine mutations can't alias the stored record.
	var c map[string]string
	if len(choices) > 0 {
		c = make(map[string]string, len(choices))
		for k, v := range choices {
			c[k] = v
		}
	}

	h.store.doc.Entries[h.key] = record{
		StepIndex: index,
		StepID:    stepID,
		Choices:   c,
		UpdatedAt: time.Now().UTC(),
	}
	return h.store.writeLocked()
}

// Delete removes the progress record for the given key (see Key) and rewrites
// the file atomically. Deleting a key that isn't present is a no-op.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.doc.Entries[key]; !ok {
		return nil
	}
	delete(s.doc.Entries, key)
	return s.writeLocked()
}

// DeleteGame removes every progress record belonging to the named game (all of
// its chapters). The game is the key segment before the first separator.
func (s *Store) DeleteGame(game string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := false
	for k := range s.doc.Entries {
		if gameOfKey(k) == game {
			delete(s.doc.Entries, k)
			removed = true
		}
	}
	if !removed {
		return nil
	}
	return s.writeLocked()
}

// DeleteChapter removes the progress record(s) for one chapter of a game,
// matching on the game + chapter prefix of the key (the title is not needed).
// This is the CLI's chapter-granularity reset, where only game and chapter are
// known.
func (s *Store) DeleteChapter(game string, chapter int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := fmt.Sprintf("%s%sch%d%s", game, keySep, chapter, keySep)
	removed := false
	for k := range s.doc.Entries {
		if strings.HasPrefix(k, prefix) {
			delete(s.doc.Entries, k)
			removed = true
		}
	}
	if !removed {
		return nil
	}
	return s.writeLocked()
}

// Clear removes every progress record. A no-op (no rewrite) on an empty store.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.doc.Entries) == 0 {
		return nil
	}
	s.doc.Entries = map[string]record{}
	return s.writeLocked()
}

// writeLocked serializes the document and writes it atomically (temp file +
// rename) so a crash mid-write can't corrupt existing progress. Caller holds mu.
func (s *Store) writeLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("creating progress dir: %w", err)
	}

	data, err := json.MarshalIndent(s.doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding progress: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".progress-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp progress file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once the rename succeeds

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp progress file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp progress file: %w", err)
	}

	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("replacing progress file: %w", err)
	}
	return nil
}
