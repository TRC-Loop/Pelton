package themepack

import (
	"archive/zip"
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// buildContainer zips the given files into an in-memory .peltontheme.
func buildContainer(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		entry, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

const testManifest = `{
	"manifestVersion": 1,
	"name": "Test Theme",
	"author": "Tester",
	"version": "1.0.0",
	"base": "dark",
	"pelton": { "min": "1.0.8" },
	"tokens": ["tokens/colors.json"],
	"css": ["css/base.css", "css/extra.css"],
	"icons": { "pencil": "icons/pencil.svg" }
}`

func testFiles() map[string]string {
	return map[string]string{
		"manifest.json":      testManifest,
		"tokens/colors.json": `{"surface-base": "#2e3440", "accent": "#88c0d0"}`,
		"css/base.css":       `@font-face { font-family: X; src: url("assets/f.woff2"); }`,
		"css/extra.css":      `body { background: url("https://tracker.example/px.png"); }`,
		"icons/pencil.svg":   `<svg viewBox="0 0 24 24" stroke="currentColor"><path d="M4 20l4-1 11-11-3-3L5 16z"/></svg>`,
		"assets/f.woff2":     "fake-font-bytes",
		"README.md":          "ignored by the engine",
	}
}

func TestReadContainer(t *testing.T) {
	p, err := ReadContainer(buildContainer(t, testFiles()))
	if err != nil {
		t.Fatal(err)
	}
	if p.Manifest.ID != "test-theme" {
		t.Errorf("id = %q, want slug of name", p.Manifest.ID)
	}
	if p.Tokens["surface-base"] != "#2e3440" {
		t.Errorf("tokens not loaded: %v", p.Tokens)
	}
	if len(p.CSSFiles) != 2 {
		t.Fatalf("css files = %d, want 2", len(p.CSSFiles))
	}
	if refs := p.RemoteRefs(); len(refs) != 1 || !strings.Contains(refs[0], "tracker.example") {
		t.Errorf("remote refs = %v, want the tracker url", refs)
	}
	if _, ok := p.Icons["pencil"]; !ok {
		t.Errorf("icon override missing")
	}
}

func TestReadContainerRejects(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]string)
	}{
		{"no manifest", func(f map[string]string) { delete(f, "manifest.json") }},
		{"future format", func(f map[string]string) {
			f["manifest.json"] = strings.Replace(testManifest, `"manifestVersion": 1`, `"manifestVersion": 99`, 1)
		}},
		{"missing css file", func(f map[string]string) { delete(f, "css/base.css") }},
		{"bad token", func(f map[string]string) { f["tokens/colors.json"] = `{"space-1": "9px"}` }},
		{"scripted icon", func(f map[string]string) { f["icons/pencil.svg"] = `<svg><script>1</script></svg>` }},
		{"zip slip", func(f map[string]string) { f["../evil"] = "x" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := testFiles()
			tt.mutate(files)
			if _, err := ReadContainer(buildContainer(t, files)); err == nil {
				t.Fatal("ReadContainer accepted a bad container")
			}
		})
	}
}

func TestInstallLoadExportRoundTrip(t *testing.T) {
	root := t.TempDir()
	p, err := ReadContainer(buildContainer(t, testFiles()))
	if err != nil {
		t.Fatal(err)
	}
	dir, err := Install(p, root, true) // block remote refs
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadInstalled(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.RemoteRefs() != nil {
		t.Errorf("blocked install still has remote refs: %v", loaded.RemoteRefs())
	}
	css := loaded.AppliedCSS()
	if !strings.Contains(css, "data:font/woff2;base64,") {
		t.Errorf("bundled asset not inlined:\n%s", css)
	}
	if strings.Contains(css, "tracker.example") {
		t.Errorf("remote ref survived blocked install:\n%s", css)
	}

	dest := filepath.Join(root, ContainerFileName(loaded.Manifest.Name))
	if err := Export(dir, dest); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadInstalled(dir); err != nil {
		t.Fatal(err)
	}
}

func TestInstallKeepsRemoteWhenAllowed(t *testing.T) {
	p, err := ReadContainer(buildContainer(t, testFiles()))
	if err != nil {
		t.Fatal(err)
	}
	dir, err := Install(p, t.TempDir(), false) // user allowed remote refs
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadInstalled(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(loaded.AppliedCSS(), "tracker.example") {
		t.Error("allowed remote ref was stripped anyway")
	}
}
