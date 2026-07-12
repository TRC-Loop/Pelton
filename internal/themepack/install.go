package themepack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Install writes a validated package to themesRoot/<id>, replacing any
// existing install of that id (which is how updates work). When blockRemote
// is set - the user's choice at the import warning - every css file is
// written with its remote references stripped, so what is on disk is exactly
// what will run; nothing re-decides at apply time. Returns the install dir.
func Install(p *Package, themesRoot string, blockRemote bool) (string, error) {
	dir := filepath.Join(themesRoot, p.Manifest.ID)
	if err := os.RemoveAll(dir); err != nil {
		return "", err
	}
	cssPaths := make(map[string]bool, len(p.CSSFiles))
	for _, f := range p.CSSFiles {
		cssPaths[f.Path] = true
	}
	for name, content := range p.Files {
		if blockRemote && cssPaths[name] {
			content = []byte(StripRemote(string(content)))
		}
		dest := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			return "", err
		}
		if err := os.WriteFile(dest, content, 0o600); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// LoadInstalled reads an installed theme folder back into a validated
// package. It goes through the exact same validation as an import, so a
// hand-edited folder (the authoring loop) gets the same safety checks, and a
// file corrupted after install fails loudly instead of half-applying.
func LoadInstalled(dir string) (*Package, error) {
	files, err := readDirEntries(dir)
	if err != nil {
		return nil, err
	}
	manifestData, ok := files["manifest.json"]
	if !ok {
		return nil, fmt.Errorf("%s has no manifest.json", dir)
	}
	m, err := parseManifest(manifestData)
	if err != nil {
		return nil, err
	}
	p := &Package{Manifest: m, Files: files}
	if err := p.loadTokens(); err != nil {
		return nil, err
	}
	if err := p.loadCSS(); err != nil {
		return nil, err
	}
	if err := p.loadIcons(); err != nil {
		return nil, err
	}
	return p, nil
}

// readDirEntries reads a theme folder into the same path-keyed map the zip
// reader produces, enforcing the same caps.
func readDirEntries(dir string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	var total int64
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Size() > maxEntryBytes {
			return fmt.Errorf("%s is larger than %d MB", rel, maxEntryBytes>>20)
		}
		if total += info.Size(); total > maxTotalBytes {
			return fmt.Errorf("theme folder larger than %d MB", maxTotalBytes>>20)
		}
		if len(files) >= maxZipEntries {
			return fmt.Errorf("theme folder has more than %d files", maxZipEntries)
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		files[strings.ReplaceAll(rel, string(filepath.Separator), "/")] = content
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
