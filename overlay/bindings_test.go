package overlay

import (
	"path/filepath"
	"testing"

	"golang.design/x/hotkey"

	"github.com/XotoX1337/GoThrough/settings"
)

func hasMod(mods []hotkey.Modifier, want hotkey.Modifier) bool {
	for _, m := range mods {
		if m == want {
			return true
		}
	}
	return false
}

func TestResolveDefaults(t *testing.T) {
	mods, key, err := resolve(settings.Defaults().Hotkeys.Next)
	if err != nil {
		t.Fatalf("resolve default Next: %v", err)
	}
	if key != hotkey.KeyRight {
		t.Fatalf("Next key = %v, want KeyRight", key)
	}
	if !hasMod(mods, hotkey.ModCtrl) || !hasMod(mods, hotkey.ModAlt) {
		t.Fatalf("Next mods = %v, want Ctrl+Alt", mods)
	}
}

func TestResolveRejectsBadBindings(t *testing.T) {
	cases := []settings.Binding{
		{Mods: []string{"hyper"}, Key: "right"}, // unknown modifier
		{Mods: []string{"ctrl"}, Key: "f99"},    // unknown key
		{Mods: []string{"ctrl"}, Key: ""},       // empty key
	}
	for _, b := range cases {
		if _, _, err := resolve(b); err == nil {
			t.Fatalf("resolve(%+v) = nil error, want rejection", b)
		}
	}
}

func TestValidateHotkeys(t *testing.T) {
	if err := validateHotkeys(settings.Defaults().Hotkeys); err != nil {
		t.Fatalf("defaults should validate: %v", err)
	}
	bad := settings.Defaults().Hotkeys
	bad.Quit.Key = "nope"
	if err := validateHotkeys(bad); err == nil {
		t.Fatal("an unknown key must fail validation")
	}
}

// TestRebindSurvivesRestart is the "do saved hotkeys still work after a restart"
// check, minus the OS RegisterHotKey call (which needs a desktop session): it
// saves a rebound hotkey, opens the file from scratch as a fresh launch would,
// and confirms the loaded binding still resolves to a registrable combination.
func TestRebindSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	// First run: rebind Next to a non-default combo and persist it.
	store, err := settings.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	ns := store.Get()
	ns.Hotkeys.Next = settings.Binding{Mods: []string{"ctrl", "shift"}, Key: "n"}
	if err := store.Save(ns); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// "Restart": a brand-new store reading the same file, exactly like
	// openSettings() does on the next launch (cmd/run.go).
	reopened, err := settings.Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	mods, key, err := resolve(reopened.Get().Hotkeys.Next)
	if err != nil {
		t.Fatalf("resolve restored binding: %v", err)
	}
	if key != hotkey.KeyN {
		t.Fatalf("restored Next key = %v, want KeyN", key)
	}
	if !hasMod(mods, hotkey.ModCtrl) || !hasMod(mods, hotkey.ModShift) {
		t.Fatalf("restored Next mods = %v, want Ctrl+Shift", mods)
	}
}
