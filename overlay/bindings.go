package overlay

import (
	"fmt"
	"strings"

	"golang.design/x/hotkey"

	"github.com/XotoX1337/GoThrough/mousehook"
	"github.com/XotoX1337/GoThrough/settings"
)

// This file is the bridge between the hotkey-agnostic settings package (plain
// strings) and golang.design/x/hotkey's typed constants. Keeping the name↔const
// mapping here lets settings stay pure Go (no hotkey import, testable without
// the Wails/CGo toolchain) and confines the platform-specific key tables to the
// overlay, which only builds on Windows anyway.

// modByName maps a settings modifier name to its hotkey.Modifier. Names are the
// lower-case tokens stored in settings.json.
var modByName = map[string]hotkey.Modifier{
	"ctrl":  hotkey.ModCtrl,
	"alt":   hotkey.ModAlt,
	"shift": hotkey.ModShift,
	"win":   hotkey.ModWin,
}

// keyByName maps a settings key name to its hotkey.Key. Covers the letters,
// digits, arrows and the handful of named keys a user would reasonably bind a
// global overlay shortcut to.
var keyByName = func() map[string]hotkey.Key {
	m := map[string]hotkey.Key{
		"space":  hotkey.KeySpace,
		"return": hotkey.KeyReturn,
		"escape": hotkey.KeyEscape,
		"delete": hotkey.KeyDelete,
		"tab":    hotkey.KeyTab,
		"left":   hotkey.KeyLeft,
		"right":  hotkey.KeyRight,
		"up":     hotkey.KeyUp,
		"down":   hotkey.KeyDown,
	}
	letters := []hotkey.Key{
		hotkey.KeyA, hotkey.KeyB, hotkey.KeyC, hotkey.KeyD, hotkey.KeyE, hotkey.KeyF,
		hotkey.KeyG, hotkey.KeyH, hotkey.KeyI, hotkey.KeyJ, hotkey.KeyK, hotkey.KeyL,
		hotkey.KeyM, hotkey.KeyN, hotkey.KeyO, hotkey.KeyP, hotkey.KeyQ, hotkey.KeyR,
		hotkey.KeyS, hotkey.KeyT, hotkey.KeyU, hotkey.KeyV, hotkey.KeyW, hotkey.KeyX,
		hotkey.KeyY, hotkey.KeyZ,
	}
	for i, k := range letters {
		m[string(rune('a'+i))] = k
	}
	digits := []hotkey.Key{
		hotkey.Key0, hotkey.Key1, hotkey.Key2, hotkey.Key3, hotkey.Key4,
		hotkey.Key5, hotkey.Key6, hotkey.Key7, hotkey.Key8, hotkey.Key9,
	}
	for i, k := range digits {
		m[string(rune('0'+i))] = k
	}
	for i := 1; i <= 12; i++ {
		m[fmt.Sprintf("f%d", i)] = hotkey.Key(int(hotkey.KeyF1) + (i - 1))
	}
	return m
}()

// mouseModByName maps a settings modifier name to a mousehook.Modifier flag.
// Kept separate from modByName because the two backends use different modifier
// types (hotkey vs mousehook).
var mouseModByName = map[string]mousehook.Modifier{
	"ctrl":  mousehook.ModCtrl,
	"alt":   mousehook.ModAlt,
	"shift": mousehook.ModShift,
	"win":   mousehook.ModWin,
}

// buttonByName maps a settings button name (with a few common aliases) to a
// mousehook.Button.
var buttonByName = map[string]mousehook.Button{
	"left":    mousehook.ButtonLeft,
	"right":   mousehook.ButtonRight,
	"middle":  mousehook.ButtonMiddle,
	"x1":      mousehook.ButtonX1,
	"back":    mousehook.ButtonX1,
	"mouse4":  mousehook.ButtonX1,
	"x2":      mousehook.ButtonX2,
	"forward": mousehook.ButtonX2,
	"mouse5":  mousehook.ButtonX2,
}

// resolveMouse translates a mouse binding into the modifier mask + button the
// mousehook backend wants. Rejects an empty/unknown button or modifier.
func resolveMouse(b settings.Binding) (mods mousehook.Modifier, btn mousehook.Button, err error) {
	for _, name := range b.Mods {
		mod, ok := mouseModByName[strings.ToLower(name)]
		if !ok {
			return 0, 0, fmt.Errorf("unknown modifier %q", name)
		}
		mods |= mod
	}
	btn, ok := buttonByName[strings.ToLower(b.Button)]
	if !ok {
		return 0, 0, fmt.Errorf("unknown mouse button %q", b.Button)
	}
	return mods, btn, nil
}

// comboLabel renders a binding for logging, e.g. "ctrl+alt+right" or
// "ctrl+alt+mouse:middle". It's purely cosmetic (uses the stored names as-is),
// distinct from the frontend's prettier formatter.
func comboLabel(b settings.Binding) string {
	trigger := b.Key
	if b.IsMouse() {
		trigger = "mouse:" + b.Button
	}
	if len(b.Mods) == 0 {
		return trigger
	}
	return strings.Join(b.Mods, "+") + "+" + trigger
}

// resolve translates a settings.Binding into the modifier/key arguments the
// hotkey library wants. It rejects an empty key or any unknown modifier/key name
// so a bad rebind surfaces an error instead of silently registering nothing.
func resolve(b settings.Binding) (mods []hotkey.Modifier, key hotkey.Key, err error) {
	for _, name := range b.Mods {
		mod, ok := modByName[strings.ToLower(name)]
		if !ok {
			return nil, 0, fmt.Errorf("unknown modifier %q", name)
		}
		mods = append(mods, mod)
	}
	key, ok := keyByName[strings.ToLower(b.Key)]
	if !ok {
		return nil, 0, fmt.Errorf("unknown key %q", b.Key)
	}
	return mods, key, nil
}
