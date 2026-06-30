package desktop

// License information for the about section. The texts are embedded by the root
// entrypoint (main.go) and handed in via Config, so they are available offline
// and are only sent across the bindings boundary when the user actually opens
// the licenses view, keeping them out of the frontend bundle.

// Licenses returns the generated third-party license manifest as json, an array
// of {group, name, license, text}. Empty when `make licenses` has not been run.
func (a *App) Licenses() string {
	if a.licenseManifest == "" {
		return "[]"
	}
	return a.licenseManifest
}

// ProgramLicense returns Pelton's own license text (GPL-3.0).
func (a *App) ProgramLicense() string {
	return a.programLicense
}
