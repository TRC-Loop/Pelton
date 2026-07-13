package desktop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TRC-Loop/Pelton/internal/themepack"
)

// The palette editor (#57) saves its themes as regular .peltontheme files in
// the themes folder, so an edited palette shows up in the gallery, exports
// and travels with backups exactly like an imported theme.

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
// existing themes. Editing keeps everything else the theme carries (author,
// icons, preview, bundled files) and only replaces name, base and tokens.
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

	manifest := themepack.Manifest{ManifestVersion: themepack.ManifestVersion}
	files := map[string][]byte{}
	var previousPath string
	if req.ID == "" {
		id, err := a.newThemeID(root, name)
		if err != nil {
			return ThemeInfoDTO{}, err
		}
		manifest.ID = id
	} else {
		existing, err := a.findTheme(req.ID)
		if err != nil {
			return ThemeInfoDTO{}, err
		}
		manifest = existing.pkg.Manifest
		for f, content := range existing.pkg.Files {
			files[f] = content
		}
		previousPath = existing.path
	}
	manifest.Name = name
	manifest.Base = req.Base
	if manifest.Tokens, err = json.Marshal(tokens); err != nil {
		return ThemeInfoDTO{}, err
	}
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	files["manifest.json"] = manifestJSON

	dest := filepath.Join(root, containerName(manifest.ID))
	err = themepack.WriteContainer(&themepack.Package{Manifest: manifest, Files: files}, dest, false)
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	// an edited legacy folder migrates to a file; drop the folder.
	if previousPath != "" && previousPath != dest {
		if err := os.RemoveAll(previousPath); err != nil {
			a.log.Warn("remove replaced theme", "path", previousPath, "err", err)
		}
	}
	p, err := a.readContainerFile(dest)
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	return a.themeInfo(p), nil
}

// newThemeID derives a fresh theme id from the name, stepping around
// existing themes with a numeric suffix.
func (a *App) newThemeID(root, name string) (string, error) {
	slug := themepack.Slug(name)
	for i := 1; i <= 100; i++ {
		id := slug
		if i > 1 {
			id = fmt.Sprintf("%s-%d", slug, i)
		}
		if _, err := a.findTheme(id); err != nil {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not find a free id for %q", name)
}
