package themepack

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// VersionRange is the app version span a theme declares itself made for. Both
// ends are optional; an empty range means "any". Being outside the range is a
// warning, never a block.
type VersionRange struct {
	Min string `json:"min,omitempty"`
	Max string `json:"max,omitempty"`
}

// Manifest is the parsed manifest.json of a container. Only ManifestVersion,
// Name and Base are required; everything else has sensible zero behavior.
type Manifest struct {
	ManifestVersion int           `json:"manifestVersion"`
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Author          string        `json:"author"`
	Version         string        `json:"version"`
	Description     string        `json:"description"`
	Homepage        string        `json:"homepage"`
	License         string        `json:"license"`
	Base            string        `json:"base"`
	Pelton          *VersionRange `json:"pelton,omitempty"`
	// Tokens is either a list of token-file paths (merged in order, later
	// wins) or a single inline token object for one-file themes.
	Tokens json.RawMessage `json:"tokens,omitempty"`
	// CSS files are concatenated in this order into one injected stylesheet.
	CSS     []string          `json:"css,omitempty"`
	Preview string            `json:"preview,omitempty"`
	Icons   map[string]string `json:"icons,omitempty"`
}

// idPattern is the shape of a theme id: a lowercase slug usable as a
// directory name on every platform.
var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// ValidID reports whether id is a well-formed theme id. The bind layer
// checks ids coming from the frontend with this before touching the
// filesystem, so an id can never be a path.
func ValidID(id string) bool {
	return idPattern.MatchString(id)
}

// parseManifest decodes and structurally validates a manifest.json. It does
// not touch referenced files; the container reader does that.
func parseManifest(data []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("manifest.json: %w", err)
	}
	if m.ManifestVersion <= 0 {
		return Manifest{}, fmt.Errorf("manifest.json: missing manifestVersion")
	}
	if m.ManifestVersion > ManifestVersion {
		return Manifest{}, fmt.Errorf("theme format version %d is newer than this Pelton understands (%d)", m.ManifestVersion, ManifestVersion)
	}
	if strings.TrimSpace(m.Name) == "" {
		return Manifest{}, fmt.Errorf("manifest.json: missing name")
	}
	if m.Base != "light" && m.Base != "dark" {
		return Manifest{}, fmt.Errorf("manifest.json: base must be \"light\" or \"dark\"")
	}
	if m.ID == "" {
		m.ID = Slug(m.Name)
	}
	if !idPattern.MatchString(m.ID) {
		return Manifest{}, fmt.Errorf("manifest.json: id %q must be a lowercase slug (a-z, 0-9, dashes)", m.ID)
	}
	return m, nil
}

// tokenPaths returns the manifest's token file list, or nil with inline set
// when the tokens field is a single embedded object instead of paths.
func (m Manifest) tokenPaths() (paths []string, inline map[string]string, err error) {
	if len(m.Tokens) == 0 {
		return nil, nil, nil
	}
	if err = json.Unmarshal(m.Tokens, &paths); err == nil {
		return paths, nil, nil
	}
	if err = json.Unmarshal(m.Tokens, &inline); err == nil {
		return nil, inline, nil
	}
	return nil, nil, fmt.Errorf("manifest.json: tokens must be a list of file paths or an object")
}

// Slug lowercases a name into a valid theme id: runs of anything outside
// a-z0-9 collapse to single dashes.
func Slug(name string) string {
	var b strings.Builder
	lastDash := true // suppress a leading dash
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	s := strings.TrimSuffix(b.String(), "-")
	if s == "" {
		s = "theme"
	}
	if len(s) > 64 {
		s = s[:64]
	}
	return s
}
