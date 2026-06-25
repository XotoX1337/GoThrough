//go:build !windows

package overlay

import "context"

// focusCursor is a no-op on non-Windows platforms. Warping the pointer into the
// overlay (X11 XWarpPointer / Wayland) isn't implemented yet; the keyboard and
// mouse hotkey paths share the same Windows-first limitation (see CLAUDE.md).
func focusCursor(_ context.Context) {}
