package themepack

import "testing"

func TestCompatWarning(t *testing.T) {
	tests := []struct {
		name     string
		r        *VersionRange
		app      string
		wantWarn bool
	}{
		{"no range", nil, "1.0.8", false},
		{"empty range", &VersionRange{}, "1.0.8", false},
		{"inside", &VersionRange{Min: "1.0.8", Max: "1.1"}, "1.0.9", false},
		{"at min", &VersionRange{Min: "1.0.8"}, "1.0.8", false},
		{"below min", &VersionRange{Min: "1.0.8"}, "1.0.7", true},
		{"max prefix allows patch", &VersionRange{Max: "1.1"}, "1.1.5", false},
		{"above max", &VersionRange{Max: "1.1"}, "1.2.0", true},
		{"v prefix", &VersionRange{Min: "1.0.8"}, "v1.0.8", false},
		{"dev build always passes", &VersionRange{Min: "9.9"}, "test-5d49e09", false},
		{"unparsable range ignored", &VersionRange{Min: "not-a-version"}, "1.0.8", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompatWarning(tt.r, tt.app)
			if (got != "") != tt.wantWarn {
				t.Fatalf("CompatWarning() = %q, wantWarn %v", got, tt.wantWarn)
			}
		})
	}
}
