package desktop

import (
	"fmt"
	"os"
	"path/filepath"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/TRC-Loop/Pelton/internal/themepack"
)

// ThemeImportPreviewDTO is the read-before-import view of a chosen theme
// file: its metadata, every stylesheet's raw content and remote references,
// and whether installing would update an already-installed theme. Nothing is
// written anywhere at preview time.
type ThemeImportPreviewDTO struct {
	// Canceled is true when the user dismissed the file dialog.
	Canceled bool         `json:"canceled"`
	Path     string       `json:"path"`
	Info     ThemeInfoDTO `json:"info"`
	// CSSFiles carry the raw stylesheet contents for the read-only viewer.
	CSSFiles []themepack.CSSFile `json:"cssFiles"`
	// UpdatesExisting is true when a theme with the same id is installed;
	// InstalledVersion is that install's version string.
	UpdatesExisting  bool   `json:"updatesExisting"`
	InstalledVersion string `json:"installedVersion"`
}

// PreviewThemeImport opens a file dialog for a .peltontheme container and
// returns everything the import modal shows. The file is parsed and fully
// validated but not installed; ConfirmThemeImport does that after the user
// has seen the css and made the remote-reference choice.
func (a *App) PreviewThemeImport() (ThemeImportPreviewDTO, error) {
	if err := a.ready(); err != nil {
		return ThemeImportPreviewDTO{}, err
	}
	path, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Import theme",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "Pelton themes (*.peltontheme)", Pattern: "*.peltontheme"},
		},
	})
	if err != nil {
		return ThemeImportPreviewDTO{}, err
	}
	if path == "" {
		return ThemeImportPreviewDTO{Canceled: true}, nil
	}
	p, err := a.readContainerFile(path)
	if err != nil {
		return ThemeImportPreviewDTO{}, err
	}
	preview := ThemeImportPreviewDTO{
		Path:     path,
		Info:     a.themeInfo(p),
		CSSFiles: p.CSSFiles,
	}
	if root, err := a.themesDir(); err == nil {
		if installed, err := themepack.LoadInstalled(filepath.Join(root, p.Manifest.ID)); err == nil {
			preview.UpdatesExisting = true
			preview.InstalledVersion = installed.Manifest.Version
		}
	}
	return preview, nil
}

// ConfirmThemeImport installs a previewed container. allowRemote is the
// user's choice from the import warning: false strips every remote reference
// from the css before it touches disk. The file is re-read and re-validated
// here so the preview and the install can never diverge.
func (a *App) ConfirmThemeImport(path string, allowRemote bool) (ThemeInfoDTO, error) {
	if err := a.ready(); err != nil {
		return ThemeInfoDTO{}, err
	}
	p, err := a.readContainerFile(path)
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	root, err := a.themesDir()
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	dir, err := themepack.Install(p, root, !allowRemote)
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	installed, err := themepack.LoadInstalled(dir)
	if err != nil {
		return ThemeInfoDTO{}, err
	}
	return a.themeInfo(installed), nil
}

// readContainerFile reads and parses a .peltontheme file with the container
// size cap enforced before the bytes are even read.
func (a *App) readContainerFile(path string) (*themepack.Package, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > 20<<20 {
		return nil, fmt.Errorf("theme file is larger than 20 MB")
	}
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	return themepack.ReadContainer(data)
}
