package desktop

import "testing"

func TestParseListUnsubscribe(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		post       bool
		wantKind   string
		wantTarget string
	}{
		{
			name:       "one-click with https and post",
			header:     "<mailto:u@list.example>, <https://list.example/u?id=1>",
			post:       true,
			wantKind:   unsubKindOneClick,
			wantTarget: "https://list.example/u?id=1",
		},
		{
			name:       "post without https falls back to mailto",
			header:     "<mailto:u@list.example>",
			post:       true,
			wantKind:   unsubKindMailto,
			wantTarget: "mailto:u@list.example",
		},
		{
			name:       "mailto preferred over plain https without post",
			header:     "<https://list.example/u>, <mailto:u@list.example?subject=unsub>",
			wantKind:   unsubKindMailto,
			wantTarget: "mailto:u@list.example?subject=unsub",
		},
		{
			name:       "https only becomes a browser link",
			header:     "<https://list.example/u>",
			wantKind:   unsubKindLink,
			wantTarget: "https://list.example/u",
		},
		{
			name:       "http only becomes a browser link",
			header:     "<http://list.example/u>",
			wantKind:   unsubKindLink,
			wantTarget: "http://list.example/u",
		},
		{
			name:   "empty header yields nothing",
			header: "",
		},
		{
			name:   "garbage without angle brackets yields nothing",
			header: "https://list.example/u",
		},
		{
			name:       "whitespace inside entries is trimmed",
			header:     "< mailto:u@list.example >",
			wantKind:   unsubKindMailto,
			wantTarget: "mailto:u@list.example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, target := parseListUnsubscribe(tt.header, tt.post)
			if kind != tt.wantKind || target != tt.wantTarget {
				t.Fatalf("got (%q, %q), want (%q, %q)", kind, target, tt.wantKind, tt.wantTarget)
			}
		})
	}
}
