// Package logging owns Pelton's log output: where it goes, what it is allowed
// to contain, and what is left behind when the app panics.
//
// Nothing here talks to the network. Logs are files in the user's own data
// directory and stay there; getting one to a bug report is the user copying it,
// deliberately. There is no reporter, no endpoint and no upload path to turn on.
//
// File logging is off by default. Until it is switched on the app logs to
// stderr exactly as it always has, and the log directory is not even created.
package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

// DirName is the log directory, created inside the app data directory the
// database lives in.
const DirName = "logs"

// Dir returns the log directory for a given app data directory.
func Dir(dataDir string) string {
	return filepath.Join(dataDir, DirName)
}

// Writer is the destination every log line goes through. It always writes to
// stderr and, once Enable has been called, also to a rotating file.
//
// Redaction happens here rather than at the handler, so it covers the finished
// line: an attribute value, an error string a server put a password into, and
// anything a future handler decides to add are all just bytes by this point.
type Writer struct {
	mu     sync.Mutex
	stderr io.Writer
	file   *rotator
	// level is referenced by the handler options, so changing it takes effect
	// on loggers that were already built.
	level slog.LevelVar
}

// NewWriter returns a Writer logging to stderr only, at info level.
func NewWriter() *Writer {
	w := &Writer{stderr: os.Stderr}
	w.level.Set(slog.LevelInfo)
	return w
}

// Logger builds the app logger. Its level and file destination stay adjustable
// through the Writer afterwards, so the settings toggle does not have to reach
// every package holding a logger.
func (w *Writer) Logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: &w.level}))
}

// SetLevel changes the threshold for every logger built from this Writer.
func (w *Writer) SetLevel(level slog.Level) {
	w.level.Set(level)
}

// Level reports the current threshold.
func (w *Writer) Level() slog.Level {
	return w.level.Level()
}

// Enable starts writing to a rotating file in dir, creating it if needed.
// Calling it again with the same directory is a no-op; with a different one it
// switches files.
func (w *Writer) Enable(dir string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		if w.file.dir == dir {
			return nil
		}
		w.file.Close()
		w.file = nil
	}
	rot, err := openRotator(dir)
	if err != nil {
		return err
	}
	w.file = rot
	return nil
}

// Disable stops file logging. stderr keeps working. It does not delete
// anything already written; RemoveLogs does that.
func (w *Writer) Disable() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

// Enabled reports whether lines are currently reaching a file.
func (w *Writer) Enabled() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file != nil
}

// Write redacts p and forwards it to stderr and, when enabled, the log file. It
// reports the length of the input rather than of what it wrote, because
// redaction changes the length and a short count would look like a failed write
// to slog.
func (w *Writer) Write(p []byte) (int, error) {
	clean := []byte(secrets.Redact(string(p)))

	w.mu.Lock()
	file := w.file
	stderr := w.stderr
	w.mu.Unlock()

	if stderr != nil {
		stderr.Write(clean)
	}
	if file != nil {
		if _, err := file.Write(clean); err != nil {
			return len(p), err
		}
	}
	return len(p), nil
}

// RemoveLogs deletes the log files in dir. Turning logging off, or deleting the
// account that produced the lines, should offer this: an old log is still a
// record of what the user was doing.
func RemoveLogs(dir string) error {
	return removeLogs(dir)
}

// messageMetadata gates the louder logging of per-message identifiers (uids,
// subjects, senders) that sync debugging needs and normal logging must not
// write. It is a separate opt-in from file logging itself.
//
// Process-wide for the same reason the redactor is: it is one policy for the
// whole process, and passing it down every constructor to express a single bool
// would touch every call site.
var messageMetadata atomic.Bool

// SetMessageMetadata turns per-message logging on or off.
func SetMessageMetadata(on bool) {
	messageMetadata.Store(on)
}

// MessageMetadata reports whether per-message identifiers may be logged.
// Callers with a message subject, sender or uid to log must check it first.
func MessageMetadata() bool {
	return messageMetadata.Load()
}

// ParseLevel maps a stored setting value to a slog level, falling back to info
// for anything unrecognized.
func ParseLevel(name string) slog.Level {
	switch strings.ToLower(name) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// LevelName is the inverse of ParseLevel, for reporting the level back to the
// settings ui.
func LevelName(level slog.Level) string {
	switch {
	case level <= slog.LevelDebug:
		return "debug"
	case level >= slog.LevelError:
		return "error"
	case level >= slog.LevelWarn:
		return "warn"
	default:
		return "info"
	}
}
