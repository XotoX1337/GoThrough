// Package overlay provides the Wails-based HUD window for GoThrough.
package overlay

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/XotoX1337/GoThrough/engine"
)

// Overlay window geometry. The window is a fixed corner panel (not fullscreen)
// so the game stays clickable everywhere outside it; it is anchored to the
// top-right of the primary screen on startup.
const (
	overlayWidth  = 340
	overlayHeight = 480
	screenMargin  = 12
)

//go:embed frontend
var assets embed.FS

// Overlay is the Wails HUD window.
type Overlay struct {
	app     *App
	hotkeys *hotkeyManager
}

// New creates an Overlay for the given engine.
func New(eng *engine.Engine) *Overlay {
	return &Overlay{app: &App{eng: eng}}
}

// Run opens the overlay window and blocks until the user closes it.
func (o *Overlay) Run() error {
	return wails.Run(&options.App{
		Title:         "GoThrough — " + o.app.eng.Title(),
		Width:         overlayWidth,
		Height:        overlayHeight,
		DisableResize: true,
		Frameless:     true,
		AlwaysOnTop:   true,
		OnStartup:     o.onStartup,
		OnShutdown:    o.onShutdown,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		Bind:             []interface{}{o.app},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
	})
}

// onStartup captures the Wails runtime context (needed to emit events and
// drive window visibility), anchors the window to the top-right of the screen,
// and starts the global hotkey listeners.
func (o *Overlay) onStartup(ctx context.Context) {
	o.app.ctx = ctx
	anchorTopRight(ctx)
	o.hotkeys = newHotkeyManager(ctx, o.app)
	o.hotkeys.start()
}

// anchorTopRight positions the window at the top-right corner of the primary
// screen at its initial size. Wails centres new windows by default, which (with
// our small window) makes the overlay look like a tiny box in the middle of the
// screen. The frontend later shrink-wraps the window to the panel via
// App.FitWindow, which re-anchors using the same margin.
func anchorTopRight(ctx context.Context) {
	if w := primaryScreenWidth(ctx); w > 0 {
		runtime.WindowSetPosition(ctx, w-overlayWidth-screenMargin, screenMargin)
	}
}

// primaryScreenWidth returns the logical-pixel width of the primary screen, or
// 0 if it can't be determined.
func primaryScreenWidth(ctx context.Context) int {
	screens, err := runtime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return 0
	}
	screen := screens[0]
	for _, s := range screens {
		if s.IsPrimary {
			screen = s
			break
		}
	}
	if screen.Size.Width != 0 {
		return screen.Size.Width
	}
	return screen.Width // fall back to the deprecated field if Size is unset
}

// onShutdown releases the global hotkeys when the window closes.
func (o *Overlay) onShutdown(_ context.Context) {
	if o.hotkeys != nil {
		o.hotkeys.stop()
	}
}
