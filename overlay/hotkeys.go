package overlay

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.design/x/hotkey"
)

// Global hotkeys (work even while the game has focus, since they go through
// the OS via RegisterHotKey). Ctrl+Alt is used to avoid clashing with typical
// in-game movement/action keys.
//
//	Ctrl+Alt+Right  → next step
//	Ctrl+Alt+Left   → previous step
//	Ctrl+Alt+H      → toggle overlay visibility
//	Ctrl+Alt+Q      → quit the overlay

// hotkeyManager owns the registered global hotkeys for the overlay's lifetime.
type hotkeyManager struct {
	ctx     context.Context
	app     *App
	keys    []*hotkey.Hotkey
	done    chan struct{}
	visible bool
}

func newHotkeyManager(ctx context.Context, app *App) *hotkeyManager {
	return &hotkeyManager{ctx: ctx, app: app, done: make(chan struct{}), visible: true}
}

func (m *hotkeyManager) start() {
	ctrlAlt := []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModAlt}
	m.register(ctrlAlt, hotkey.KeyRight, m.app.next)
	m.register(ctrlAlt, hotkey.KeyLeft, m.app.prev)
	m.register(ctrlAlt, hotkey.KeyH, m.toggleVisible)
	m.register(ctrlAlt, hotkey.KeyQ, m.quit)
}

func (m *hotkeyManager) quit() { runtime.Quit(m.ctx) }

// register binds one global hotkey and spawns a listener that runs act on each
// keydown until stop() is called. A failed registration (e.g. the combination
// is already taken) is logged and skipped rather than aborting the others.
func (m *hotkeyManager) register(mods []hotkey.Modifier, key hotkey.Key, act func()) {
	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		runtime.LogErrorf(m.ctx, "hotkey: register %v failed: %v", hk, err)
		return
	}
	m.keys = append(m.keys, hk)

	down := hk.Keydown()
	go func() {
		for {
			select {
			case <-down:
				act()
			case <-m.done:
				return
			}
		}
	}()
}

func (m *hotkeyManager) toggleVisible() {
	m.visible = !m.visible
	if m.visible {
		runtime.WindowShow(m.ctx)
	} else {
		runtime.WindowHide(m.ctx)
	}
}

func (m *hotkeyManager) stop() {
	close(m.done)
	for _, hk := range m.keys {
		_ = hk.Unregister()
	}
}
