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

	"github.com/XotoX1337/GoThrough/configstore"
	"github.com/XotoX1337/GoThrough/engine"
	"github.com/XotoX1337/GoThrough/settings"
)

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

// New creates an Overlay for the given engine and settings store.
// Pass eng=nil to start in config-picker mode.
func New(eng *engine.Engine, set *settings.Store) *Overlay {
	return &Overlay{app: &App{eng: eng, set: set}}
}

// Run opens the overlay window and blocks until the user closes it.
func (o *Overlay) Run() error {
	title := "GoThrough"
	if o.app.eng != nil {
		title = "GoThrough — " + o.app.eng.Title()
	}
	return wails.Run(&options.App{
		Title:         title,
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

func (o *Overlay) onStartup(ctx context.Context) {
	o.app.ctx = ctx
	anchorTopRight(ctx)
	o.hotkeys = newHotkeyManager(ctx, o.app)
	o.app.hotkeys = o.hotkeys
	o.hotkeys.apply(o.app.set.Get().Hotkeys)

	// Background: check for configs in the repo newer than the bundled set.
	go func() {
		entries, err := configstore.FetchNewRemote(ctx)
		if err != nil {
			runtime.LogInfof(ctx, "configstore: remote check failed (offline?): %v", err)
			return
		}
		if len(entries) > 0 {
			runtime.EventsEmit(ctx, "configs:remote", entries)
		}
	}()
}

func anchorTopRight(ctx context.Context) {
	if w := primaryScreenWidth(ctx); w > 0 {
		runtime.WindowSetPosition(ctx, w-overlayWidth-screenMargin, screenMargin)
	}
}

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
	return screen.Width
}

func (o *Overlay) onShutdown(_ context.Context) {
	if o.hotkeys != nil {
		o.hotkeys.stop()
	}
}
