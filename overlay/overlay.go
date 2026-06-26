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
	log.Println("onStartup: positioning window")
	o.restoreWindowPos(ctx)
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

	// Background: refresh the catalog from the CDN and auto-update any
	// already-cached games whose chapters changed upstream. Pushes the fresh
	// catalog to the frontend when done. The binary is fully usable offline.
	go o.app.refreshUpdates()
}

func (o *Overlay) onShutdown(_ context.Context) {
	log.Println("onShutdown: called")
	if o.hotkeys != nil {
		o.hotkeys.stop()
	}
}

// restoreWindowPos places the window where the user last left it, or anchors it
// to the top-right corner when no position has been saved yet. A saved position
// is clamped to the primary screen so a window saved on a monitor that is no
// longer present can't end up off-screen.
func (o *Overlay) restoreWindowPos(ctx context.Context) {
	pos := o.app.set.Get().WindowPos
	if !pos.Set {
		anchorTopRight(ctx)
		return
	}
	x, y := clampToScreen(ctx, pos.X, pos.Y, overlayWidth, overlayHeight)
	runtime.WindowSetPosition(ctx, x, y)
}

func anchorTopRight(ctx context.Context) {
	if w, _ := primaryScreenSize(ctx); w > 0 {
		runtime.WindowSetPosition(ctx, w-overlayWidth-screenMargin, screenMargin)
	}
}

// clampToScreen keeps a window of the given size (logical px) within the primary
// screen so the overlay stays reachable. It mirrors the maxX/maxY clamp the
// frontend uses while dragging (Wails' Screen exposes only size, not monitor
// origin, so a precise multi-monitor clamp isn't possible — this is the same
// primary-screen approximation the drag handler already relies on).
func clampToScreen(ctx context.Context, x, y, winW, winH int) (int, int) {
	w, h := primaryScreenSize(ctx)
	if w > 0 {
		maxX := w - winW
		if maxX < 0 {
			maxX = 0
		}
		if x < 0 {
			x = 0
		} else if x > maxX {
			x = maxX
		}
	}
	if h > 0 {
		maxY := h - winH
		if maxY < 0 {
			maxY = 0
		}
		if y < 0 {
			y = 0
		} else if y > maxY {
			y = maxY
		}
	}
	return x, y
}

// primaryScreenSize returns the primary screen's logical-pixel size (the unit
// Wails' WindowSetPosition/Size use), or 0,0 if it can't be determined.
func primaryScreenSize(ctx context.Context) (int, int) {
	screens, err := runtime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return 0, 0
	}
	screen := screens[0]
	for _, s := range screens {
		if s.IsPrimary {
			screen = s
			break
		}
	}
	w, h := screen.Size.Width, screen.Size.Height
	if w == 0 {
		w = screen.Width
	}
	if h == 0 {
		h = screen.Height
	}
	return w, h
}
