package imap

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"

	// registers legacy charset decoders (ISO-8859-*, Windows-125x, ...)
	_ "github.com/emersion/go-message/charset"
)

// Mailbox summarises a selected mailbox.
type Mailbox struct {
	Name        string
	NumMessages uint32
	UIDNext     imap.UID
	// UIDValidity invalidates cached UIDs when it changes.
	UIDValidity uint32
}

// MessageHeader is the envelope-level summary used for listings.
type MessageHeader struct {
	SeqNum  uint32
	UID     imap.UID
	Subject string
	From    string
	To      string
	Date    time.Time
	Flags   []imap.Flag
}

// Message is a fully parsed message with extracted bodies and attachments.
type Message struct {
	UID         imap.UID
	MessageID   string // rfc Message-ID header, for threading and dedup
	Subject     string
	From        string
	To          string
	Cc          string
	Date        time.Time
	Flags       []imap.Flag
	Text        string
	HTML        string
	Size        int64 // raw rfc822 byte length
	Attachments []Attachment
}

// Attachment holds attachment metadata and its decoded content.
type Attachment struct {
	Filename    string
	ContentType string
	ContentID   string // set for inline cid-referenced parts
	Content     []byte
}

// Select opens a mailbox. IMAP selects one mailbox per connection, so this
// must precede the fetch and flag methods.
func (c *Client) Select(mailbox string) (*Mailbox, error) {
	// go-imap encodes non-ASCII names as modified UTF-7 when needed
	data, err := c.raw.Select(mailbox, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("imap: select %q: %w", mailbox, err)
	}
	return &Mailbox{
		Name:        mailbox,
		NumMessages: data.NumMessages,
		UIDNext:     data.UIDNext,
		UIDValidity: data.UIDValidity,
	}, nil
}

// FetchRecentHeaders returns up to limit recent messages, newest first. No
// bodies are fetched, so it stays cheap on large mailboxes.
func (c *Client) FetchRecentHeaders(limit int) ([]MessageHeader, error) {
	mbox := c.raw.Mailbox()
	if mbox == nil {
		return nil, fmt.Errorf("imap: no mailbox selected")
	}
	total := mbox.NumMessages
	if total == 0 || limit <= 0 {
		return nil, nil
	}

	// take the tail window by sequence number; each header also carries the
	// stable UID since sequence numbers shift as the mailbox changes
	var start uint32 = 1
	if uint32(limit) < total {
		start = total - uint32(limit) + 1
	}
	seqSet := imap.SeqSet{}
	seqSet.AddRange(start, total)

	options := &imap.FetchOptions{Envelope: true, Flags: true, UID: true}
	buffers, err := c.raw.Fetch(seqSet, options).Collect()
	if err != nil {
		return nil, fmt.Errorf("imap: fetch headers: %w", err)
	}

	headers := make([]MessageHeader, 0, len(buffers))
	for _, b := range buffers {
		headers = append(headers, headerFromBuffer(b))
	}
	reverse(headers) // ascending -> newest first
	return headers, nil
}

// FetchMessage fetches and parses a full message by UID.
func (c *Client) FetchMessage(uid imap.UID) (*Message, error) {
	// PEEK so reading the body does not set \Seen
	section := &imap.FetchItemBodySection{Peek: true}
	options := &imap.FetchOptions{
		Envelope:    true,
		Flags:       true,
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{section},
	}

	buffers, err := c.raw.Fetch(imap.UIDSetNum(uid), options).Collect()
	if err != nil {
		return nil, fmt.Errorf("imap: fetch message uid %d: %w", uid, err)
	}
	if len(buffers) == 0 {
		return nil, fmt.Errorf("imap: message uid %d not found", uid)
	}
	buf := buffers[0]

	raw := buf.FindBodySection(section)
	if raw == nil {
		return nil, fmt.Errorf("imap: message uid %d returned no body", uid)
	}

	msg := &Message{UID: buf.UID, Flags: buf.Flags, Size: int64(len(raw))}
	if buf.Envelope != nil {
		msg.MessageID = buf.Envelope.MessageID
		msg.Subject = buf.Envelope.Subject
		msg.From = formatAddresses(buf.Envelope.From)
		msg.To = formatAddresses(buf.Envelope.To)
		msg.Cc = formatAddresses(buf.Envelope.Cc)
		msg.Date = buf.Envelope.Date
	}

	if err := parseBody(raw, msg); err != nil {
		return nil, fmt.Errorf("imap: parse message uid %d: %w", uid, err)
	}
	return msg, nil
}

