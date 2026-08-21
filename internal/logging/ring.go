package logging

import (
	"strings"
	"sync"
)

// Ring keeps the most recent log lines in memory so something can read them
// back without opening a file, which the developer overlays do (#188). It is
// fed the same redacted bytes that reach stderr and the log file, so a secret
// stripped from one is stripped from all three.
//
// Nothing is written to disk here and nothing leaves the process. The buffer is
// fixed size: once full, the oldest line is dropped rather than the app growing
// a memory leak on a long session.
type Ring struct {
	mu    sync.Mutex
	lines []Line
	// next is the sequence number the next line gets. It counts every line ever
	// added, including the ones already dropped, so a reader can tell that it
	// missed some rather than silently seeing a gap.
	next  uint64
	first int
	count int
}

// Line is one buffered log line with the sequence number it was given.
type Line struct {
	Seq  uint64 `json:"seq"`
	Text string `json:"text"`
}

// NewRing returns a Ring holding at most capacity lines. A capacity below one
// is treated as one.
func NewRing(capacity int) *Ring {
	if capacity < 1 {
		capacity = 1
	}
	return &Ring{lines: make([]Line, capacity)}
}

// Add buffers one line, dropping the oldest when full. Trailing newlines are
// stripped, since a reader wants the line rather than the framing.
func (r *Ring) Add(text string) {
	text = strings.TrimRight(text, "\r\n")
	if text == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	pos := (r.first + r.count) % len(r.lines)
	r.lines[pos] = Line{Seq: r.next, Text: text}
	r.next++
	if r.count == len(r.lines) {
		r.first = (r.first + 1) % len(r.lines)
		return
	}
	r.count++
}

// Since returns the buffered lines with a sequence number at or after seq, plus
// the sequence to ask for next. Passing 0 returns everything still buffered.
//
// Lines that were dropped before the reader got to them are simply not there;
// the gap is visible in the sequence numbers, which is the point of keeping
// them.
func (r *Ring) Since(seq uint64) ([]Line, uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]Line, 0, r.count)
	for i := range r.count {
		line := r.lines[(r.first+i)%len(r.lines)]
		if line.Seq >= seq {
			out = append(out, line)
		}
	}
	return out, r.next
}

// Clear empties the buffer. Sequence numbers keep counting, so a reader holding
// an old one is not sent back through lines it already has.
func (r *Ring) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.first = 0
	r.count = 0
}
