// Package rfc822 parses a raw internet message into the pieces Pelton stores:
// the envelope headers, a plain and an html body, and the decoded attachments.
//
// It exists because messages reach Pelton two ways. Over imap the server hands
// back a parsed ENVELOPE alongside the raw source, so only the body needs
// walking; a message read out of an .eml or .mbox file has no server to ask,
// so its headers have to be parsed too. Both paths share the body walk here
// rather than keeping two copies of it.
package rfc822

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"

	// registers legacy charset decoders (ISO-8859-*, Windows-125x, ...)
	_ "github.com/emersion/go-message/charset"
)

// Message is a parsed message. Text and HTML hold the first body part of each
// type; a message that has only one of them leaves the other empty.
type Message struct {
	MessageID string
	Subject   string
	From      string
	To        string
	Cc        string
	Date      time.Time
	Text      string
	HTML      string
	// Size is the raw source length in bytes.
	Size        int64
	Attachments []Attachment
	// ListUnsubscribe is the raw List-Unsubscribe header value, and
	// ListUnsubscribePost whether the message declares RFC 8058 one-click
	// support via List-Unsubscribe-Post.
	ListUnsubscribe     string
	ListUnsubscribePost bool
	// Header is the message's top-level header, for the fields Pelton reads
	// only in one place and so does not promote above (a client's own
	// X- headers, for instance).
	Header mail.Header
}

// Attachment holds attachment metadata and its decoded content. ContentID is
// set for inline cid-referenced parts.
type Attachment struct {
	Filename    string
	ContentType string
	ContentID   string
	Content     []byte
}

// Parse reads a complete message, headers and all. A malformed date or address
// header is not fatal: the corresponding field is left empty, since a message
// that is readable apart from one bad header is still worth keeping.
func Parse(raw []byte) (*Message, error) {
	msg := &Message{Size: int64(len(raw))}
	mr, err := reader(raw)
	if err != nil {
		return nil, err
	}

	msg.Subject, _ = mr.Header.Subject()
	msg.MessageID, _ = mr.Header.MessageID()
	msg.From = addressList(&mr.Header, "From")
	msg.To = addressList(&mr.Header, "To")
	msg.Cc = addressList(&mr.Header, "Cc")
	if date, err := mr.Header.Date(); err == nil {
		msg.Date = date
	}

	if err := parts(mr, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// ParseBody reads only the bodies and attachments, for callers that already
// have the envelope from elsewhere (the imap server's ENVELOPE response).
func ParseBody(raw []byte) (*Message, error) {
	msg := &Message{Size: int64(len(raw))}
	mr, err := reader(raw)
	if err != nil {
		return nil, err
	}
	if err := parts(mr, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// reader opens a mail reader over raw. An unknown charset is reported by
// CreateReader but leaves the reader usable, so it is not treated as failure.
func reader(raw []byte) (*mail.Reader, error) {
	mr, err := mail.CreateReader(bytes.NewReader(raw))
	if err != nil && !message.IsUnknownCharset(err) {
		return nil, fmt.Errorf("rfc822: create mail reader: %w", err)
	}
	return mr, nil
}

// parts walks the message body, filling in the bodies, the attachments and the
// unsubscribe headers.
func parts(mr *mail.Reader, msg *Message) error {
	msg.Header = mr.Header
	msg.ListUnsubscribe = mr.Header.Get("List-Unsubscribe")
	msg.ListUnsubscribePost = strings.Contains(strings.ToLower(mr.Header.Get("List-Unsubscribe-Post")), "one-click")

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil && !message.IsUnknownCharset(err) {
			return fmt.Errorf("rfc822: read part: %w", err)
		}

		switch header := part.Header.(type) {
		case *mail.InlineHeader:
			body, err := io.ReadAll(part.Body)
			if err != nil {
				return fmt.Errorf("rfc822: read inline part: %w", err)
			}
			contentType, _, _ := header.ContentType()
			if strings.EqualFold(contentType, "text/html") {
				if msg.HTML == "" {
					msg.HTML = string(body)
				}
			} else if msg.Text == "" {
				msg.Text = string(body)
			}
		case *mail.AttachmentHeader:
			filename, _ := header.Filename()
			contentType, _, _ := header.ContentType()
			content, err := io.ReadAll(part.Body)
			if err != nil {
				return fmt.Errorf("rfc822: read attachment part: %w", err)
			}
			// content-id arrives wrapped in angle brackets, strip them
			contentID := strings.Trim(header.Get("Content-Id"), "<>")
			msg.Attachments = append(msg.Attachments, Attachment{
				Filename:    filename,
				ContentType: contentType,
				ContentID:   contentID,
				Content:     content,
			})
		}
	}
	return nil
}

// addressList renders one address header as `Name <user@host>, ...`, matching
// how the imap path formats the server's ENVELOPE so both look the same in the
// message list. An unparseable header falls back to its raw value rather than
// dropping the sender entirely.
func addressList(h *mail.Header, key string) string {
	addrs, err := h.AddressList(key)
	if err != nil {
		return strings.TrimSpace(h.Get(key))
	}
	parts := make([]string, 0, len(addrs))
	for _, a := range addrs {
		switch {
		case a.Name != "" && a.Address != "":
			parts = append(parts, fmt.Sprintf("%s <%s>", a.Name, a.Address))
		case a.Address != "":
			parts = append(parts, a.Address)
		case a.Name != "":
			parts = append(parts, a.Name)
		}
	}
	return strings.Join(parts, ", ")
}
