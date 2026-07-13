package themepack

import (
	"os"
	"testing"
)

func TestPresetsLoadAndValidate(t *testing.T) {
	presets := Presets()
	if len(presets) == 0 {
		t.Fatal("no embedded presets loaded")
	}
	seen := make(map[string]bool)
	for _, p := range presets {
		m := p.Manifest
		if seen[m.ID] {
			t.Errorf("duplicate preset id %q", m.ID)
		}
		seen[m.ID] = true
		if m.Base != "light" && m.Base != "dark" {
			t.Errorf("preset %s: base %q", m.ID, m.Base)
		}
		if len(p.Tokens) == 0 {
			t.Errorf("preset %s: no tokens", m.ID)
		}
		if len(p.CSSFiles) != 0 || len(p.Icons) != 0 {
			t.Errorf("preset %s: presets must be token-only", m.ID)
		}
	}
}

func TestWriteContainerRoundTrip(t *testing.T) {
	presets := Presets()
	if len(presets) == 0 {
		t.Fatal("no presets")
	}
	dest := t.TempDir() + "/roundtrip.peltontheme"
	if err := WriteContainer(presets[0], dest, false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	back, err := ReadContainer(data)
	if err != nil {
		t.Fatal(err)
	}
	if back.Manifest.ID != presets[0].Manifest.ID {
		t.Fatalf("round trip changed id: %q -> %q", presets[0].Manifest.ID, back.Manifest.ID)
	}
	if len(back.Tokens) != len(presets[0].Tokens) {
		t.Fatalf("round trip changed token count: %d -> %d", len(presets[0].Tokens), len(back.Tokens))
	}
}
