package progress

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/XotoX1337/GoThrough/config"
)

func sampleWalkthrough() *config.Walkthrough {
	return &config.Walkthrough{Game: "Gothic 2", Chapter: 1, Title: "Arrival"}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	wt := sampleWalkthrough()

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, _, _, ok := store.For(wt).Load(); ok {
		t.Fatal("expected no saved progress for a fresh store")
	}
	if err := store.For(wt).Save(3, 42, map[string]string{"guild": "A"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Reopen from disk to confirm it persisted.
	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	index, stepID, choices, ok := reopened.For(wt).Load()
	if !ok {
		t.Fatal("expected saved progress after reopen")
	}
	if index != 3 || stepID != 42 {
		t.Fatalf("got index=%d stepID=%d, want 3/42", index, stepID)
	}
	if choices["guild"] != "A" {
		t.Fatalf("got choices=%v, want guild=A", choices)
	}
}

func TestSaveOverwrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	store, _ := Open(path)
	h := store.For(sampleWalkthrough())

	if err := h.Save(1, 10, nil); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if err := h.Save(5, 60, nil); err != nil {
		t.Fatalf("second save: %v", err)
	}
	index, stepID, _, _ := h.Load()
	if index != 5 || stepID != 60 {
		t.Fatalf("got index=%d stepID=%d, want 5/60", index, stepID)
	}
}

func TestSaveCopiesChoiceMap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	store, _ := Open(path)
	h := store.For(sampleWalkthrough())

	c := map[string]string{"guild": "A"}
	if err := h.Save(0, 1, c); err != nil {
		t.Fatalf("save: %v", err)
	}
	c["guild"] = "MUTATED" // must not affect the stored record
	_, _, got, _ := h.Load()
	if got["guild"] != "A" {
		t.Fatalf("stored choices aliased caller's map: got %v", got)
	}
}

func TestLoadsOldFileWithoutChoices(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	// An old document with no "choices" field (nil map is omitted by
	// `omitempty`, reproducing the pre-v0.9 on-disk shape) must still load.
	old, err := json.Marshal(document{
		Version: 1,
		Entries: map[string]record{
			Key(sampleWalkthrough()): {StepIndex: 2, StepID: 30},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, old, 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open old: %v", err)
	}
	index, stepID, choices, ok := store.For(sampleWalkthrough()).Load()
	if !ok || index != 2 || stepID != 30 {
		t.Fatalf("old load: index=%d stepID=%d ok=%v, want 2/30/true", index, stepID, ok)
	}
	if choices != nil {
		t.Fatalf("old load: choices=%v, want nil", choices)
	}
}

func TestKeyDistinguishesWalkthroughs(t *testing.T) {
	a := &config.Walkthrough{Game: "Gothic 2", Chapter: 1, Title: "Arrival"}
	b := &config.Walkthrough{Game: "Gothic 2", Chapter: 2, Title: "Arrival"}
	if Key(a) == Key(b) {
		t.Fatal("different chapters must produce different keys")
	}

	path := filepath.Join(t.TempDir(), "progress.json")
	store, _ := Open(path)
	if err := store.For(a).Save(2, 20, nil); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if _, _, _, ok := store.For(b).Load(); ok {
		t.Fatal("walkthrough b must not see walkthrough a's progress")
	}
}

func TestDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	store, _ := Open(path)
	wt := sampleWalkthrough()
	if err := store.For(wt).Save(3, 30, nil); err != nil {
		t.Fatalf("save: %v", err)
	}

	if err := store.Delete(Key(wt)); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, _, _, ok := store.For(wt).Load(); ok {
		t.Fatal("record still present after Delete")
	}

	// Deleting a missing key is a no-op, not an error.
	if err := store.Delete("nope"); err != nil {
		t.Fatalf("Delete of missing key: %v", err)
	}

	// The deletion must persist across reopen.
	reopened, _ := Open(path)
	if _, _, _, ok := reopened.For(wt).Load(); ok {
		t.Fatal("record came back after reopen")
	}
}

func TestDeleteGame(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	store, _ := Open(path)

	g2ch1 := &config.Walkthrough{Game: "Gothic 2", Chapter: 1, Title: "Arrival"}
	g2ch2 := &config.Walkthrough{Game: "Gothic 2", Chapter: 2, Title: "City"}
	other := &config.Walkthrough{Game: "PoE", Chapter: 1, Title: "Act 1"}
	for _, wt := range []*config.Walkthrough{g2ch1, g2ch2, other} {
		if err := store.For(wt).Save(1, 1, nil); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	if err := store.DeleteGame("Gothic 2"); err != nil {
		t.Fatalf("DeleteGame: %v", err)
	}
	if _, _, _, ok := store.For(g2ch1).Load(); ok {
		t.Fatal("Gothic 2 ch1 still present")
	}
	if _, _, _, ok := store.For(g2ch2).Load(); ok {
		t.Fatal("Gothic 2 ch2 still present")
	}
	if _, _, _, ok := store.For(other).Load(); !ok {
		t.Fatal("DeleteGame removed an unrelated game")
	}
}

func TestDeleteChapter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	store, _ := Open(path)

	ch1 := &config.Walkthrough{Game: "Gothic 2", Chapter: 1, Title: "Arrival"}
	ch2 := &config.Walkthrough{Game: "Gothic 2", Chapter: 2, Title: "City"}
	for _, wt := range []*config.Walkthrough{ch1, ch2} {
		if err := store.For(wt).Save(1, 1, nil); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	// Delete by game + chapter, without knowing the title.
	if err := store.DeleteChapter("Gothic 2", 1); err != nil {
		t.Fatalf("DeleteChapter: %v", err)
	}
	if _, _, _, ok := store.For(ch1).Load(); ok {
		t.Fatal("chapter 1 still present")
	}
	if _, _, _, ok := store.For(ch2).Load(); !ok {
		t.Fatal("DeleteChapter removed chapter 2")
	}
}

func TestClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	store, _ := Open(path)
	store.For(&config.Walkthrough{Game: "A", Chapter: 1, Title: "x"}).Save(1, 1, nil)
	store.For(&config.Walkthrough{Game: "B", Chapter: 1, Title: "y"}).Save(1, 1, nil)

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	reopened, _ := Open(path)
	if _, _, _, ok := reopened.For(&config.Walkthrough{Game: "A", Chapter: 1, Title: "x"}).Load(); ok {
		t.Fatal("records survived Clear")
	}
}

func TestOpenMissingFileIsEmpty(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err != nil {
		t.Fatalf("Open of missing file should not error: %v", err)
	}
	if _, _, _, ok := store.For(sampleWalkthrough()).Load(); ok {
		t.Fatal("missing file should yield empty store")
	}
}

func TestOpenCorruptFileErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); err == nil {
		t.Fatal("expected error opening corrupt progress file")
	}
}
