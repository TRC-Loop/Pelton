package logging

import (
	"slices"
	"strconv"
	"strings"
	"sync"
)

// placeholder replaces a secret wherever it is found in a log line.
const placeholder = "[redacted]"

// Redactor removes known secrets from text on its way to a log file.
//
// It matches registered values, not field names or patterns. A pattern has to
// guess which strings are secret, and everything it guesses wrong is a password
// written to disk. The credentials package registers each secret it hands out
// instead, so a value is removed wherever it turns up, including in the middle
// of an error string from a server that echoed it back.
type Redactor struct {
	mu sync.RWMutex
	// values are the literal strings to remove, longest first so a secret that
	// contains another one is replaced whole.
	values []string
}

// Add registers a secret. Empty values are ignored. Adding the same value twice
// is a no-op, so callers can register on every load without growing the list.
//
// The escaped form is registered alongside the raw one: slog's text and json
// handlers quote strings, so a password with a quote or a backslash reaches the
// writer escaped and would not match its own bytes.
func (r *Redactor) Add(secret string) {
	if secret == "" {
		return
	}
	r.add(secret)
	if escaped := strings.Trim(strconv.Quote(secret), `"`); escaped != secret {
		r.add(escaped)
	}
}

func (r *Redactor) add(value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if slices.Contains(r.values, value) {
		return
	}
	r.values = append(r.values, value)
	// longest first: a shorter secret that happens to be a substring of a
	// longer one must not chop the longer one into unrecognizable pieces.
	for i := len(r.values) - 1; i > 0 && len(r.values[i]) > len(r.values[i-1]); i-- {
		r.values[i], r.values[i-1] = r.values[i-1], r.values[i]
	}
}

// Redact returns s with every registered secret replaced by [redacted].
func (r *Redactor) Redact(s string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, value := range r.values {
		s = strings.ReplaceAll(s, value, placeholder)
	}
	return s
}

// Reset drops every registered secret. Only for tests.
func (r *Redactor) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values = nil
}

// secrets is the process-wide redactor. Logging is process-wide and so is the
// set of secrets the process holds, and threading a redactor through every
// package that might one day log an error would be a lot of plumbing for one
// list.
var secrets Redactor

// Register adds a secret to the process-wide redactor, so it can never appear
// in a log or crash file. internal/credentials calls this for every secret it
// stores or loads.
func Register(secret string) {
	secrets.Add(secret)
}

// Redact returns s with every registered secret removed. Exported for callers
// that build text outside the log writer, like the diagnostics summary.
func Redact(s string) string {
	return secrets.Redact(s)
}
