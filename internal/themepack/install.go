package themepack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadInstalled reads an extracted theme folder back into a validated
// package. Themes normally live as .peltontheme files, but folders from the
// previous layout (and hand-authored folders) keep working; they go through
// the exact same validation as a container, so a corrupted or edited folder
// fails loudly instead of half-applying.
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
