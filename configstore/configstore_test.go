package configstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestEmbeddedIndexValid checks that the embedded index.json parses, is
// non-empty, and that every entry carries the fields the picker and the
// update-detection path rely on. genindex builds this file from the real
// configs, so a malformed or stale index fails here loudly rather than leaving
// the picker silently empty.
func TestEmbeddedIndexValid(t *testing.T) {
	var idx Index
	if err := json.Unmarshal(embeddedIndex, &idx); err != nil {
		t.Fatalf("embedded index.json does not parse: %v", err)
	}
	if idx.Schema != 1 {
		t.Errorf("schema = %d, want 1", idx.Schema)
	}
	if len(idx.Configs) == 0 {
		t.Fatal("embedded index has no configs")
	}
	for i, e := range idx.Configs {
		if e.Game == "" {
			t.Errorf("config %d: missing game", i)
		}
		if e.Title == "" {
			t.Errorf("config %d (%s): missing title", i, e.Path)
		}
		if e.Path == "" {
			t.Errorf("config %d: missing path", i)
		}
		if len(e.Hash) != 64 {
			t.Errorf("config %d (%s): hash %q is not a 64-char sha256-hex", i, e.Path, e.Hash)
		}
	}
}

// TestClearCache verifies the cache directory (and everything in it) is removed,
// and that clearing an already-absent cache is not an error. The user-config dir
// is redirected to a temp dir so the real cache is never touched.
func TestClearCache(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("AppData", tmp)         // Windows: os.UserConfigDir uses %AppData%
	t.Setenv("XDG_CONFIG_HOME", tmp) // Linux: os.UserConfigDir uses $XDG_CONFIG_HOME

	dir, err := CacheDir()
	if err != nil {
		t.Fatalf("CacheDir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "gothic2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "gothic2", "day1.yaml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ClearCache(); err != nil {
		t.Fatalf("ClearCache: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("cache dir still present after ClearCache (stat err=%v)", err)
	}

	// A second clear, with the cache already gone, must be a no-op.
	if err := ClearCache(); err != nil {
		t.Fatalf("ClearCache on missing dir: %v", err)
	}
}

// TestListEmbeddedSorted verifies the embedded catalog comes back ordered by
// game then chapter, the order the two-level picker renders in.
func TestListEmbeddedSorted(t *testing.T) {
	entries := ListEmbedded()
	if len(entries) == 0 {
		t.Fatal("ListEmbedded returned nothing")
	}
	for i := 1; i < len(entries); i++ {
		prev, cur := entries[i-1], entries[i]
		if prev.Game > cur.Game || (prev.Game == cur.Game && prev.Chapter > cur.Chapter) {
			t.Errorf("not sorted at %d: %q/%d after %q/%d",
				i, cur.Game, cur.Chapter, prev.Game, prev.Chapter)
		}
	}
}
