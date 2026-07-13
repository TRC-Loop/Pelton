package desktop

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/TRC-Loop/Pelton/internal/themepack"
)

// settingThemeID is the selected custom theme's id; empty means the built-in
// default (light/dark/system per the theme setting).
const settingThemeID = "theme_id"

// ThemeInfoDTO describes one installed theme for the settings gallery.
type ThemeInfoDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Base        string `json:"base"`
	HasCSS      bool   `json:"hasCss"`
	// RemoteRefs are network references still present in the installed css
	// (only non-empty when the user chose Allow at import).
	RemoteRefs []string `json:"remoteRefs"`
	// Preview is the theme's screenshot as a data uri, or "".
	Preview string `json:"preview"`
	// CompatWarning is set when the running app version is outside the range
	// the theme declares itself made for. Informational only.
	CompatWarning string `json:"compatWarning"`
	// Builtin marks a preset shipped inside the app: not on disk, cannot be
	// deleted or exported, always listed first.
	Builtin bool `json:"builtin"`
	// Swatches are a few of the theme's token colors for the gallery card,
	// for themes without a preview screenshot.
	Swatches []string `json:"swatches"`
}

// ThemeApplyDTO is everything the frontend needs to apply a theme: the base
// to fall back on, validated token overrides, the concatenated css with
// bundled assets inlined, and sanitized icon override svgs.
type ThemeApplyDTO struct {
	ID     string            `json:"id"`
	Base   string            `json:"base"`
	Tokens map[string]string `json:"tokens"`
	CSS    string            `json:"css"`
	Icons  map[string]string `json:"icons"`
}

// themesDir returns the themes folder next to the database, creating it on
// first use.
func (a *App) themesDir() (string, error) {
	if a.dataDir == "" {
		return "", errors.New("storage is not available")
	}
	dir := filepath.Join(a.dataDir, "themes")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// themeDirByID resolves an installed theme's folder, rejecting malformed ids
// so an id from the frontend can never traverse paths.
func (a *App) themeDirByID(id string) (string, error) {
	if !themepack.ValidID(id) {
		return "", fmt.Errorf("invalid theme id %q", id)
	}
	root, err := a.themesDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, id)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "", fmt.Errorf("theme %q is not installed", id)
	}
	return dir, nil
}

// themeInfo builds the gallery DTO for a parsed theme. Slice fields must not
// be nil: they cross the json bridge, and the frontend expects arrays.
func (a *App) themeInfo(p *themepack.Package) ThemeInfoDTO {
	m := p.Manifest
	refs := p.RemoteRefs()
	if refs == nil {
		refs = []string{}
	}
	return ThemeInfoDTO{
		ID:            m.ID,
		Name:          m.Name,
		Author:        m.Author,
		Version:       m.Version,
		Description:   m.Description,
		Base:          m.Base,
		HasCSS:        len(m.CSS) > 0,
		RemoteRefs:    refs,
		Preview:       p.PreviewDataURI(),
		CompatWarning: themepack.CompatWarning(m.Pelton, a.version),
		Swatches:      swatchTokens(p.Tokens),
	}
}

// swatchTokens picks a handful of directly renderable token colors for the
// gallery card, in a fixed order. Derived values (rgba, color-mix) are
// skipped; the card only needs a recognizable strip, not the full palette.
func swatchTokens(tokens map[string]string) []string {
	swatches := []string{}
	for _, name := range []string{
		"surface-base", "surface-raised", "text-primary",
		"success", "warning", "danger",
	} {
		if v, ok := tokens[name]; ok && strings.HasPrefix(v, "#") {
			swatches = append(swatches, v)
		}
	}
	return swatches
}

// ListThemes returns every installed theme. Folders that fail to parse are
// skipped (and logged) rather than failing the whole list, so one broken
// hand-edited theme cannot hide the rest.
func (a *App) ListThemes() ([]ThemeInfoDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	root, err := a.themesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	presets := themepack.Presets()
	infos := make([]ThemeInfoDTO, 0, len(presets)+len(entries))
	for _, p := range presets {
		info := a.themeInfo(p)
		info.Builtin = true
		infos = append(infos, info)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := themepack.LoadInstalled(filepath.Join(root, e.Name()))
		if err != nil {
			a.log.Warn("skip unreadable theme", "dir", e.Name(), "err", err)
			continue
		}
		infos = append(infos, a.themeInfo(p))
	}
	return infos, nil
}

// GetThemeApply loads an installed theme in apply form. The frontend calls it
// at startup for the persisted selection and again when the user activates a
// theme.
func (a *App) GetThemeApply(id string) (ThemeApplyDTO, error) {
	if err := a.ready(); err != nil {
		return ThemeApplyDTO{}, err
	}
	if p, ok := themepack.Preset(id); ok {
		return ThemeApplyDTO{
			ID:     p.Manifest.ID,
			Base:   p.Manifest.Base,
			Tokens: p.Tokens,
			CSS:    p.AppliedCSS(),
			Icons:  p.Icons,
		}, nil
	}
	dir, err := a.themeDirByID(id)
	if err != nil {
		return ThemeApplyDTO{}, err
	}
	p, err := themepack.LoadInstalled(dir)
	if err != nil {
		return ThemeApplyDTO{}, err
	}
	return ThemeApplyDTO{
		ID:     p.Manifest.ID,
		Base:   p.Manifest.Base,
		Tokens: p.Tokens,
		CSS:    p.AppliedCSS(),
		Icons:  p.Icons,
	}, nil
}

// DeleteTheme removes an installed theme's folder. If it was the selected
// theme, the selection resets to the built-in default so the next launch does
// not chase a missing folder.
func (a *App) DeleteTheme(id string) error {
	if err := a.ready(); err != nil {
		return err
	}
	if _, ok := themepack.Preset(id); ok {
		return fmt.Errorf("built-in themes cannot be deleted")
	}
	dir, err := a.themeDirByID(id)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if a.stringSetting(settingThemeID, "") == id {
		return a.store.Set(a.ctx, settingThemeID, "")
	}
	return nil
}

// ExportTheme zips an installed theme back into a shareable .peltontheme
// file via a save dialog. Returns the chosen path, or "" if the user
// canceled.
func (a *App) ExportTheme(id string) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	dir, err := a.themeDirByID(id)
	if err != nil {
		return "", err
	}
	p, err := themepack.LoadInstalled(dir)
	if err != nil {
		return "", err
	}
	dest, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		DefaultFilename: themepack.ContainerFileName(p.Manifest.Name),
		Title:           "Export theme",
	})
	if err != nil || dest == "" {
		return "", err
	}
	if err := themepack.Export(dir, dest); err != nil {
		return "", err
	}
	return dest, nil
}
