package desktop

import "testing"

func TestParseMailto(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want MailtoDraft
	}{
		{
			name: "empty scheme only",
			in:   "mailto:",
			want: MailtoDraft{},
		},
		{
			name: "single recipient",
			in:   "mailto:spud@pelton.email",
			want: MailtoDraft{To: "spud@pelton.email"},
		},
		{
			name: "multiple recipients in path",
			in:   "mailto:a@x.test,b@y.test",
			want: MailtoDraft{To: "a@x.test, b@y.test"},
		},
		{
			name: "subject and body decoded",
			in:   "mailto:a@x.test?subject=Hello%20there&body=Line%20one",
			want: MailtoDraft{To: "a@x.test", Subject: "Hello there", Body: "Line one"},
		},
		{
			name: "cc and bcc",
			in:   "mailto:a@x.test?cc=c@x.test&bcc=d@x.test",
			want: MailtoDraft{To: "a@x.test", Cc: "c@x.test", Bcc: "d@x.test"},
		},
		{
			name: "extra to in query merges with path",
			in:   "mailto:a@x.test?to=b@y.test",
			want: MailtoDraft{To: "a@x.test, b@y.test"},
		},
		{
			name: "plus is a literal plus, not a space (RFC 6068)",
			in:   "mailto:a+tag@x.test?subject=one+two",
			want: MailtoDraft{To: "a+tag@x.test", Subject: "one+two"},
		},
		{
			name: "crlf in body preserved",
			in:   "mailto:a@x.test?body=first%0D%0Asecond",
			want: MailtoDraft{To: "a@x.test", Body: "first\r\nsecond"},
		},
		{
			name: "unknown headers ignored",
			in:   "mailto:a@x.test?subject=hi&x-custom=nope&priority=high",
			want: MailtoDraft{To: "a@x.test", Subject: "hi"},
		},
		{
			name: "case-insensitive scheme and field names",
			in:   "MAILTO:a@x.test?SUBJECT=Hi",
			want: MailtoDraft{To: "a@x.test", Subject: "Hi"},
		},
		{
			name: "no recipient, subject only",
			in:   "mailto:?subject=hi",
			want: MailtoDraft{Subject: "hi"},
		},
		{
			name: "encoded characters in recipient decoded",
			in:   "mailto:john.doe%40x.test",
			want: MailtoDraft{To: "john.doe@x.test"},
		},
		{
			name: "not a mailto url",
			in:   "https://example.test",
			want: MailtoDraft{},
		},
		{
			name: "double-slash form tolerated",
			in:   "mailto://a@x.test?subject=hi",
			want: MailtoDraft{To: "a@x.test", Subject: "hi"},
		},
		{
			name: "malformed percent escape passed through",
			in:   "mailto:a@x.test?subject=100%",
			want: MailtoDraft{To: "a@x.test", Subject: "100%"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMailto(tt.in)
			if got != tt.want {
				t.Errorf("parseMailto(%q)\n got  %+v\n want %+v", tt.in, got, tt.want)
			}
		})
	}
}

func TestFirstMailtoArg(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"none", []string{"pelton", "--flag"}, ""},
		{"present", []string{"pelton", "mailto:a@x.test?subject=hi"}, "mailto:a@x.test?subject=hi"},
		{"first wins", []string{"mailto:a@x.test", "mailto:b@y.test"}, "mailto:a@x.test"},
		{"case insensitive", []string{"MailTo:a@x.test"}, "MailTo:a@x.test"},
		{"empty", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstMailtoArg(tt.args); got != tt.want {
				t.Errorf("firstMailtoArg(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestConsumePendingMailto(t *testing.T) {
	a := &App{}
	if got := a.ConsumePendingMailto(); got.Present {
		t.Fatalf("expected no pending mailto on a fresh app, got %+v", got)
	}
	a.setPendingMailto(MailtoDraft{To: "a@x.test", Subject: "hi"})
	got := a.ConsumePendingMailto()
	if !got.Present || got.Draft.To != "a@x.test" || got.Draft.Subject != "hi" {
		t.Fatalf("unexpected consumed draft: %+v", got)
	}
	// a second consume must be empty: the draft is one-shot.
	if again := a.ConsumePendingMailto(); again.Present {
		t.Fatalf("expected pending mailto to be cleared after consume, got %+v", again)
	}
}
