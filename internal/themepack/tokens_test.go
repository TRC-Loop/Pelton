package themepack

import "testing"

func TestValidateTokens(t *testing.T) {
	tests := []struct {
		name    string
		tokens  map[string]string
		wantErr bool
	}{
		{"colors", map[string]string{"surface-base": "#2e3440", "accent": "#88c0d0"}, false},
		{"double dash prefix accepted", map[string]string{"--text-primary": "#eceff4"}, false},
		{"font stack", map[string]string{"font-ui": `"Inter", system-ui, sans-serif`}, false},
		{"color-mix", map[string]string{"selection-bg": "color-mix(in srgb, #88c0d0 20%, transparent)"}, false},
		{"shadow", map[string]string{"shadow-overlay": "0 8px 24px rgba(0,0,0,0.5)"}, false},
		{"unknown token", map[string]string{"row-pad-y": "40px"}, true},
		{"density token blocked", map[string]string{"space-1": "10px"}, true},
		{"semicolon injection", map[string]string{"accent": "#fff; background: red"}, true},
		{"brace escape", map[string]string{"accent": "#fff } body { display: none"}, true},
		{"url in value", map[string]string{"accent": "url(https://evil/a)"}, true},
		{"at rule", map[string]string{"accent": "@import 'x'"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateTokens(tt.tokens)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateTokens() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSlug(t *testing.T) {
	for in, want := range map[string]string{
		"Nord":          "nord",
		"My Cool Theme": "my-cool-theme",
		"  Ümlaut!! ":   "mlaut",
		"---":           "theme",
	} {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}