// FetchRawMessage returns a message's RFC 822 source by UID, exactly as the
// server stores it, undecoded and unparsed (PEEK, so reading it never sets
// \Seen).
func (c *Client) FetchRawMessage(uid imap.UID) ([]byte, error) {
	section := &imap.FetchItemBodySection{Peek: true}
	options := &imap.FetchOptions{
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{section},
	}

	buffers, err := c.raw.Fetch(imap.UIDSetNum(uid), options).Collect()
	if err != nil {
		return nil, fmt.Errorf("imap: fetch raw message uid %d: %w", uid, err)
	}
	if len(buffers) == 0 {
		return nil, fmt.Errorf("imap: message uid %d not found", uid)
	}
	raw := buffers[0].FindBodySection(section)
	if raw == nil {
		return nil, fmt.Errorf("imap: message uid %d returned no body", uid)
	}
	return raw, nil
}

// FetchAllFlags returns the UID and flags of every message in the selected
// mailbox and nothing else, so it stays cheap. sync diffs this against the
// local cache to find new, deleted and reflagged messages. over very large
// mailboxes CONDSTORE would let us ask only for what changed since last sync,
// but a full compare is correct and is enough for now.
func (c *Client) FetchAllFlags() ([]MessageHeader, error) {
	mbox := c.raw.Mailbox()
	if mbox == nil {
		return nil, fmt.Errorf("imap: no mailbox selected")
	}
	if mbox.NumMessages == 0 {
		return nil, nil
	}

	seqSet := imap.SeqSet{}
	seqSet.AddRange(1, mbox.NumMessages)
	options := &imap.FetchOptions{Flags: true, UID: true}
	buffers, err := c.raw.Fetch(seqSet, options).Collect()
	if err != nil {
		return nil, fmt.Errorf("imap: fetch all flags: %w", err)
	}

	headers := make([]MessageHeader, 0, len(buffers))
	for _, b := range buffers {
		headers = append(headers, MessageHeader{UID: b.UID, Flags: b.Flags})
	}
	return headers, nil
}

func headerFromBuffer(b *imapclient.FetchMessageBuffer) MessageHeader {
	h := MessageHeader{SeqNum: b.SeqNum, UID: b.UID, Flags: b.Flags}
	if b.Envelope != nil {
		h.Subject = b.Envelope.Subject
		h.From = formatAddresses(b.Envelope.From)
		h.To = formatAddresses(b.Envelope.To)
		h.Date = b.Envelope.Date
	}
	return h
}

// parseBody extracts text, HTML and attachment metadata from a raw message.
func parseBody(raw []byte, msg *Message) error {
	mr, err := mail.CreateReader(bytes.NewReader(raw))
	// unknown charset is non-fatal: reader is still usable
	if err != nil && !message.IsUnknownCharset(err) {
		return fmt.Errorf("create mail reader: %w", err)
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil && !message.IsUnknownCharset(err) {
			return fmt.Errorf("read part: %w", err)
		}

		switch header := part.Header.(type) {
		case *mail.InlineHeader:
			body, err := io.ReadAll(part.Body)
			if err != nil {
				return fmt.Errorf("read inline part: %w", err)
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
				return fmt.Errorf("read attachment part: %w", err)
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
