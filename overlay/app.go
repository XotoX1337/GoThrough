package overlay

import (
	"context"
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/XotoX1337/GoThrough/engine"
	"github.com/XotoX1337/GoThrough/settings"
)

// stepChangedEvent is emitted to the frontend whenever the active step changes
// via a global hotkey (button-driven changes return the new state directly).
const stepChangedEvent = "step:changed"

// StepInfo is the data shape sent to the Wails frontend.
type StepInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Current     int    `json:"current"`
	Total       int    `json:"total"`
	IsFirst     bool   `json:"isFirst"`
	IsLast      bool   `json:"isLast"`
}

// MetaInfo describes the loaded walkthrough for the HUD header.
type MetaInfo struct {
	Game  string `json:"game"`
	Title string `json:"title"`
}

// App is the Go backend exposed to the frontend via Wails bindings.
//
// Engine access is guarded by mu because step changes arrive from two
// goroutines: the WebView thread (frontend-bound method calls) and the global
// hotkey listener (see hotkeys.go).
type App struct {
	mu  sync.Mutex
	eng *engine.Engine
	ctx context.Context // set in OnStartup; nil until the window is up

	set     *settings.Store
	hotkeys *hotkeyManager // set in OnStartup, once the window (and ctx) exist
}

// Meta returns the walkthrough header info (game + title).
func (a *App) Meta() MetaInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	return MetaInfo{Game: a.eng.Game(), Title: a.eng.Title()}
}

// Steps returns every step in the walkthrough so the HUD can render its checklist.
func (a *App) Steps() []StepInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	steps := a.eng.Steps()
	total := len(steps)
	out := make([]StepInfo, total)
	for i, s := range steps {
		out[i] = StepInfo{
			Title:       s.Title,
			Description: s.Description,
			Current:     i + 1,
			Total:       total,
			IsFirst:     i == 0,
			IsLast:      i == total-1,
		}
	}
	return out
}

func (a *App) CurrentStep() StepInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.stepInfo()
}

func (a *App) Next() StepInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	_ = a.eng.Next()
	return a.stepInfo()
}

func (a *App) Prev() StepInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	_ = a.eng.Prev()
	return a.stepInfo()
}

// Goto jumps to a 0-based step index (used by checklist row clicks).
func (a *App) Goto(index int) StepInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	_ = a.eng.Goto(index)
	return a.stepInfo()
}

// Settings returns the current user settings for the HUD's settings panel.
func (a *App) Settings() settings.Settings {
	return a.set.Get()
}

// SaveHotkeys validates, persists, and re-registers a new set of hotkey
// bindings. It returns the stored settings on success so the frontend can
// re-render; an invalid binding (unknown key/modifier) is rejected and the
// previous bindings stay in effect. Called from the HUD settings panel.
func (a *App) SaveHotkeys(hk settings.Hotkeys) (settings.Settings, error) {
	if err := validateHotkeys(hk); err != nil {
		return a.set.Get(), err
	}

	ns := a.set.Get()
	ns.Hotkeys = hk
	if err := a.set.Save(ns); err != nil {
		return a.set.Get(), fmt.Errorf("saving settings: %w", err)
	}

	a.mu.Lock()
	hm := a.hotkeys
	a.mu.Unlock()
	if hm != nil {
		hm.apply(hk)
	}
	return ns, nil
}

// validateHotkeys checks that every binding resolves to a real key/modifier
// combination before it is persisted or registered.
func validateHotkeys(hk settings.Hotkeys) error {
	for name, b := range map[string]settings.Binding{
		"next": hk.Next, "prev": hk.Prev, "toggleHide": hk.ToggleHide, "quit": hk.Quit,
	} {
		var err error
		if b.IsMouse() {
			_, _, err = resolveMouse(b)
		} else {
			_, _, err = resolve(b)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

// FitWindow shrink-wraps the OS window to the given content size (logical px),
// keeping the window's current top-right corner fixed. This confines the
// translucent window backdrop to the panel (no surrounding frame, no
// click-dead-zone), makes the panel grow left/down on resize, and leaves a
// user-moved window where they dragged it. Called by the frontend after any
// layout change.
func (a *App) FitWindow(width, height int) {
	a.mu.Lock()
	ctx := a.ctx
	a.mu.Unlock()
	if ctx == nil || width < 1 || height < 1 {
		return
	}
	x, y := runtime.WindowGetPosition(ctx)
	curW, _ := runtime.WindowGetSize(ctx)
	right := x + curW
	runtime.WindowSetSize(ctx, width, height)
	runtime.WindowSetPosition(ctx, right-width, y)
}

// next/prev are the hotkey-driven counterparts: they mutate the engine and
// push the new state to the frontend via an event (no return value, since the
// caller is Go, not JS).
func (a *App) next() { a.advance((*engine.Engine).Next) }
func (a *App) prev() { a.advance((*engine.Engine).Prev) }

func (a *App) advance(move func(*engine.Engine) error) {
	a.mu.Lock()
	_ = move(a.eng)
	info := a.stepInfo()
	ctx := a.ctx
	a.mu.Unlock()
	if ctx != nil {
		runtime.EventsEmit(ctx, stepChangedEvent, info)
	}
}

// stepInfo builds the StepInfo for the active step. Caller must hold a.mu.
func (a *App) stepInfo() StepInfo {
	current, total := a.eng.Progress()
	step := a.eng.Current()
	return StepInfo{
		Title:       step.Title,
		Description: step.Description,
		Current:     current,
		Total:       total,
		IsFirst:     current == 1,
		IsLast:      a.eng.Done(),
	}
}
