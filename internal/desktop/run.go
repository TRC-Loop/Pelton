// Package desktop is the wails application layer: the bound App struct, its
// frontend-facing methods (the bind_*.go files), the dtos, the native menu and
// the runtime event plumbing. main.go in the repo root is a thin entrypoint that
// embeds the built frontend and calls Run; everything else lives here so the
// project root stays clean. No mail, crypto, sync or storage logic lives in this
// package; it all delegates to internal/*.
package desktop

import (
	"embed"
	"os"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailswindows "github.com/wailsapp/wails/v2/pkg/options/windows"
)

// Config carries what the root entrypoint owns: the embedded frontend assets,
// the build version string and the embedded license texts.
type Config struct {
	Assets          embed.FS
	Version         string
	LicenseManifest string
	ProgramLicense  string
	// DemoMode runs the app in the cosmetic screenshot mode (--potatoes-are-nice).
	DemoMode bool
}

// Run constructs and runs the wails application. It returns wails.Run's error.
func Run(cfg Config) error {
	// WebKitGTK's DMA-BUF renderer misbehaves on many Wayland compositors
	// (notably GNOME, which Fedora ships): the webview's GPU surface is not
	// reallocated when the window manager resizes the window, so after a
	// maximise the interface keeps rendering at its old size with the window
	// background showing through the right and bottom. Falling back to the
	// non-DMA-BUF path fixes that. Only set on Linux, and only when the user
	// hasn't already chosen a value, so a distro package or power user can
	// still override it.
	if goruntime.GOOS == "linux" {
		if _, set := os.LookupEnv("WEBKIT_DISABLE_DMABUF_RENDERER"); !set {
			os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
		}
	}

	app := newApp(cfg.Version)
	app.licenseManifest = cfg.LicenseManifest
	app.programLicense = cfg.ProgramLicense
	app.demoMode = cfg.DemoMode

	return wails.Run(&options.App{
		Title:     "Pelton",
		Width:     1280,
		Height:    820,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: cfg.Assets,
		},
		// neutral dark surface so the native window chrome matches the ui before
		// the frontend paints. the real colors come from the css token theme.
		BackgroundColour: &options.RGBA{R: 17, G: 18, B: 20, A: 1},
		// keep the app running when the window is closed, like macos Mail: closing
		// hides the window and background sync continues; the dock icon reopens it,
		// and Quit (Cmd+Q) in the menu actually exits.
		HideWindowOnClose: true,
		OnStartup:         app.startup,
		OnDomReady:        app.domReady,
		OnShutdown:        app.shutdown,
		Menu:              app.buildMenu(),
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   "Pelton",
				Message: "An open-source desktop mail client.\nVersion " + cfg.Version,
			},
		},
		Windows: &wailswindows.Options{
			Theme: wailswindows.SystemDefault,
		},
		Bind: []interface{}{
			app,
		},
	})
}
