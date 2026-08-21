package logging

import (
	"log/slog"
	"strings"
	"testing"
)

func TestRingKeepsRecentLines(t *testing.T) {
	r := NewRing(3)
	for _, line := range []string{"one", "two", "three"} {
		r.Add(line)
	}

	lines, next := r.Since(0)
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if lines[0].Text != "one" || lines[2].Text != "three" {
		t.Errorf("lines = %v, want one..three in order", lines)
	}
	if next != 3 {
		t.Errorf("next = %d, want 3", next)
	}
}

// The buffer is fixed size on purpose: a long session must not grow it.
func TestRingDropsOldestWhenFull(t *testing.T) {
	r := NewRing(2)
	for _, line := range []string{"one", "two", "three", "four"} {
		r.Add(line)
	}

	lines, next := r.Since(0)
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	if lines[0].Text != "three" || lines[1].Text != "four" {
		t.Errorf("lines = %v, want the last two", lines)
	}
	// the sequence keeps counting through the dropped lines, so a reader can
	// see that two went missing rather than believing it has everything.
	if lines[0].Seq != 2 {
		t.Errorf("first Seq = %d, want 2", lines[0].Seq)
	}
	if next != 4 {
		t.Errorf("next = %d, want 4", next)
	}
}

func TestRingSinceReturnsOnlyNewLines(t *testing.T) {
	r := NewRing(10)
	r.Add("one")
	r.Add("two")

	_, next := r.Since(0)
	r.Add("three")

	lines, _ := r.Since(next)
	if len(lines) != 1 || lines[0].Text != "three" {
		t.Errorf("lines = %v, want only three", lines)
	}
}

func TestRingClearKeepsCounting(t *testing.T) {
	r := NewRing(10)
	r.Add("one")
	r.Clear()

	lines, next := r.Since(0)
	if len(lines) != 0 {
		t.Errorf("lines = %v, want none after Clear", lines)
	}
	if next != 1 {
		t.Errorf("next = %d, want the sequence to keep counting", next)
	}
}

// The overlay reads the same bytes the log file gets, so a password stripped
// from one has to be stripped from all of them.
func TestBufferedLinesAreRedacted(t *testing.T) {
	t.Cleanup(secrets.Reset)
	Register("hunter2")

	w := NewWriter()
	w.stderr = nil
	ring := w.Buffer(10)

	slog.New(slog.NewTextHandler(w, nil)).Info("login", "cmd", "LOGIN user hunter2")

	lines, _ := ring.Since(0)
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	if strings.Contains(lines[0].Text, "hunter2") {
		t.Errorf("the buffered line kept the password: %q", lines[0].Text)
	}
}

// Buffering is opt-in: a normal run must not allocate a buffer or collect
// anything in memory.
func TestWriterDoesNotBufferUntilAsked(t *testing.T) {
	w := NewWriter()
	if w.ring != nil {
		t.Error("a new Writer is already buffering")
	}
	if first, second := w.Buffer(5), w.Buffer(5); first != second {
		t.Error("Buffer started a second ring instead of returning the first")
	}
}
