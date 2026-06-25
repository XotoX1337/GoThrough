// Package overlay provides the Wails-based HUD window for GoThrough.
package overlay

import (
	"context"
	"embed"
	"log"

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

// New creates an Overlay for the given engine and settings store. srcPath is the
// on-disk path the walkthrough was loaded from (CLI `run`), so a `next:` hand-off
// can be resolved relative to it; pass "" in picker mode (the picker sets it via
// LoadConfig). Pass eng=nil to start in config-picker mode.
func New(eng *engine.Engine, set *settings.Store, srcPath string) *Overlay {
	return &Overlay{app: &App{eng: eng, set: set, curPath: srcPath, curEmbedded: false}}
}

// Run opens the overlay window and blocks until the user closes it.
func (o *Overlay) Run() error {
	title := "GoThrough"
	if o.app.eng != nil {
		title = "GoThrough — " + o.app.eng.Title()
	}
	log.Printf("overlay.Run: picker=%v title=%q", o.app.eng == nil, title)
	err := wails.Run(&options.App{
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
	log.Printf("overlay.Run: wails.Run returned err=%v", err)
	return err
}

func (o *Overlay) onStartup(ctx context.Context) {
	log.Println("onStartup: begin")
	o.app.ctx = ctx
	log.Println("onStartup: anchoring window")
	anchorTopRight(ctx)
	log.Println("onStartup: creating hotkey manager")
	o.hotkeys = newHotkeyManager(ctx, o.app)
	o.app.hotkeys = o.hotkeys
	// Reopen the last walkthrough (if any and still valid) before the frontend
	// queries IsPicker(), so the app starts where the user left off. No-op when
	// launched via `run <config>` (a walkthrough is already loaded).
	log.Println("onStartup: restoring last config")
	o.app.restoreLastConfig()
	log.Printf("onStartup: applying hotkeys: %+v", o.app.set.Get().Hotkeys)
	o.hotkeys.apply(o.app.set.Get().Hotkeys)
	log.Println("onStartup: done")

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

func (o *Overlay) onShutdown(_ context.Context) {
	log.Println("onShutdown: called")
	if o.hotkeys != nil {
		o.hotkeys.stop()
	}
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
