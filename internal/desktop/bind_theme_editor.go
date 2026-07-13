package desktop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TRC-Loop/Pelton/internal/themepack"
)

// The palette editor (#57) saves its themes as regular installed theme
// folders: a manifest.json with inline tokens, nothing else. That way an
// edited palette shows up in the gallery, exports as a .peltontheme and
// travels with backups exactly like an imported theme.

// SaveThemeRequest is a palette-editor save: name, light/dark base and the
// token overrides. ID is set when editing an existing installed theme and
// empty when creating a new one.
type SaveThemeRequest struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Base   string            `json:"base"`
	Tokens map[string]string `json:"tokens"`
}

// SaveCustomTheme validates and writes a palette-editor theme, returning its
// gallery info. New themes get an id derived from the name, kept clear of
// presets and existing installs.
func (a *App) SaveCustomTheme(req SaveThemeRequest) (ThemeInfoDTO, error) {
	if err := a.ready(); err != nil {
		return ThemeInfoDTO{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return ThemeInfoDTO{}, fmt.Errorf("the theme needs a name")
	}
	if req.Base != "light" && req.Base != "dark" {
		return ThemeInfoDTO{}, fmt.Errorf("base must be light or dark")
	}
	tokens, err := themepack.ValidateTokens(req.Tokens)
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	root, err := a.themesDir()
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	id := req.ID
	if id == "" {
		id, err = a.newThemeID(root, name)
		if err != nil {
			return ThemeInfoDTO{}, err
		}
	} else if _, err := a.themeDirByID(id); err != nil {
		// editing requires an existing installed theme; presets are not on
		// disk, so their ids fail here too.
		return ThemeInfoDTO{}, err
	}
	tokensJSON, err := json.Marshal(tokens)
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	manifest, err := json.MarshalIndent(themepack.Manifest{
		ManifestVersion: themepack.ManifestVersion,
		ID:              id,
		Name:            name,
		Base:            req.Base,
		Tokens:          tokensJSON,
	}, "", "  ")
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	dir := filepath.Join(root, id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return ThemeInfoDTO{}, err
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), manifest, 0o600); err != nil {
		return ThemeInfoDTO{}, err
	}
	p, err := themepack.LoadInstalled(dir)
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	return a.themeInfo(p), nil
}

// newThemeID derives a fresh theme id from the name, stepping around presets
// and already installed themes with a numeric suffix.
func (a *App) newThemeID(root, name string) (string, error) {
	slug := themepack.Slug(name)
	for i := 1; i <= 100; i++ {
		id := slug
		if i > 1 {
			id = fmt.Sprintf("%s-%d", slug, i)
		}
		if _, ok := themepack.Preset(id); ok {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, id)); os.IsNotExist(err) {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not find a free id for %q", name)
}
