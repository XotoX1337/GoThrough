package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenMissingFileYieldsDefaults(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err != nil {
		t.Fatalf("Open of missing file should not error: %v", err)
	}
	got := store.Get()
	if got.Hotkeys.Next.Key != "right" || got.Hotkeys.Quit.Key != "q" {
		t.Fatalf("missing file should yield defaults, got %+v", got.Hotkeys)
	}
	if got.Version != fileVersion {
		t.Fatalf("Version = %d, want %d", got.Version, fileVersion)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	ns := Defaults()
	ns.Hotkeys.Next = Binding{Mods: []string{"ctrl", "shift"}, Key: "n"}
	if err := store.Save(ns); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	got := reopened.Get().Hotkeys.Next
	if got.Key != "n" || len(got.Mods) != 2 || got.Mods[0] != "ctrl" || got.Mods[1] != "shift" {
		t.Fatalf("round-trip mismatch: got %+v", got)
	}
}

// A partial file (e.g. written by an older version that only knew some
// bindings) must keep defaults for the fields it omits rather than zeroing them.
func TestOpenPartialFileKeepsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	partial := `{"hotkeys":{"next":{"mods":["ctrl"],"key":"n"}}}`
	if err := os.WriteFile(path, []byte(partial), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	hk := store.Get().Hotkeys
	if hk.Next.Key != "n" {
		t.Fatalf("explicit binding lost: %+v", hk.Next)
	}
	if hk.Prev.Key != "left" || hk.Quit.Key != "q" {
		t.Fatalf("omitted bindings should retain defaults, got prev=%+v quit=%+v", hk.Prev, hk.Quit)
	}
}

func TestOpenCorruptFileErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); err == nil {
		t.Fatal("expected error opening corrupt settings file")
	}
}
