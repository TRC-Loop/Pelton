// Package themepack implements the .peltontheme container format: a zip
// holding a manifest.json that declares token override files, css files, icon
// overrides and assets. The package parses and validates containers for
// import, installs them as extracted folders (so a theme stays hand-editable
// on disk), loads installed themes for applying, and zips a folder back up
// for export. Everything security-relevant (path safety, size caps, token
// allowlisting, css remote-reference scanning, svg sanitizing) lives here so
// the bind layer and the frontend never see unvalidated theme content.
//
// File map: manifest.go (manifest schema + parsing), tokens.go (allowlist +
// value checks), css.go (remote-reference scan/strip), svg.go (icon
// sanitizing), version.go (app version range warnings), container.go (zip
// reading), install.go (install/load/export on disk).
package themepack

// ManifestVersion is the container format version this engine understands. A
// container declaring a higher version is refused outright: an unknown format
// cannot degrade gracefully, unlike an app-version mismatch which is only a
// warning (see CompatWarning).
const ManifestVersion = 1

// size caps for a container and its parts. generous for themes, small enough
// that a hostile zip cannot balloon memory or disk.
const (
	maxContainerBytes = 20 << 20  // compressed container file
	maxTotalBytes     = 50 << 20  // sum of uncompressed entries
	maxEntryBytes     = 10 << 20  // any single uncompressed entry
	maxCSSTotalBytes  = 1 << 20   // all css files together
	maxSVGBytes       = 256 << 10 // one icon svg
	maxAssetInline    = 5 << 20   // one asset inlined as a data: uri
	maxZipEntries     = 512
	maxTokenValueLen  = 300
)
