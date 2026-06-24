//go:build windows

package mousehook

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows global mouse hotkeys via a low-level mouse hook (WH_MOUSE_LL). The
// hook callback runs system-wide before the focused app sees the event, so it
// can both detect and (for a matched combo) swallow the button. Modifier state
// is read at click time with GetAsyncKeyState — RegisterHotKey-style atomic
// modifier+button combos don't exist for the mouse on Windows.

const (
	whMouseLL = 14
	hcAction  = 0

	wmQuit = 0x0012

	wmLButtonDown = 0x0201
	wmLButtonUp   = 0x0202
	wmRButtonDown = 0x0204
	wmRButtonUp   = 0x0205
	wmMButtonDown = 0x0207
	wmMButtonUp   = 0x0208
	wmXButtonDown = 0x020B
	wmXButtonUp   = 0x020C

	vkShift   = 0x10
	vkControl = 0x11
	vkMenu    = 0x12 // Alt
	vkLWin    = 0x5B
	vkRWin    = 0x5C

	xbutton1 = 0x0001
	xbutton2 = 0x0002
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procSetWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procGetMessage          = user32.NewProc("GetMessageW")
	procPostThreadMessage   = user32.NewProc("PostThreadMessageW")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")

	kernel32               = windows.NewLazySystemDLL("kernel32.dll")
	procGetCurrentThreadId = kernel32.NewProc("GetCurrentThreadId")
)

// msllHookStruct is the MSLLHOOKSTRUCT passed to the low-level mouse hook.
type msllHookStruct struct {
	pt          struct{ x, y int32 }
	mouseData   uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

type winMsg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct{ x, y int32 }
}

// activeManager is the manager whose hook is currently installed. Only one runs
// at a time (the overlay holds a single hotkey manager); the C callback reads it
// lock-free. swallowedMask, written only from the single hook thread, lives on
// the manager and needs no separate synchronisation.
var activeManager atomic.Pointer[Manager]

// callbackPtr is the trampoline passed to SetWindowsHookEx, created once.
var callbackPtr = windows.NewCallback(mouseProc)

// Manager owns the installed hook and dispatches matched combos.
type Manager struct {
	core
	hook         uintptr
	threadID     uint32
	swallowedMask uint8       // buttons whose button-down we swallowed (hook thread only)
	acts         chan func()  // matched callbacks, run off the hook thread
	stopWorker   chan struct{}
	started      bool
}

// New creates an inactive manager.
func New() *Manager { return &Manager{} }

// Start installs the hook and begins intercepting. With no bindings it is a
// no-op (no hook installed — zero overhead and no system-wide footprint until a
// mouse hotkey actually exists).
func (m *Manager) Start() error {
	if len(m.bindings) == 0 {
		return nil
	}
	m.acts = make(chan func(), 16)
	m.stopWorker = make(chan struct{})
	go m.worker()

	ready := make(chan error, 1)
	go m.loop(ready)
	if err := <-ready; err != nil {
		close(m.stopWorker)
		return err
	}
	m.started = true
	return nil
}

// loop installs the hook on a dedicated, locked OS thread and runs the message
// pump the low-level hook requires, until Stop posts WM_QUIT.
func (m *Manager) loop(ready chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	tid, _, _ := procGetCurrentThreadId.Call()
	m.threadID = uint32(tid)

	// hMod may be 0 for a low-level hook (WH_MOUSE_LL): the procedure lives in
	// the current process, not an injected DLL.
	hook, _, callErr := procSetWindowsHookEx.Call(uintptr(whMouseLL), callbackPtr, 0, 0)
	if hook == 0 {
		ready <- fmt.Errorf("SetWindowsHookEx(WH_MOUSE_LL): %v", callErr)
		return
	}
	m.hook = hook
	activeManager.Store(m)
	ready <- nil

	var msg winMsg
	for {
		r, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		// GetMessage returns 0 on WM_QUIT and -1 (^uintptr(0)) on error.
		if r == 0 || r == ^uintptr(0) {
			break
		}
	}
	procUnhookWindowsHookEx.Call(hook)
	activeManager.CompareAndSwap(m, nil)
}

