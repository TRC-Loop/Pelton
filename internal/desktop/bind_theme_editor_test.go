package desktop

import (
	"slices"
	"strings"
	"testing"

	"github.com/TRC-Loop/Pelton/internal/themepack"
)

func TestSetEditorCSSAddsOnce(t *testing.T) {
	manifest := themepack.Manifest{CSS: []string{"css/theme.css"}}
	files := map[string][]byte{}
	for range 2 {
		if err := setEditorCSS(&manifest, files, "body { color: red; }"); err != nil {
			t.Fatal(err)
		}
	}
	if got := slices.Index(manifest.CSS, editorCSSPath); got != 1 {
		t.Errorf("editor css at %d, want appended after the theme's own", got)
	}
	if len(manifest.CSS) != 2 {
		t.Errorf("css list %v, want no duplicates", manifest.CSS)
	}
	if _, ok := files[editorCSSPath]; !ok {
		t.Error("editor css not written")
	}
}

func TestSetEditorCSSStripsRemote(t *testing.T) {
	manifest := themepack.Manifest{}
	files := map[string][]byte{}
	css := `@import url("https://evil.example/x.css");
body { background: url("https://tracker.example/px.png"); }`
	if err := setEditorCSS(&manifest, files, css); err != nil {
		t.Fatal(err)
	}
	got := string(files[editorCSSPath])
	if strings.Contains(got, "tracker.example") || strings.Contains(got, "@import") {
		t.Errorf("remote references survived: %q", got)
	}
}

func TestSetEditorCSSClears(t *testing.T) {
	manifest := themepack.Manifest{CSS: []string{"css/theme.css", editorCSSPath}}
	files := map[string][]byte{editorCSSPath: []byte("body{}")}
	if err := setEditorCSS(&manifest, files, "   "); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(manifest.CSS, editorCSSPath) {
		t.Errorf("css list still references the editor sheet: %v", manifest.CSS)
	}
	if _, ok := files[editorCSSPath]; ok {
		t.Error("editor css file not removed")
	}
	if !slices.Contains(manifest.CSS, "css/theme.css") {
		t.Error("the theme's own stylesheet was dropped")
	}
}

func TestSetEditorCSSTooLarge(t *testing.T) {
	manifest := themepack.Manifest{}
	if err := setEditorCSS(&manifest, map[string][]byte{}, strings.Repeat("a", maxEditorCSS+1)); err == nil {
		t.Fatal("oversized css accepted")
	}
}
