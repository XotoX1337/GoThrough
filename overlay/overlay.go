// Package overlay provides the Wails-based HUD window for GoThrough.
package overlay

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/XotoX1337/GoThrough/engine"
)

//go:embed frontend
var assets embed.FS

// Overlay is the Wails HUD window.
type Overlay struct {
	app   *App
	title string
}

// New creates an Overlay for the given engine.
func New(eng *engine.Engine, walkthroughTitle string) *Overlay {
	return &Overlay{
		app:   &App{eng: eng},
		title: walkthroughTitle,
	}
}

// Run opens the overlay window and blocks until the user closes it.
func (o *Overlay) Run() error {
	return wails.Run(&options.App{
		Title:         "GoThrough — " + o.title,
		Width:         440,
		Height:        220,
		DisableResize: true,
		Frameless:     true,
		AlwaysOnTop:   false, // enabled in v0.3
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
