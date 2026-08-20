package mailimport

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// readAll drains an MboxReader into the message bodies it produced.
func readAll(t *testing.T, content string) []string {
	t.Helper()
	r := NewMboxReader(strings.NewReader(content))
	var out []string
	for {
		msg, err := r.Next()
		if errors.Is(err, io.EOF) {
			return out
		}
		if err != nil {
			t.Fatalf("next: %v", err)
		}
		out = append(out, string(msg))
	}
}

func TestMboxSplitsOnSeparatorLines(t *testing.T) {
	const content = "From alice@example.com Mon Jan  1 00:00:00 2020\r\n" +
		"Subject: one\r\n\r\nbody one\r\n" +
		"\r\n" +
		"From bob@example.com Tue Jan  2 00:00:00 2020\r\n" +
		"Subject: two\r\n\r\nbody two\r\n"

	msgs := readAll(t, content)
	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2", len(msgs))
	}
	if !strings.Contains(msgs[0], "Subject: one") || strings.Contains(msgs[0], "Subject: two") {
		t.Fatalf("first message spilled into the second: %q", msgs[0])
	}
	if !strings.Contains(msgs[1], "body two") {
		t.Fatalf("second message = %q", msgs[1])
	}
}

// a "From " line inside a body is only a separator when a blank line precedes
// it. Without this check a message quoting one would be cut in half.
func TestMboxKeepsFromLineInsideBody(t *testing.T) {
	const content = "From alice@example.com Mon Jan  1 00:00:00 2020\r\n" +
		"Subject: quoting\r\n\r\n" +
		"here is what they sent:\r\n" +
		"From nobody@example.com Wed Jan  3 00:00:00 2020\r\n" +
		"and that was it\r\n"

	msgs := readAll(t, content)
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if !strings.Contains(msgs[0], "and that was it") {
		t.Fatalf("body was truncated at the inner From line: %q", msgs[0])
	}
}

// the write side escapes body lines that would read as separators, so the read
// side has to put the original text back.
func TestMboxUnescapesQuotedFromLines(t *testing.T) {
	const content = "From alice@example.com Mon Jan  1 00:00:00 2020\r\n" +
		"Subject: escaped\r\n\r\n" +
		">From here on\r\n" +
		">>From there on\r\n" +
		">not a from line\r\n"

	msgs := readAll(t, content)
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	body := msgs[0]
	if !strings.Contains(body, "\r\nFrom here on\r\n") {
		t.Fatalf("single-quoted From was not unescaped: %q", body)
	}
	if !strings.Contains(body, "\r\n>From there on\r\n") {
		t.Fatalf("double-quoted From lost the wrong number of markers: %q", body)
	}
	if !strings.Contains(body, "\r\n>not a from line\r\n") {
		t.Fatalf("an unrelated quoted line was rewritten: %q", body)
	}
}

func TestMboxRejectsNonMbox(t *testing.T) {
	r := NewMboxReader(strings.NewReader("Subject: plain eml\r\n\r\nhello\r\n"))
	if _, err := r.Next(); !errors.Is(err, ErrNotMbox) {
		t.Fatalf("next on a plain message = %v, want ErrNotMbox", err)
	}
}

func TestMboxEmptyFile(t *testing.T) {
	r := NewMboxReader(strings.NewReader(""))
	if _, err := r.Next(); !errors.Is(err, ErrNotMbox) {
		t.Fatalf("next on an empty file = %v, want ErrNotMbox", err)
	}
}

// a file that ends without a trailing newline still holds a whole message.
func TestMboxLastMessageWithoutTrailingNewline(t *testing.T) {
	const content = "From alice@example.com Mon Jan  1 00:00:00 2020\r\n" +
		"Subject: last\r\n\r\nno newline at eof"

	msgs := readAll(t, content)
	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if !strings.HasSuffix(msgs[0], "no newline at eof") {
		t.Fatalf("last message = %q", msgs[0])
	}
}
