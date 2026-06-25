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
	"sync"
	"time"

	"github.com/XotoX1337/GoThrough/config"
)

// fileVersion tags the on-disk schema so future formats can be migrated.
// v2 added the per-walkthrough Branches map (v0.7); v1 files (no Branches)
// still load — the field simply unmarshals to nil, meaning "no choices yet".
const fileVersion = 2

// record is the saved position for one walkthrough. StepID is kept alongside
// the index so progress survives steps being inserted or removed from the
// config: on restore we prefer to re-find the step by ID and only fall back to
// the raw index. Branches records chosen branch options (persistKey -> option
// label) so a branching walkthrough resumes on the same path.
type record struct {
	StepIndex int               `json:"stepIndex"`
	StepID    int               `json:"stepId"`
	Branches  map[string]string `json:"branches,omitempty"`
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
	return fmt.Sprintf("%s\x1fch%d\x1f%s", wt.Game, wt.Chapter, wt.Title)
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

// Load returns the saved step index, ID, and branch choices for this
// walkthrough. ok is false when no progress has been recorded yet.
func (h *Handle) Load() (index, stepID int, branches map[string]string, ok bool) {
	h.store.mu.Lock()
	defer h.store.mu.Unlock()
	rec, ok := h.store.doc.Entries[h.key]
	return rec.StepIndex, rec.StepID, rec.Branches, ok
}

// Save records the current position and branch choices and atomically rewrites
// the progress file. It satisfies engine.Persister.
func (h *Handle) Save(index, stepID int, branches map[string]string) error {
	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	// Copy the map so later engine mutations can't alias the stored record.
	var b map[string]string
	if len(branches) > 0 {
		b = make(map[string]string, len(branches))
		for k, v := range branches {
			b[k] = v
		}
	}

	h.store.doc.Entries[h.key] = record{
		StepIndex: index,
		StepID:    stepID,
		Branches:  b,
		UpdatedAt: time.Now().UTC(),
	}
	return h.store.writeLocked()
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
