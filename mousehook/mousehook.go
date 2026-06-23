// Package mousehook provides global mouse-button hotkeys — the mechanism Win32
// RegisterHotKey (used by golang.design/x/hotkey for keyboard hotkeys) does not
// cover. It is the mouse-side sibling of the keyboard hotkey path: a binding is
// a set of modifiers plus one mouse button, and a matching press fires a
// callback even while another window has focus.
//
// Platform support: Windows (low-level mouse hook, WH_MOUSE_LL) and Linux/X11
// (XGrabButton). Other platforms (and Linux under Wayland) return ErrUnsupported
// — the same global-hotkey limitation the keyboard side already has there.
//
// Suppression semantics mirror the keyboard side: only the *trigger* (the mouse
// button) of a matched combo is swallowed so the focused app doesn't also see
// it; the modifier keys always pass through. A button press that doesn't match
// any registered binding (e.g. the bare button, or a different modifier set) is
// never swallowed. The match is exact on modifiers.
//
// The package is self-contained — it depends on neither the settings nor the
// overlay package nor golang.design/x/hotkey — so it can be reasoned about and
// (on its supported platforms) built in isolation.
package mousehook

import "errors"

// ErrUnsupported is returned by Start on platforms without a global mouse-hotkey
// implementation (anything but Windows and Linux/X11).
var ErrUnsupported = errors.New("mousehook: global mouse hotkeys unsupported on this platform")

// Button identifies a physical mouse button.
type Button int

const (
	ButtonLeft Button = 1 + iota
	ButtonRight
	ButtonMiddle
	ButtonX1 // first side button ("back")
	ButtonX2 // second side button ("forward")
)

// Modifier is a keyboard modifier flag; values are OR'd into a mask so a
// binding's required modifiers can be compared for an exact match.
type Modifier uint8

const (
	ModCtrl Modifier = 1 << iota
	ModAlt
	ModShift
	ModWin
)

// binding is one registered mouse hotkey.
type binding struct {
	mods Modifier
	btn  Button
	act  func()
}

// core holds the platform-independent binding list. Each platform's Manager
// embeds it and adds the OS-specific machinery (hook/grab + message loop).
type core struct {
	bindings []binding
}

// Add registers a mouse hotkey. Call before Start; bindings added after Start
// are ignored until the next Start. mods is the OR of the required Modifier
// flags (0 for none).
func (c *core) Add(mods Modifier, btn Button, act func()) {
	c.bindings = append(c.bindings, binding{mods: mods, btn: btn, act: act})
}

// bit returns the swallow-tracking bit for a button.
func bit(b Button) uint8 { return 1 << uint(b) }
