package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Check for overlay mode:    HeartRateOverlay.exe --overlay <state-file-path>
	if len(os.Args) >= 3 && os.Args[1] == "--overlay" {
		runOverlay(os.Args[2])
		return
	}

	// Normal mode: settings window
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "HeartRateOverlay",
		Width:     800,
		Height:    600,
		MinWidth:  500,
		MinHeight: 400,
		Frameless:       true,
		AlwaysOnTop:     false,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func runOverlay(statePath string) {
	overlayApp := NewOverlayApp(statePath)

	err := wails.Run(&options.App{
		Title:     "HR Overlay",
		Width:     300,
		Height:    200,
		MinWidth:  100,
		MinHeight: 60,
		Frameless:       true,
		AlwaysOnTop:     true,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			BackdropType:         windows.None,
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  overlayApp.startup,
		OnShutdown: overlayApp.shutdown,
		Bind: []interface{}{
			overlayApp,
		},
	})

	if err != nil {
		println("Overlay error:", err.Error())
	}
}
