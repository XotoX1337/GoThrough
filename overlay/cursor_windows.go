//go:build windows

package overlay

import (
	"context"
	"syscall"
	"unsafe"
)

// Win32 plumbing for the "focus overlay" hotkey: teleport the mouse pointer to
// the centre of the overlay window and raise it. Coordinates come from
// GetWindowRect and go straight to SetCursorPos — both are physical screen
// pixels, so the move is DPI-correct without any scaling math.
var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procSetCursorPos        = user32.NewProc("SetCursorPos")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procEnumWindows         = user32.NewProc("EnumWindows")
	procGetWindowThreadPID  = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible     = user32.NewProc("IsWindowVisible")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procAttachThreadInput   = user32.NewProc("AttachThreadInput")
	procBringWindowToTop    = user32.NewProc("BringWindowToTop")
	procShowWindow          = user32.NewProc("ShowWindow")
	procSetActiveWindow     = user32.NewProc("SetActiveWindow")
	procSetFocus            = user32.NewProc("SetFocus")

	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetCurrentProcID   = kernel32.NewProc("GetCurrentProcessId")
	procGetCurrentThreadID = kernel32.NewProc("GetCurrentThreadId")
)

const swRestore = 9 // ShowWindow nCmdShow: restore + activate

type winRect struct{ left, top, right, bottom int32 }

// focusCursor warps the mouse pointer to the centre of the overlay window and
// raises it to the foreground. In games that capture the mouse to steer the
// character the pointer otherwise can't reach the HUD; this teleports it there
// so the user can click. The ctx arg is unused on Windows (we locate our own
// top-level window via the process id) but kept for parity with the stub.
func focusCursor(_ context.Context) {
	hwnd := findOwnTopWindow()
	if hwnd == 0 {
		return
	}
	var r winRect
	if ret, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r))); ret == 0 {
		return
	}
	cx := (r.left + r.right) / 2
	cy := (r.top + r.bottom) / 2
	forceForeground(hwnd)
	procSetCursorPos.Call(uintptr(cx), uintptr(cy))
}

// forceForeground raises hwnd and gives it keyboard focus, working around
// Windows' foreground lock: a background process (we are, while the game is in
// front) is normally not allowed to call SetForegroundWindow — the call is
// ignored and the taskbar button flashes instead. The standard workaround is to
// briefly attach our thread's input queue to the current foreground window's
// thread; while attached the two threads share focus state, so SetForegroundWindow
// is permitted. We detach again immediately.
func forceForeground(hwnd uintptr) {
	fg, _, _ := procGetForegroundWindow.Call()
	if fg == hwnd {
		return // already focused
	}

	fgThread, _, _ := procGetWindowThreadPID.Call(fg, 0)
	ourThread, _, _ := procGetCurrentThreadID.Call()

	attached := false
	if fg != 0 && fgThread != ourThread {
		if ret, _, _ := procAttachThreadInput.Call(fgThread, ourThread, 1); ret != 0 {
			attached = true
		}
	}

	procShowWindow.Call(hwnd, uintptr(swRestore))
	procBringWindowToTop.Call(hwnd)
	procSetForegroundWindow.Call(hwnd)
	procSetActiveWindow.Call(hwnd)
	procSetFocus.Call(hwnd)

	if attached {
		procAttachThreadInput.Call(fgThread, ourThread, 0)
	}
}

// findOwnTopWindow returns the handle of this process's first visible top-level
// window (the overlay). Matching on the process id rather than the window title
// keeps it stable even though the title changes with the loaded walkthrough.
func findOwnTopWindow() uintptr {
	pid, _, _ := procGetCurrentProcID.Call()
	var found uintptr
	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		var wpid uint32
		procGetWindowThreadPID.Call(hwnd, uintptr(unsafe.Pointer(&wpid)))
		if uintptr(wpid) == pid {
			if vis, _, _ := procIsWindowVisible.Call(hwnd); vis != 0 {
				found = hwnd
				return 0 // stop enumerating
			}
		}
		return 1 // continue
	})
	procEnumWindows.Call(cb, 0)
	return found
}
