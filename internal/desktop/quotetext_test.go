package desktop

import "testing"

func TestQuoteTextPrefersTheTextPart(t *testing.T) {
	tests := []struct {
		name  string
		plain string
		html  string
		want  string
	}{
		{
			name:  "text part wins when present",
			plain: "the plain part",
			html:  "<p>the html part</p>",
			want:  "the plain part",
		},
		{
			name: "html only falls back to rendered text",
			html: "<p>Hi Arne,</p><p>Here is the thing.</p>",
			want: "Hi Arne,\nHere is the thing.",
		},
		{
			name:  "a whitespace-only text part is not a text part",
			plain: "\n  \n",
			html:  "<p>real content</p>",
			want:  "real content",
		},
		{
			name: "nothing at all",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := quoteText(tt.plain, tt.html); got != tt.want {
				t.Errorf("quoteText() = %q, want %q", got, tt.want)
			}
		})
	}
}
