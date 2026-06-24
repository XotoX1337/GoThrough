package overlay

import "golang.design/x/hotkey"

// Linux/X11 modifier constants for the alt/win names. x/hotkey has no ModAlt /
// ModWin on Linux; the X11 modifier masks are Mod1 (Alt) and Mod4 (Super/Win).
// See bindings.go's modByName.
const (
	modAlt = hotkey.Mod1
	modWin = hotkey.Mod4
)
