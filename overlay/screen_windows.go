//go:build windows

package overlay

import (
	"context"
	"unsafe"
)

// Multi-monitor window placement. Wails' runtime exposes only screen *sizes*
// (no monitor origin) and its WindowSetPosition is relative to whichever monitor
// the window currently sits on — so it can't move the overlay across monitors or
// clamp it to the whole desktop. On Windows we go to Win32 directly:
// GetSystemMetrics gives the virtual-screen rect (the union of all monitors) and
// SetWindowPos moves the window in absolute screen pixels. Everything here is
// physical pixels, like GetWindowRect/SetCursorPos in cursor_windows.go, so it
// stays DPI-correct without scaling math.
var (
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procSetWindowPos     = user32.NewProc("SetWindowPos")
)

const (
	smXVirtualScreen  = 76
	smYVirtualScreen  = 77
	smCXVirtualScreen = 78
	smCYVirtualScreen = 79

	swpNoSize     = 0x0001
	swpNoZorder   = 0x0004
	swpNoActivate = 0x0010
)

func sysMetric(i int) int {
	ret, _, _ := procGetSystemMetrics.Call(uintptr(i))
	return int(int32(ret)) // GetSystemMetrics returns a signed C int (can be < 0)
}

// virtualScreenRect returns the bounding rectangle of all monitors. x/y can be
// negative when a monitor sits left of / above the primary. Physical px.
func virtualScreenRect() (x, y, w, h int) {
	return sysMetric(smXVirtualScreen), sysMetric(smYVirtualScreen),
		sysMetric(smCXVirtualScreen), sysMetric(smCYVirtualScreen)
}

// clampToVirtual keeps a w×h window fully inside the virtual screen so it can
// roam across monitors but never hang over the outer edge into nothing.
func clampToVirtual(x, y, w, h int) (int, int) {
	vx, vy, vw, vh := virtualScreenRect()
	if vw > 0 {
		if maxX := vx + vw - w; x > maxX {
			x = maxX
		}
		if x < vx {
			x = vx
		}
	}
	if vh > 0 {
		if maxY := vy + vh - h; y > maxY {
			y = maxY
		}
		if y < vy {
			y = vy
		}
	}
	return x, y
}

// clampToWorkArea clamps a saved position back into reach on restore. The passed
// w/h are the overlay's logical size; on Windows that matches physical px on a
// 100%-DPI monitor and is a fine approximation elsewhere.
func clampToWorkArea(_ context.Context, x, y, w, h int) (int, int) {
	return clampToVirtual(x, y, w, h)
}

// moveOverlayAbs clamps (x,y) to the virtual screen and moves the overlay there
// in absolute coordinates via SetWindowPos, returning the clamped top-left. The
// window's real size is read from GetWindowRect so the clamp uses physical px
// (the passed w/h are ignored on Windows). If the window can't be located yet
// (not shown) it clamps with the default size and skips the move.
func moveOverlayAbs(_ context.Context, x, y, _, _ int) (int, int) {
	hwnd := findOwnTopWindow()
	w, h := overlayWidth, overlayHeight
	if hwnd != 0 {
		var r winRect
		if ret, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r))); ret != 0 {
			w = int(r.right - r.left)
			h = int(r.bottom - r.top)
		}
	}
	x, y = clampToVirtual(x, y, w, h)
	if hwnd != 0 {
		procSetWindowPos.Call(hwnd, 0, uintptr(x), uintptr(y), 0, 0,
			uintptr(swpNoSize|swpNoZorder|swpNoActivate))
	}
	return x, y
}