// worker runs matched callbacks off the hook thread, so a slow action can't trip
// the low-level-hook timeout (which would make Windows silently drop the hook).
func (m *Manager) worker() {
	for {
		select {
		case act := <-m.acts:
			if act != nil {
				act()
			}
		case <-m.stopWorker:
			return
		}
	}
}

// Stop uninstalls the hook and stops the worker. Safe to call if Start failed or
// was a no-op.
func (m *Manager) Stop() {
	if m.started && m.threadID != 0 {
		procPostThreadMessage.Call(uintptr(m.threadID), wmQuit, 0, 0)
	}
	if m.stopWorker != nil {
		select {
		case <-m.stopWorker: // already closed
		default:
			close(m.stopWorker)
		}
	}
	m.started = false
}

// mouseProc is the low-level mouse hook callback. It must stay fast: it only
// classifies the event, checks modifier state, and either swallows (return 1) or
// passes the event on. The matched action is handed to the worker goroutine.
//
// lParam is typed unsafe.Pointer (it always points at an MSLLHOOKSTRUCT for
// WH_MOUSE_LL) so reading the struct needs no uintptr→pointer cast.
func mouseProc(nCode uintptr, wParam uintptr, lParam unsafe.Pointer) uintptr {
	m := activeManager.Load()
	if m == nil || int32(nCode) != hcAction {
		return callNext(nCode, wParam, lParam)
	}

	btn, isUp, ok := classify(wParam, lParam)
	if !ok {
		return callNext(nCode, wParam, lParam)
	}

	if isUp {
		// Swallow the button-up iff we swallowed its button-down, so the focused
		// app never sees an unpaired up (which could leave it in a stuck state).
		// Modifiers may already be released by now, so we can't re-match here —
		// the remembered down is the authority.
		if m.swallowedMask&bit(btn) != 0 {
			m.swallowedMask &^= bit(btn)
			return 1
		}
		return callNext(nCode, wParam, lParam)
	}

	mods := currentMods()
	for i := range m.bindings {
		b := &m.bindings[i]
		if b.btn == btn && b.mods == mods { // exact modifier match
			m.swallowedMask |= bit(btn)
			select {
			case m.acts <- b.act:
			default: // worker busy — drop rather than block the hook
			}
			return 1 // swallow the trigger
		}
	}
	return callNext(nCode, wParam, lParam)
}

func callNext(nCode, wParam uintptr, lParam unsafe.Pointer) uintptr {
	r, _, _ := procCallNextHookEx.Call(0, nCode, wParam, uintptr(lParam))
	return r
}

// classify maps a mouse message to a button and whether it's a button-up. ok is
// false for non-button events (e.g. moves, wheel), which are passed straight
// through.
func classify(wParam uintptr, lParam unsafe.Pointer) (btn Button, isUp bool, ok bool) {
	switch wParam {
	case wmLButtonDown:
		return ButtonLeft, false, true
	case wmLButtonUp:
		return ButtonLeft, true, true
	case wmRButtonDown:
		return ButtonRight, false, true
	case wmRButtonUp:
		return ButtonRight, true, true
	case wmMButtonDown:
		return ButtonMiddle, false, true
	case wmMButtonUp:
		return ButtonMiddle, true, true
	case wmXButtonDown, wmXButtonUp:
		hs := (*msllHookStruct)(lParam)
		b := ButtonX1
		if (hs.mouseData>>16)&0xffff == xbutton2 {
			b = ButtonX2
		}
		return b, wParam == wmXButtonUp, true
	}
	return 0, false, false
}

func currentMods() Modifier {
	var m Modifier
	if keyDown(vkControl) {
		m |= ModCtrl
	}
	if keyDown(vkMenu) {
		m |= ModAlt
	}
	if keyDown(vkShift) {
		m |= ModShift
	}
	if keyDown(vkLWin) || keyDown(vkRWin) {
		m |= ModWin
	}
	return m
}

func keyDown(vk int) bool {
	r, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return r&0x8000 != 0
}

// bit returns the swallow-tracking bit for a button (Windows-only).
func bit(b Button) uint8 { return 1 << uint(b) }
