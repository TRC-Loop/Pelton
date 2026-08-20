package mailimport

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

// MboxReader walks an mbox file, returning one raw message per call. mbox has
// no length prefix or framing: messages are separated by a line beginning
// "From " that sits at the start of the file or directly after a blank line.
// A body line that would look like a separator is escaped on write by
// prefixing it with ">", which Next undoes.
//
// It reads one message into memory at a time rather than the whole file, since
// an archive exported from another client is routinely gigabytes.
type MboxReader struct {
	br *bufio.Reader
	// pending holds the "From " line that ended the previous message, since
	// finding it is what tells us the previous one is complete.
	pending []byte
	// afterBlank tracks whether the last line written was empty, which is half
	// of what makes a "From " line a separator rather than body text.
	afterBlank bool
	started    bool
}

// mboxBufSize is the read buffer. Messages routinely carry base64 lines and
// long headers; the buffer only affects read syscalls, not the maximum line.
const mboxBufSize = 64 * 1024

// NewMboxReader reads mbox messages from r.
func NewMboxReader(r io.Reader) *MboxReader {
	return &MboxReader{br: bufio.NewReaderSize(r, mboxBufSize)}
}

// Next returns the next message's raw source, or io.EOF when the file is done.
// A file that is not mbox at all (no separator line anywhere) yields no
// messages and returns ErrNotMbox, so the caller can say so rather than
// silently importing nothing.
func (m *MboxReader) Next() ([]byte, error) {
	var msg bytes.Buffer

	if m.pending != nil {
		m.pending = nil
	} else if !m.started {
		// the first separator has to be found before anything is a message.
		if err := m.seekFirstSeparator(); err != nil {
			return nil, err
		}
	}
	m.started = true
	m.afterBlank = false

	for {
		line, err := m.readLine()
		if err == io.EOF {
			if msg.Len() == 0 {
				return nil, io.EOF
			}
			return msg.Bytes(), nil
		}
		if err != nil {
			return nil, err
		}

		if m.afterBlank && isSeparator(line) {
			m.pending = line
			return msg.Bytes(), nil
		}
		m.afterBlank = isBlank(line)
		msg.Write(unescapeFrom(line))
	}
}

// ErrNotMbox is returned when a file handed to MboxReader has no mbox
// separator line at all, which means it is some other format.
var ErrNotMbox = errors.New("mailimport: not an mbox file")

// seekFirstSeparator consumes everything up to and including the first
// separator line. A well-formed mbox starts with one on line 1.
func (m *MboxReader) seekFirstSeparator() error {
	for {
		line, err := m.readLine()
		if err == io.EOF {
			return ErrNotMbox
		}
		if err != nil {
			return err
		}
		if isSeparator(line) {
			return nil
		}
		if !isBlank(line) {
			// content before any separator means this is not an mbox; stopping
			// here avoids scanning a whole large file to reach the same answer.
			return ErrNotMbox
		}
	}
}

// readLine reads one line including its terminator. bufio.Scanner is not used
// because its token limit would fail on messages with very long single lines.
func (m *MboxReader) readLine() ([]byte, error) {
	line, err := m.br.ReadBytes('\n')
	if err == io.EOF && len(line) > 0 {
		return line, nil
	}
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("mailimport: read mbox: %w", err)
	}
	return line, nil
}

var fromPrefix = []byte("From ")

func isSeparator(line []byte) bool {
	return bytes.HasPrefix(line, fromPrefix)
}

func isBlank(line []byte) bool {
	return len(bytes.TrimRight(line, "\r\n")) == 0
}

// unescapeFrom reverses the write-side quoting of body lines that would
// otherwise read as separators. mboxrd escapes any run of ">" before "From ",
// so exactly one ">" comes back off.
func unescapeFrom(line []byte) []byte {
	i := 0
	for i < len(line) && line[i] == '>' {
		i++
	}
	if i == 0 || !bytes.HasPrefix(line[i:], fromPrefix) {
		return line
	}
	return line[1:]
}
