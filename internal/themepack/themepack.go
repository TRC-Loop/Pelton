// Package themepack implements the .peltontheme container format: a zip
// holding a manifest.json that declares token override files, css files, icon
// overrides and assets. Installed themes live as .peltontheme files in the
// user's themes folder; the package parses and validates containers, writes
// them back out (imports, editor saves and the seeded default themes all go
// through WriteContainer), and still reads extracted folders from the
// previous layout. Everything security-relevant (path safety, size caps,
// token allowlisting, css remote-reference scanning, svg sanitizing) lives
// here so the bind layer and the frontend never see unvalidated theme
// content.
//
// File map: manifest.go (manifest schema + parsing), tokens.go (allowlist +
// value checks), css.go (remote-reference scan/strip), svg.go (icon
// sanitizing), version.go (app version range warnings), container.go (zip
// reading), export.go (zip writing), install.go (legacy folder reading),
// presets.go (embedded default themes for seeding).
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
