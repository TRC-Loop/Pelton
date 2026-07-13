package themepack

import (
	"strings"
	"testing"
)

func TestCheckSVG(t *testing.T) {
	tests := []struct {
		name    string
		svg     string
		wantErr bool
	}{
		{"plain icon", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M4 4h16"/></svg>`, false},
		{"xml decl and comment", `<?xml version="1.0"?><!-- ok --><svg viewBox="0 0 24 24"><circle r="4"/></svg>`, false},
		{"currentColor stroke", `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M1 1"/></svg>`, false},
		{"not svg", `<html><body>hi</body></html>`, true},
		{"script element", `<svg><script>alert(1)</script></svg>`, true},
		{"event handler", `<svg onload="alert(1)"><path d="M1 1"/></svg>`, true},
		{"foreignObject", `<svg><foreignObject><body/></foreignObject></svg>`, true},
		{"href", `<svg><use href="#x"/></svg>`, true},
		{"xlink href", `<svg><image xlink:href="https://evil/a.png"/></svg>`, true},
		{"css url", `<svg><style>path{fill:url(https://evil/a)}</style><path/></svg>`, true},
		{"doctype", `<!DOCTYPE svg><svg><path/></svg>`, true},
		{"oversized", `<svg>` + strings.Repeat("<path/>", maxSVGBytes/7+1) + `</svg>`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkSVG([]byte(tt.svg))
			if (err != nil) != tt.wantErr {
				t.Fatalf("checkSVG() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckIconName(t *testing.T) {
	for name, ok := range map[string]bool{
		"pencil": true, "arrow-left": true, "x2": true,
		"Pencil": false, "a b": false, "": false, "-lead": false,
	} {
		if err := checkIconName(name); (err == nil) != ok {
			t.Errorf("checkIconName(%q) error = %v, want ok=%v", name, err, ok)
		}
	}
}
