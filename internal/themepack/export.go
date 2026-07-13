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
	return writeZip(files, dest)
}

// WriteContainer writes a parsed package back out as a .peltontheme file.
// When blockRemote is set - the user's choice at the import warning - every
// css file is written with its remote references stripped, so the container
// on disk is exactly what will run. This is how themes land in the themes
// folder: imports copy through here, the palette editor saves through here,
// and the default themes are seeded through here.
func WriteContainer(p *Package, dest string, blockRemote bool) error {
	files := make(map[string][]byte, len(p.Files))
	cssPaths := make(map[string]bool, len(p.CSSFiles))
	for _, f := range p.CSSFiles {
		cssPaths[f.Path] = true
	}
	for name, content := range p.Files {
		if blockRemote && cssPaths[name] {
			content = []byte(StripRemote(string(content)))
		}
		files[name] = content
	}
	return writeZip(files, dest)
}

// writeZip writes a path-keyed file map as a zip at dest, entries in sorted
// order for reproducibility.
func writeZip(files map[string][]byte, dest string) error {
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
