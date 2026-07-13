package desktop

import (
	"strings"
	"testing"
)

func TestParseUserLocale(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
		base    string
	}{
		{
			name: "full file",
			data: `{"name": "Test", "base": "de", "strings": {"a": "b"}}`,
			base: "de",
		},
		{
			name: "base defaults to english",
			data: `{"name": "Test", "strings": {"a": "b"}}`,
			base: "en",
		},
		{
			name:    "unknown base",
			data:    `{"name": "Test", "base": "xx", "strings": {"a": "b"}}`,
			wantErr: true,
		},
		{
			name:    "missing name",
			data:    `{"strings": {"a": "b"}}`,
			wantErr: true,
		},
		{
			name:    "no strings",
			data:    `{"name": "Test"}`,
			wantErr: true,
		},
		{
			name:    "not json",
			data:    `hello`,
			wantErr: true,
		},
		{
			name:    "strings not a map of strings",
			data:    `{"name": "Test", "strings": {"a": 1}}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := parseUserLocale([]byte(tt.data))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if f.Base != tt.base {
				t.Fatalf("base = %q, want %q", f.Base, tt.base)
			}
		})
	}
}

func TestLocaleIDPattern(t *testing.T) {
	valid := []string{"pt-br", "tr", "no.bokmaal", "x_1"}
	for _, id := range valid {
		if !localeIDPattern.MatchString(id) {
			t.Errorf("id %q should be valid", id)
		}
	}
	invalid := []string{"", "PT", "../evil", "a/b", strings.Repeat("a", 70), "-lead"}
	for _, id := range invalid {
		if localeIDPattern.MatchString(id) {
			t.Errorf("id %q should be invalid", id)
		}
	}
}
