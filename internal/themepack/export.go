package themepack

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Export zips an installed theme folder back into a .peltontheme container
// at dest, so an installed (possibly hand-edited) theme round-trips into a
// shareable file. Entries are written in sorted order for reproducibility.
func Export(dir, dest string) error {
	files, err := readDirEntries(dir)
	if err != nil {
		return err
	}
	if _, ok := files["manifest.json"]; !ok {
		return fmt.Errorf("%s has no manifest.json", dir)
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	out, err := os.OpenFile(filepath.Clean(dest), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	w := zip.NewWriter(out)
	for _, name := range names {
		entry, err := w.Create(name)
		if err != nil {
			out.Close()
			return err
		}
		if _, err := entry.Write(files[name]); err != nil {
			out.Close()
			return err
		}
	}
	if err := w.Close(); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// ContainerFileName is the canonical export file name for a theme.
func ContainerFileName(name string) string {
	clean := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, r) {
			return '-'
		}
		return r
	}, strings.TrimSpace(name))
	if clean == "" {
		clean = "theme"
	}
	return clean + ".peltontheme"
}
