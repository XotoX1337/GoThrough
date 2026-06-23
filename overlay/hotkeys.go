package overlay

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.design/x/hotkey"

	"github.com/XotoX1337/GoThrough/mousehook"
	"github.com/XotoX1337/GoThrough/settings"
)

// Global hotkeys (work even while the game has focus, since they go through the
// OS via RegisterHotKey). The actual combinations come from settings (defaults
// are Ctrl+Alt+arrows / H / Q) and can be rebound at runtime via apply, which
// unregisters the old set and registers the new one.
//
//	next        → next step
//	prev        → previous step
//	toggleHide  → toggle overlay visibility
//	quit        → quit the overlay

// hotkeyManager owns the registered global hotkeys for the overlay's lifetime.
// It is safe for concurrent use: apply (HUD thread, on rebind) and stop
// (shutdown) both touch the registered set.
type hotkeyManager struct {
	ctx context.Context
	app *App

	mu      sync.Mutex
	keys    []*hotkey.Hotkey
	mouse   *mousehook.Manager // global mouse-button hotkeys (nil until apply)
	done    chan struct{}      // closed to signal the current listeners to exit
	visible bool
}

func newHotkeyManager(ctx context.Context, app *App) *hotkeyManager {
	return &hotkeyManager{ctx: ctx, app: app, visible: true}
}

// apply (re)registers the global hotkeys from the given bindings. Any previously
// registered hotkeys are released first, so this doubles as both the initial
// registration and the runtime rebind path.
func (m *hotkeyManager) apply(hk settings.Hotkeys) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.releaseLocked()
	m.done = make(chan struct{})
	m.mouse = mousehook.New()

	m.bindLocked("next", hk.Next, m.app.next)
	m.bindLocked("prev", hk.Prev, m.app.prev)
	m.bindLocked("toggleHide", hk.ToggleHide, m.toggleVisible)
	m.bindLocked("quit", hk.Quit, m.quit)

	// One mouse hook serves all mouse bindings; Start is a no-op (no hook
	// installed) when none of the four bindings is a mouse button.
	if err := m.mouse.Start(); err != nil {
		runtime.LogErrorf(m.ctx, "hotkey: mouse backend: %v", err)
	}
}

// bindLocked routes a binding to the keyboard or mouse backend. Caller holds mu.
func (m *hotkeyManager) bindLocked(name string, b settings.Binding, act func()) {
	if b.IsMouse() {
		m.registerMouseLocked(name, b, act)
		return
	}
	m.registerLocked(name, b, act)
}

// registerMouseLocked adds a mouse binding to the (not-yet-started) mouse
// manager. Caller holds mu.
func (m *hotkeyManager) registerMouseLocked(name string, b settings.Binding, act func()) {
	mods, btn, err := resolveMouse(b)
	if err != nil {
		runtime.LogErrorf(m.ctx, "hotkey: invalid mouse binding for %q %+v: %v", name, b, err)
		return
	}
	m.mouse.Add(mods, btn, act)
	runtime.LogInfof(m.ctx, "hotkey: registered %q → %s", name, comboLabel(b))
}

func (m *hotkeyManager) quit() { runtime.Quit(m.ctx) }

// registerLocked binds one global hotkey and spawns a listener that runs act on
// each keydown until the current done channel is closed. A binding that can't be
// resolved (unknown key/modifier) or registered (e.g. the combination is already
// taken) is logged and skipped rather than aborting the others. Caller holds mu.
func (m *hotkeyManager) registerLocked(name string, b settings.Binding, act func()) {
	mods, key, err := resolve(b)
	if err != nil {
		runtime.LogErrorf(m.ctx, "hotkey: invalid binding for %q %+v: %v", name, b, err)
		return
	}

	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		runtime.LogErrorf(m.ctx, "hotkey: register %q (%s) failed: %v", name, comboLabel(b), err)
		return
	}
	m.keys = append(m.keys, hk)
	// Logged so a user can confirm, after a restart, that the bindings loaded
	// from settings.json were actually registered (see cmd/run.go console output).
	runtime.LogInfof(m.ctx, "hotkey: registered %q → %s", name, comboLabel(b))

	down := hk.Keydown()
	done := m.done
	go func() {
		for {
			select {
			case <-down:
				act()
			case <-done:
				return
			}
		}
	}()
}

func (m *hotkeyManager) toggleVisible() {
	m.mu.Lock()
	m.visible = !m.visible
	visible := m.visible
	m.mu.Unlock()
	if visible {
		runtime.WindowShow(m.ctx)
	} else {
		runtime.WindowHide(m.ctx)
	}
}

// releaseLocked stops the current listeners and unregisters every hotkey. Caller
// holds mu. Safe to call when nothing is registered yet.
func (m *hotkeyManager) releaseLocked() {
	if m.done != nil {
		close(m.done)
		m.done = nil
	}
	for _, hk := range m.keys {
		_ = hk.Unregister()
	}
	m.keys = nil
	if m.mouse != nil {
		m.mouse.Stop()
		m.mouse = nil
	}
}

func (m *hotkeyManager) stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.releaseLocked()
}
