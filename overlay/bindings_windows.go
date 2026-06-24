package overlay

import "golang.design/x/hotkey"

// Windows modifier constants for the alt/win names. x/hotkey defines ModAlt and
// ModWin only on Windows (and macOS); see bindings.go's modByName.
const (
	modAlt = hotkey.ModAlt
	modWin = hotkey.ModWin
)
