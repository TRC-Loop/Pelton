package themepack

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"
)

// presetsFS embeds the default themes that are seeded into the user's themes
// folder as regular .peltontheme files on first use. Each preset is a normal
// theme (manifest.json with inline tokens) and goes through the same parsing
// and validation as an imported container. Once seeded, the files in the
// themes folder are the themes; the embedded copies are only the seed source.
//
//go:embed presets
var presetsFS embed.FS

var (
	presetsOnce sync.Once
	presetsByID map[string]*Package
	presetsList []*Package
)

// loadPresets parses every embedded preset once. A preset failing to parse is
// a programming error caught by the package tests, so it panics rather than
// being skipped silently.
func loadPresets() {
	presetsByID = make(map[string]*Package)
	entries, err := presetsFS.ReadDir("presets")
	if err != nil {
		panic(fmt.Sprintf("themepack: read embedded presets: %v", err))
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := loadPresetDir("presets/" + e.Name())
		if err != nil {
			panic(fmt.Sprintf("themepack: preset %s: %v", e.Name(), err))
		}
		presetsByID[p.Manifest.ID] = p
		presetsList = append(presetsList, p)
	}
	sort.Slice(presetsList, func(i, j int) bool {
		return presetsList[i].Manifest.Name < presetsList[j].Manifest.Name
	})
}

// loadPresetDir reads one embedded preset folder into the same path-keyed map
// the zip and folder readers produce, then validates it like any theme.
func loadPresetDir(dir string) (*Package, error) {
	files := make(map[string][]byte)
	err := fs.WalkDir(presetsFS, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, err := presetsFS.ReadFile(p)
		if err != nil {
			return err
		}
		files[strings.TrimPrefix(p, dir+"/")] = content
		return nil
	})
	if err != nil {
		return nil, err
	}
	manifestData, ok := files["manifest.json"]
	if !ok {
		return nil, fmt.Errorf("no manifest.json")
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

// Presets returns the default themes to seed the themes folder with, sorted
// by name.
func Presets() []*Package {
	presetsOnce.Do(loadPresets)
	return presetsList
}
