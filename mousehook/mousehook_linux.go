//go:build linux

package mousehook

import (
	"errors"
	"fmt"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// Linux/X11 global mouse hotkeys via XGrabButton — the pointer-button analogue
// of XGrabKey (which is how keyboard hotkeys work on X11). Unlike the Windows
// hook, the grab is an atomic modifier+button combo: only the exact combo we
// grab is intercepted, so the bare button (and other modifier sets) reach the
// focused window untouched, and the matched press is delivered to us instead of
// the app — i.e. "swallow only the full combo" falls out for free.
//
// This targets X11 only; under a native Wayland session global grabs don't work
// (the same limitation the X11 keyboard-hotkey backend already has). NOT
// runtime-verified on the Windows dev box — compile-checked via cross-build.

// lockVariants are the lock-key modifier masks we additionally OR into each grab
// so the binding still fires with NumLock/CapsLock engaged. Incoming events are
// masked by these before matching.
var lockVariants = []uint16{
	0,
	xproto.ModMaskLock,                       // CapsLock
	xproto.ModMask2,                          // NumLock (typical)
	xproto.ModMaskLock | xproto.ModMask2,     //
}

const lockMask = xproto.ModMaskLock | xproto.ModMask2

// Manager owns the X connection and the passive button grabs.
type Manager struct {
	core
	conn    *xgb.Conn
	root    xproto.Window
	acts    chan func()
	done    chan struct{}
	started bool
}

func New() *Manager { return &Manager{} }

// Start opens an X connection and grabs each binding's modifier+button combo on
// the root window. A fatal setup failure (no display/screen) is returned and
// nothing is started; per-grab failures (e.g. another client already owns the
// combo) are joined into the returned error but don't stop the others — the
// manager still runs for the grabs that succeeded.
func (m *Manager) Start() error {
	if len(m.bindings) == 0 {
		return nil
	}

	conn, err := xgb.NewConn()
	if err != nil {
		return fmt.Errorf("x11 connect: %w", err)
	}
	m.conn = conn
	m.root = xproto.Setup(conn).DefaultScreen(conn).Root

	var grabErrs []error
	for i := range m.bindings {
		b := &m.bindings[i]
		xbtn := xButton(b.btn)
		if xbtn == 0 {
			grabErrs = append(grabErrs, fmt.Errorf("unsupported button %v", b.btn))
			continue
		}
		base := xMods(b.mods)
		for _, extra := range lockVariants {
			cookie := xproto.GrabButtonChecked(conn, false, m.root,
				uint16(xproto.EventMaskButtonPress),
				xproto.GrabModeAsync, xproto.GrabModeAsync,
				xproto.Window(0), xproto.Cursor(0),
				xbtn, base|extra)
			if err := cookie.Check(); err != nil {
				grabErrs = append(grabErrs, fmt.Errorf("grab button %v mods %#x: %w", b.btn, base|extra, err))
			}
		}
	}

	m.acts = make(chan func(), 16)
	m.done = make(chan struct{})
	go m.worker()
	go m.loop()
	m.started = true

	return errors.Join(grabErrs...)
}

// loop reads X events until the connection is closed by Stop, dispatching
// matching button presses.
func (m *Manager) loop() {
	for {
		ev, err := m.conn.WaitForEvent()
		if ev == nil && err == nil {
			return // connection closed (Stop)
		}
		if err != nil {
			continue
		}
		if bp, ok := ev.(xproto.ButtonPressEvent); ok {
			m.dispatch(bp)
		}
	}
}

// dispatch runs the action for the binding matching a button press. The lock
// modifiers are stripped before comparing, and the match on the remaining
// modifiers is exact.
func (m *Manager) dispatch(e xproto.ButtonPressEvent) {
	masked := e.State &^ lockMask
	for i := range m.bindings {
		b := &m.bindings[i]
		if xButton(b.btn) == byte(e.Detail) && xMods(b.mods) == masked {
			select {
			case m.acts <- b.act:
			default:
			}
			return
		}
	}
}

// worker runs matched actions off the X event goroutine.
func (m *Manager) worker() {
	for {
		select {
		case act := <-m.acts:
			if act != nil {
				act()
			}
		case <-m.done:
			return
		}
	}
}

// Stop releases the grabs and closes the connection. Safe if Start failed.
func (m *Manager) Stop() {
	if m.conn == nil {
		return
	}
	if m.started {
		close(m.done)
		for i := range m.bindings {
			b := &m.bindings[i]
			xbtn := xButton(b.btn)
			if xbtn == 0 {
				continue
			}
			base := xMods(b.mods)
			for _, extra := range lockVariants {
				xproto.UngrabButton(m.conn, xbtn, m.root, base|extra)
			}
		}
		m.started = false
	}
	m.conn.Close() // unblocks WaitForEvent in loop
}

func xButton(b Button) byte {
	switch b {
	case ButtonLeft:
		return 1
	case ButtonMiddle:
		return 2
	case ButtonRight:
		return 3
	case ButtonX1:
		return 8
	case ButtonX2:
		return 9
	}
	return 0
}

func xMods(m Modifier) uint16 {
	var s uint16
	if m&ModCtrl != 0 {
		s |= xproto.ModMaskControl
	}
	if m&ModAlt != 0 {
		s |= xproto.ModMask1
	}
	if m&ModShift != 0 {
		s |= xproto.ModMaskShift
	}
	if m&ModWin != 0 {
		s |= xproto.ModMask4
	}
	return s
}
