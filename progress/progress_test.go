package progress

import (
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
	if _, _, ok := store.For(wt).Load(); ok {
		t.Fatal("expected no saved progress for a fresh store")
	}
	if err := store.For(wt).Save(3, 42); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Reopen from disk to confirm it persisted.
	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	index, stepID, ok := reopened.For(wt).Load()
	if !ok {
		t.Fatal("expected saved progress after reopen")
	}
	if index != 3 || stepID != 42 {
		t.Fatalf("got index=%d stepID=%d, want 3/42", index, stepID)
	}
}

func TestSaveOverwrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")
	store, _ := Open(path)
	h := store.For(sampleWalkthrough())

	if err := h.Save(1, 10); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if err := h.Save(5, 60); err != nil {
		t.Fatalf("second save: %v", err)
	}
	index, stepID, _ := h.Load()
	if index != 5 || stepID != 60 {
		t.Fatalf("got index=%d stepID=%d, want 5/60", index, stepID)
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
	if err := store.For(a).Save(2, 20); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if _, _, ok := store.For(b).Load(); ok {
		t.Fatal("walkthrough b must not see walkthrough a's progress")
	}
}

func TestOpenMissingFileIsEmpty(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err != nil {
		t.Fatalf("Open of missing file should not error: %v", err)
	}
	if _, _, ok := store.For(sampleWalkthrough()).Load(); ok {
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
