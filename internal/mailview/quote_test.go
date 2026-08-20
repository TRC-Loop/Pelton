package mailview

import "testing"

func TestTextForQuoteKeepsLineStructure(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "paragraphs become separate lines",
			html: "<p>First line.</p><p>Second line.</p>",
			want: "First line.\nSecond line.",
		},
		{
			name: "br breaks a line",
			html: "<p>One<br>Two</p>",
			want: "One\nTwo",
		},
		{
			name: "source wrapping is not a line break",
			html: "<p>a sentence that the sender's\n   html happened to wrap</p>",
			want: "a sentence that the sender's html happened to wrap",
		},
		{
			name: "inline markup stays on its line",
			html: "<p>hello <b>bold</b> and <i>italic</i> world</p>",
			want: "hello bold and italic world",
		},
		{
			name: "entities are decoded",
			html: "<p>Tom &amp; Jerry &lt;3</p>",
			want: "Tom & Jerry <3",
		},
		{
			name: "list items get their own lines",
			html: "<ul><li>one</li><li>two</li></ul>",
			want: "one\ntwo",
		},
		{
			name: "table cells get their own lines",
			html: "<table><tr><td>left</td><td>right</td></tr></table>",
			want: "left\nright",
		},
		{
			name: "script and style content is dropped",
			html: "<p>visible</p><script>var x = 1</script><style>p{color:red}</style>",
			want: "visible",
		},
		{
			name: "nested layout tables do not produce blank screenfuls",
			html: "<table><tr><td><table><tr><td><div><p>hi</p></div></td></tr></table></td></tr></table>",
			want: "hi",
		},
		{
			name: "empty input",
			html: "",
			want: "",
		},
		{
			name: "markup with no text",
			html: "<div></div><p></p>",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TextForQuote(tt.html); got != tt.want {
				t.Errorf("TextForQuote() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestTextForQuoteOnARealisticNewsletter is the shape of the message from the
// report: html only, wrapped in layout tables, which used to quote as nothing
// at all.
func TestTextForQuoteOnARealisticNewsletter(t *testing.T) {
	const html = `<html><body style="margin:0">
<table width="100%"><tr><td align="center">
  <table width="600"><tr><td>
    <h1>Weekly update</h1>
    <p>Hi Arne,</p>
    <p>Here is what happened this week. Nothing much,
       but we wrote it down anyway.</p>
    <ul><li>One thing</li><li>Another thing</li></ul>
    <p>Cheers,<br>The team</p>
  </td></tr></table>
</td></tr></table>
</body></html>`

	want := "Weekly update\nHi Arne,\nHere is what happened this week. Nothing much, but we wrote it down anyway.\nOne thing\nAnother thing\nCheers,\nThe team"
	if got := TextForQuote(html); got != want {
		t.Errorf("TextForQuote() =\n%q\nwant\n%q", got, want)
	}
}

// TestPlainTextStillFlattens guards the split: the snippet and search path
// deliberately collapses to one line, and must not pick up line breaks from
// this change.
func TestPlainTextStillFlattens(t *testing.T) {
	const html = "<p>First line.</p><p>Second line.</p>"
	if got := PlainText(html); got != "First line. Second line." {
		t.Errorf("PlainText() = %q, want it still flattened to one line", got)
	}
}
