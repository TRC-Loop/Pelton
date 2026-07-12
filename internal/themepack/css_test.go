package themepack

import (
	"strings"
	"testing"
)

func TestScanCSS(t *testing.T) {
	tests := []struct {
		name string
		css  string
		want int
	}{
		{"empty", "", 0},
		{"local asset", `@font-face { src: url("assets/font.woff2"); }`, 0},
		{"data uri", `body { background: url(data:image/png;base64,AAAA); }`, 0},
		{"http url", `body { background: url("https://evil.example/px.png"); }`, 1},
		{"protocol relative", `body { background: url(//evil.example/px.png); }`, 1},
		{"single quotes", `body { background: url('http://evil.example/a'); }`, 1},
		{"import", `@import "https://fonts.example/css";`, 1},
		{"import local also flagged", `@import "other.css";`, 1},
		{"mixed", `@import url(http://a/b); body { background: url("assets/x.png"), url("https://c/d"); }`, 3},
		{"case insensitive", `body { background: URL("HTTPS://E/F"); }`, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scanCSS(tt.css); len(got) != tt.want {
				t.Fatalf("scanCSS() found %d refs %v, want %d", len(got), got, tt.want)
			}
		})
	}
}

func TestStripRemote(t *testing.T) {
	css := `@import "https://fonts.example/css";
@font-face { src: url("assets/font.woff2"); }
body { background: url("https://evil.example/px.png"); color: red; }`
	got := StripRemote(css)
	if strings.Contains(got, "evil.example") || strings.Contains(got, "@import") {
		t.Fatalf("remote refs survived stripping: %s", got)
	}
	if !strings.Contains(got, `url("assets/font.woff2")`) {
		t.Fatalf("local asset ref was stripped: %s", got)
	}
	if scanCSS(got) != nil {
		t.Fatalf("stripped css still scans as remote: %v", scanCSS(got))
	}
}

func TestIsRemoteRef(t *testing.T) {
	tests := []struct {
		ref  string
		want bool
	}{
		{"assets/a.png", false},
		{"./assets/a.png", false},
		{"data:image/png;base64,AAAA", false},
		{"DATA:image/png;base64,AAAA", false},
		{"", false},
		{"https://a/b", true},
		{"http://a/b", true},
		{"//a/b", true},
		{"ftp://a/b", true},
		{"c:local-looking-but-scheme", true},
	}
	for _, tt := range tests {
		if got := isRemoteRef(tt.ref); got != tt.want {
			t.Errorf("isRemoteRef(%q) = %v, want %v", tt.ref, got, tt.want)
		}
	}
}
