// Package mailexport writes messages out as .eml files: the raw rfc 822 source
// exactly as the server holds it, under a name built from the message's own
// metadata. It is used by the export-on-archive option, so a message can leave
// the server later without leaving the user's disk.
//
// The package is pure: it takes bytes and metadata and touches nothing but the
// directory it is told to write to. Deciding what to export, and fetching the
// source, is the caller's job.
package mailexport

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// Subfolder modes decide how exported files are grouped under the base
// directory. A mailbox archived over years turns into an unusable directory
// otherwise.
const (
	SubfoldersNone  = "none"
	SubfoldersYear  = "year"
	SubfoldersMonth = "month"
)

// DefaultTemplate is the file name used when the account has none set: the
// date first so the directory sorts chronologically, then the subject.
const DefaultTemplate = "{date}_{subject}"

// maxNameBytes caps the file name before the extension. The limit on every
// filesystem we target is 255 bytes, and the collision suffix needs room.
const maxNameBytes = 200

// Meta is what a file name can be built from. Every field may be empty, which
// is what mail with no subject or no Message-ID gives us.
type Meta struct {
	// Date is the message's own date, used for both the name and the
	// year/month subfolder. A zero date falls back to the export time.
	Date      time.Time
	Subject   string
	From      string
	MessageID string
}

// Options is one account's export configuration.
type Options struct {
	// Dir is the base directory. An empty Dir means export is not configured
	// and Write refuses rather than guessing a location.
	Dir string
	// Subfolders is one of the Subfolders constants. Anything else is treated
	// as SubfoldersNone.
	Subfolders string
	// Template is the file name pattern. Empty means DefaultTemplate.
	Template string
}

// FileName renders a template into a file name, without the .eml extension.
// The placeholders are {date}, {time}, {from}, {subject} and {id}; anything
// else in the template is kept literally. Every substituted value is sanitized
// for use as a single path element, so a subject can never introduce a
// separator or climb out of the directory. A template that renders to nothing
// falls back to the date and time, which is never empty.
func FileName(meta Meta, template string) string {
	when := meta.Date
	if when.IsZero() {
		when = time.Now()
	}
	if strings.TrimSpace(template) == "" {
		template = DefaultTemplate
	}
	replacer := strings.NewReplacer(
		"{date}", when.Format("2006-01-02"),
		"{time}", when.Format("150405"),
		"{from}", sanitize(meta.From),
		"{subject}", sanitize(meta.Subject),
		"{id}", sanitize(strings.Trim(meta.MessageID, "<>")),
	)
	name := trimName(collapse(sanitize(replacer.Replace(template))))
	if name == "" {
		return when.Format("2006-01-02_150405")
	}
	return name
}

// TargetDir returns the directory a message with this metadata belongs in,
// creating nothing. It is Dir plus the year or month subfolder when one is
// configured.
func (o Options) TargetDir(meta Meta) string {
	when := meta.Date
	if when.IsZero() {
		when = time.Now()
	}
	switch o.Subfolders {
	case SubfoldersYear:
		return filepath.Join(o.Dir, when.Format("2006"))
	case SubfoldersMonth:
		return filepath.Join(o.Dir, when.Format("2006"), when.Format("01"))
	default:
		return o.Dir
	}
}

// Write saves raw under the configured directory and returns the path written.
// It creates the subfolder as needed and never overwrites: an existing name
// gains a -2, -3 suffix, because two different messages sharing a date and
// subject is ordinary and losing one of them silently is not.
func (o Options) Write(meta Meta, raw []byte) (string, error) {
	if strings.TrimSpace(o.Dir) == "" {
		return "", fmt.Errorf("mailexport: no export directory configured")
	}
	if len(raw) == 0 {
		return "", fmt.Errorf("mailexport: nothing to write")
	}
	dir := o.TargetDir(meta)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("mailexport: create %s: %w", dir, err)
	}
	base := FileName(meta, o.Template)
	for n := 1; ; n++ {
		name := base + ".eml"
		if n > 1 {
			name = fmt.Sprintf("%s-%d.eml", base, n)
		}
		path := filepath.Join(dir, name)
		// O_EXCL so two exports racing for the same name cannot both win.
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("mailexport: create %s: %w", path, err)
		}
		if _, err := f.Write(raw); err != nil {
			f.Close()
			os.Remove(path)
			return "", fmt.Errorf("mailexport: write %s: %w", path, err)
		}
		if err := f.Close(); err != nil {
			return "", fmt.Errorf("mailexport: close %s: %w", path, err)
		}
		return path, nil
	}
}

// reservedNames are the device names Windows refuses to use as a file name, in
// any directory and with any extension.
var reservedNames = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com1": true, "com2": true, "com3": true, "com4": true, "com5": true,
	"com6": true, "com7": true, "com8": true, "com9": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true, "lpt5": true,
	"lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
}

// sanitize reduces a value to something safe as one path element on every
// platform: no separators, no characters Windows rejects, no control runes.
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '/' || r == '\\' || r == ':':
			b.WriteRune('-')
		case strings.ContainsRune(`<>"|?*`, r):
			b.WriteRune('-')
		case unicode.IsControl(r):
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// collapse squeezes runs of whitespace into single spaces, so a wrapped or
// padded subject does not become a file name full of gaps.
func collapse(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// trimName enforces the length cap on a whole rune, drops the leading dots and
// trailing dots or spaces that Windows and the shell both dislike, and steps
// around the reserved device names.
func trimName(s string) string {
	for len(s) > maxNameBytes {
		_, size := lastRune(s)
		s = s[:len(s)-size]
	}
	s = strings.Trim(s, " .")
	if s == "" {
		return ""
	}
	if reservedNames[strings.ToLower(s)] {
		return s + "_"
	}
	return s
}

// lastRune returns the final rune of a non-empty string and its width.
func lastRune(s string) (rune, int) {
	runes := []rune(s)
	last := runes[len(runes)-1]
	return last, len(string(last))
}
