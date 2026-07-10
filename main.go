// Command pelton is the desktop mail client entrypoint. It embeds the built
// frontend and hands control to the desktop package, which owns the wails app
// and all the frontend bindings. Keeping this file tiny keeps the repo root
// uncluttered; the application code lives in internal/desktop.
package main

import (
	"embed"
	"os"
	"slices"

	"github.com/TRC-Loop/Pelton/internal/desktop"
)

//go:embed all:frontend/dist
var assets embed.FS

// licenseManifest is the generated list of third-party licenses (run
// `make licenses`); programLicense is Pelton's own GPL-3.0 text. They are
// embedded here, at the module root where the files live, and handed to the
// desktop layer to serve to the about section on demand.
//
//go:embed licenses/manifest.json
var licenseManifest string

//go:embed LICENSE
var programLicense string

// version is overridden at build time with -ldflags "-X main.version=<v>" (see
// the Makefile) and defaults to "dev".
var version = "dev"

func main() {
	// --potatoes-are-nice launches a purely-cosmetic demo mode used for website
	// screenshots: the ui shows fixed potato-themed sample data instead of real
	// accounts and mail. Nothing else changes.
	demoMode := slices.Contains(os.Args[1:], "--potatoes-are-nice")

	if err := desktop.Run(desktop.Config{
		Assets:          assets,
		Version:         version,
		LicenseManifest: licenseManifest,
		ProgramLicense:  programLicense,
		DemoMode:        demoMode,
	}); err != nil {
		println("Error:", err.Error())
	}
}
