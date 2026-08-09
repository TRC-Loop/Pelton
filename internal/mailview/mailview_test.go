package mailview

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestLinks(t *testing.T) {
	tests := []struct {
		name  string
		html  string
		plain string
		want  []string
	}{
		{
			name: "anchor hrefs, quoted and bare",
			html: `<a href="https://one.example/">a</a> <a href='https://two.example/'>b</a> <a href=https://three.example/>c</a>`,
			want: []string{"https://one.example/", "https://two.example/", "https://three.example/"},
		},
		{
			name: "entities in the href are decoded",
			html: `<a href="https://example.com/p?a=1&amp;b=2">go</a>`,
			want: []string{"https://example.com/p?a=1&b=2"},
		},
		{
			// an image is remote content, not something the reader can click,
			// and scanning every tracking pixel would burn the request quota.
			name: "image sources are not links",
			html: `<img src="https://tracker.example/pixel.gif"><a href="https://real.example/">x</a>`,
			want: []string{"https://real.example/"},
		},
		{
			name: "other schemes are skipped",
			html: `<a href="mailto:a@b.example">mail</a><a href="cid:abc">inline</a><a href="ftp://x.example/">ftp</a>`,
			want: []string{},
		},
		{
			name:  "bare urls in plain text",
			plain: "see https://plain.example/path for details",
			want:  []string{"https://plain.example/path"},
		},
		{
			name:  "sentence punctuation is not part of a bare url",
			plain: "go to https://example.com/page, or https://example.com/other.",
			want:  []string{"https://example.com/page", "https://example.com/other"},
		},
		{
			// an href states the target exactly. trimming it would scan (and
			// badge) a different address than the one the link opens, which is
			// worse than scanning one character too many.
			name: "an href keeps punctuation the reader would click",
			html: `<a href="https://en.example.org/wiki/Foo_(bar)">x</a><a href="https://example.com/end.">y</a>`,
			want: []string{"https://en.example.org/wiki/Foo_(bar)", "https://example.com/end."},
		},
		{
			// the ui's linkifier stops a bare url at these characters, so the
			// backend has to as well or the badge lands on nothing.
			name:  "a bare url stops where the ui stops it",
			plain: "see (https://example.com/a) and [https://example.com/b]",
			want:  []string{"https://example.com/a", "https://example.com/b"},
		},
		{
			name:  "the same target is only returned once",
			html:  `<a href="https://dup.example/">one</a><a href="https://dup.example/">two</a>`,
			plain: "https://dup.example/",
			want:  []string{"https://dup.example/"},
		},
		{
			name: "attributes before href do not swallow it",
			html: `<a class="btn" target="_blank" href="https://late.example/">x</a>`,
			want: []string{"https://late.example/"},
		},
		{
			name: "nothing to find",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Links(tt.html, tt.plain)
			if !slices.Equal(got, tt.want) {
				t.Errorf("Links() = %q, want %q", got, tt.want)
			}
		})
	}
}

// A message can carry more anchors than anyone would ever scan; the cap is what
// keeps one open message from turning into thousands of api lookups.
func TestLinksIsCapped(t *testing.T) {
	var b strings.Builder
	for i := range maxLinks * 2 {
		fmt.Fprintf(&b, `<a href="https://example.com/%d">x</a>`, i)
	}
	if got := len(Links(b.String(), "")); got != maxLinks {
		t.Errorf("len(Links()) = %d, want %d", got, maxLinks)
	}
}
