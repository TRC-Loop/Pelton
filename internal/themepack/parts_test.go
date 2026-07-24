package themepack

import (
	"os"
	"path/filepath"
	"testing"
)

// roundTrip writes a package out and reads it back, the way an import lands
// in the themes folder.
func roundTrip(t *testing.T, p *Package) *Package {
	t.Helper()
	dest := filepath.Join(t.TempDir(), "theme.peltontheme")
	if err := WriteContainer(p, dest, true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	out, err := ReadContainer(data)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestSelectPartsDropsCSS(t *testing.T) {
	p, err := ReadContainer(buildContainer(t, testFiles()))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.SelectParts(true, false); err != nil {
		t.Fatal(err)
	}
	if _, ok := p.Files["css/base.css"]; ok {
		t.Error("css file still in container")
	}
	out := roundTrip(t, p)
	if len(out.CSSFiles) != 0 {
		t.Errorf("css survived the round trip: %d files", len(out.CSSFiles))
	}
	if out.Tokens["accent"] != "#88c0d0" {
		t.Errorf("tokens lost: %v", out.Tokens)
	}
	if len(out.Icons) != 1 {
		t.Errorf("icons lost: %v", out.Icons)
	}
}

func TestSelectPartsDropsTokens(t *testing.T) {
	p, err := ReadContainer(buildContainer(t, testFiles()))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.SelectParts(false, true); err != nil {
		t.Fatal(err)
	}
	if _, ok := p.Files["tokens/colors.json"]; ok {
		t.Error("token file still in container")
	}
	out := roundTrip(t, p)
	if len(out.Tokens) != 0 {
		t.Errorf("tokens survived the round trip: %v", out.Tokens)
	}
	if len(out.CSSFiles) != 2 {
		t.Errorf("css lost: %d files", len(out.CSSFiles))
	}
}

func TestSelectPartsKeepsEverything(t *testing.T) {
	p, err := ReadContainer(buildContainer(t, testFiles()))
	if err != nil {
		t.Fatal(err)
	}
	before := len(p.Files)
	if err := p.SelectParts(true, true); err != nil {
		t.Fatal(err)
	}
	if len(p.Files) != before {
		t.Errorf("files changed: %d -> %d", before, len(p.Files))
	}
	out := roundTrip(t, p)
	if len(out.Tokens) != 2 || len(out.CSSFiles) != 2 {
		t.Errorf("parts lost: %d tokens, %d css", len(out.Tokens), len(out.CSSFiles))
	}
}
