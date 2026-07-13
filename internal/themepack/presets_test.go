package themepack

import (
	"strings"
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
		if !strings.HasPrefix(m.ID, "preset-") {
			t.Errorf("preset id %q must carry the preset- prefix so it cannot collide with installed themes", m.ID)
		}
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
		if by, ok := Preset(m.ID); !ok || by != p {
			t.Errorf("Preset(%q) does not return the listed package", m.ID)
		}
	}
}
