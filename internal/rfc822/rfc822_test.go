package rfc822

import (
	"strings"
	"testing"
)

// multipart/related inside multipart/mixed is the shape a real message with an
// inline image and a real attachment arrives in, and it is what the imap path
// depends on this parser getting right.
const multipartMessage = "Message-ID: <root@example.com>\r\n" +
	"Subject: =?UTF-8?B?R3LDvMOfZQ==?=\r\n" +
	"From: \"Doe, Jane\" <jane@example.com>\r\n" +
	"To: a@example.com, Bob <b@example.com>\r\n" +
	"Cc: c@example.com\r\n" +
	"Date: Mon, 06 Jan 2020 10:00:00 +0000\r\n" +
	"List-Unsubscribe: <https://example.com/u>\r\n" +
	"List-Unsubscribe-Post: List-Unsubscribe=One-Click\r\n" +
	"MIME-Version: 1.0\r\n" +
	"Content-Type: multipart/mixed; boundary=outer\r\n" +
	"\r\n" +
	"--outer\r\n" +
	"Content-Type: multipart/alternative; boundary=inner\r\n" +
	"\r\n" +
	"--inner\r\n" +
	"Content-Type: text/plain; charset=utf-8\r\n" +
	"\r\n" +
	"plain body\r\n" +
	"--inner\r\n" +
	"Content-Type: text/html; charset=utf-8\r\n" +
	"\r\n" +
	"<p>html body</p>\r\n" +
	"--inner--\r\n" +
	"--outer\r\n" +
	"Content-Type: image/png\r\n" +
	"Content-Disposition: attachment; filename=\"logo.png\"\r\n" +
	"Content-Id: <logo@example.com>\r\n" +
	"Content-Transfer-Encoding: base64\r\n" +
	"\r\n" +
	"aGVsbG8=\r\n" +
	"--outer\r\n" +
	"Content-Type: text/plain\r\n" +
	"Content-Disposition: attachment; filename=\"notes.txt\"\r\n" +
	"\r\n" +
	"attached text\r\n" +
	"--outer--\r\n"

func TestParseMultipart(t *testing.T) {
	msg, err := Parse([]byte(multipartMessage))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if msg.Subject != "Grüße" {
		t.Errorf("subject = %q, want Grüße", msg.Subject)
	}
	if msg.MessageID != "root@example.com" {
		t.Errorf("message id = %q", msg.MessageID)
	}
	// a display name containing a comma must stay one address, not split into
	// two, which is exactly what naive comma-splitting gets wrong.
	if msg.From != "Doe, Jane <jane@example.com>" {
		t.Errorf("from = %q", msg.From)
	}
	if msg.To != "a@example.com, Bob <b@example.com>" {
		t.Errorf("to = %q", msg.To)
	}
	if msg.Cc != "c@example.com" {
		t.Errorf("cc = %q", msg.Cc)
	}
	if msg.Date.IsZero() {
		t.Error("date was not parsed")
	}
	if strings.TrimSpace(msg.Text) != "plain body" {
		t.Errorf("text = %q", msg.Text)
	}
	if !strings.Contains(msg.HTML, "<p>html body</p>") {
		t.Errorf("html = %q", msg.HTML)
	}
	if !msg.ListUnsubscribePost {
		t.Error("one-click unsubscribe was not detected")
	}
	if msg.Size != int64(len(multipartMessage)) {
		t.Errorf("size = %d, want %d", msg.Size, len(multipartMessage))
	}

	if len(msg.Attachments) != 2 {
		t.Fatalf("got %d attachments, want 2", len(msg.Attachments))
	}
	inline := msg.Attachments[0]
	if inline.ContentID != "logo@example.com" {
		t.Errorf("content id = %q, want the angle brackets stripped", inline.ContentID)
	}
	if string(inline.Content) != "hello" {
		t.Errorf("inline content = %q, want the base64 decoded", inline.Content)
	}
	if strings.TrimSpace(string(msg.Attachments[1].Content)) != "attached text" {
		t.Errorf("attachment content = %q", msg.Attachments[1].Content)
	}
}

// ParseBody is what the imap path uses: the server supplies the envelope, so
// only the bodies and attachments are read here.
func TestParseBodySkipsEnvelope(t *testing.T) {
	msg, err := ParseBody([]byte(multipartMessage))
	if err != nil {
		t.Fatalf("parse body: %v", err)
	}
	if msg.Subject != "" || msg.From != "" {
		t.Errorf("ParseBody filled envelope fields: subject=%q from=%q", msg.Subject, msg.From)
	}
	if strings.TrimSpace(msg.Text) != "plain body" || len(msg.Attachments) != 2 {
		t.Errorf("ParseBody did not read the body: %+v", msg)
	}
}

// a message with a charset Go has no decoder for still has readable headers and
// structure, so it must not fail the whole parse.
func TestParseUnknownCharsetIsNotFatal(t *testing.T) {
	const raw = "Subject: hi\r\n" +
		"From: a@example.com\r\n" +
		"Content-Type: text/plain; charset=x-made-up\r\n" +
		"\r\n" +
		"body\r\n"
	msg, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if msg.Subject != "hi" {
		t.Errorf("subject = %q", msg.Subject)
	}
}

// a broken address header must not cost the whole message: the raw value is
// better than nothing in the sender column.
func TestParseFallsBackOnUnparseableAddress(t *testing.T) {
	const raw = "Subject: hi\r\nFrom: not an address at all\r\n\r\nbody\r\n"
	msg, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if msg.From != "not an address at all" {
		t.Errorf("from = %q, want the raw header value", msg.From)
	}
}

// TestParseBodyReadsAuthenticationHeaders: the headers are dropped once a
// message is stored, so anything not read here is gone for good.
func TestParseBodyReadsAuthenticationHeaders(t *testing.T) {
	raw := []byte("From: Someone <someone@example.com>\r\n" +
		"Reply-To: Other <other@elsewhere.test>\r\n" +
		"Authentication-Results: mine.example.org; spf=fail smtp.mailfrom=evil.test\r\n" +
		"Authentication-Results: relay.example.net; dkim=pass header.d=example.com\r\n" +
		"Subject: Hi\r\n" +
		"\r\n" +
		"body\r\n")

	msg, err := ParseBody(raw)
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if msg.ReplyTo != "Other <other@elsewhere.test>" {
		t.Errorf("ReplyTo = %q", msg.ReplyTo)
	}
	if len(msg.AuthResults) != 2 {
		t.Fatalf("AuthResults = %v, want both headers in order", msg.AuthResults)
	}
	if !strings.Contains(msg.AuthResults[0], "spf=fail") {
		t.Errorf("first header = %q, want the receiving server's one first", msg.AuthResults[0])
	}
}

// TestParseBodyWithoutAuthenticationHeaders: most mail has none, and that must
// come out empty rather than as anything a check could read as a failure.
func TestParseBodyWithoutAuthenticationHeaders(t *testing.T) {
	msg, err := ParseBody([]byte("From: a@example.com\r\nSubject: Hi\r\n\r\nbody\r\n"))
	if err != nil {
		t.Fatalf("ParseBody: %v", err)
	}
	if len(msg.AuthResults) != 0 || msg.ReplyTo != "" {
		t.Errorf("AuthResults = %v, ReplyTo = %q, want both empty", msg.AuthResults, msg.ReplyTo)
	}
}
