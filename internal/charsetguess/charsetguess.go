// Package charsetguess makes sure text taken out of a message is valid UTF-8,
// including when the message is wrong about its own encoding.
//
// A well formed mail names its charset and go-message converts it. Plenty of
// real mail does not: it names nothing, which the rfc says means us-ascii while
// the body is plainly not, or it names something no table knows. Left alone
// those bytes travel unconverted into the database, the search index and the
// reading pane, which is where mojibake comes from. Other clients guess instead
// of giving up, and so does this: the bytes are run past a detector ported from
// ICU, and whatever comes back is decoded and cleaned so that nothing invalid
// leaves the parser.
//
// Importing the package replaces go-message's charset hook, so both bodies and
// rfc 2047 encoded-words in headers come through here.
package charsetguess

import (
	"bytes"
	"io"
	"mime"
	"strings"
	"sync/atomic"
	"unicode/utf8"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/charset"
	"github.com/gogs/chardet"
)

// Auto is the fallback setting that detects from the bytes. It is the default.
const Auto = "auto"

// LastResort is decoded from when detection is off or produces nothing usable.
// It maps every byte to a character, so it cannot fail, and it is right for
// most of the mail that reaches this path.
const LastResort = "windows-1252"

// minConfidence is what the detector has to report before its answer is taken.
// Below it the guess is worse than the last resort: a short body has too few
// bytes for the statistics to mean anything, and ICU's scale reflects that.
const minConfidence = 30

// replacement stands in for a byte that survived every decode attempt.
const replacement = "�"

// fallback is the user's setting: Auto, or the name of an encoding to decode
// undeclared and unknown text as. It is read on every message, so it is an
// atomic rather than something the parser has to be handed.
var fallback atomic.Value

func init() {
	fallback.Store(Auto)
	// go-message/charset set this in its own init; ours knows the same table
	// and, unlike it, always answers.
	message.CharsetReader = Reader
}

// SetFallback records how text with no usable charset should be read. An empty
// or unrecognised value means Auto.
func SetFallback(name string) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || name == Auto {
		fallback.Store(Auto)
		return
	}
	if _, err := charset.Reader(name, strings.NewReader("")); err != nil {
		fallback.Store(Auto)
		return
	}
	fallback.Store(name)
}

// Fallback returns the current setting.
func Fallback() string {
	name, _ := fallback.Load().(string)
	if name == "" {
		return Auto
	}
	return name
}

// Reader converts a named charset to UTF-8. It replaces go-message's hook, so
// it is called for every text part and every encoded-word that names something
// other than utf-8 or us-ascii.
//
// It never reports an error. go-message treats a charset error as reason to
// hand back the bytes untouched, which is the whole bug: a name nobody knows is
// answered here with a detected one instead.
func Reader(name string, input io.Reader) (io.Reader, error) {
	if r, err := charset.Reader(name, input); err == nil {
		return r, nil
	}
	raw, err := io.ReadAll(input)
	if err != nil {
		return input, nil
	}
	text, _ := Decode(raw)
	return strings.NewReader(text), nil
}

// WordDecoder decodes rfc 2047 encoded-words through the same path, for
// libraries that take their own decoder rather than go-message's hook.
func WordDecoder() *mime.WordDecoder {
	return &mime.WordDecoder{CharsetReader: Reader}
}

// Known reports whether a charset name can be decoded by name alone. rfc822
// uses it to tell a message that was right about itself from one whose text had
// to be guessed at.
func Known(name string) bool {
	if name == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "utf-8", "utf8", "us-ascii", "ascii":
		return true
	}
	_, err := charset.Reader(name, strings.NewReader(""))
	return err == nil
}

// Decode turns bytes of unknown encoding into UTF-8 and names what it read them
// as. Bytes that are already valid UTF-8 are returned unchanged with an empty
// name, since nothing was guessed and there is nothing to tell the reader.
func Decode(raw []byte) (string, string) {
	if utf8.Valid(raw) {
		return string(raw), ""
	}
	name := Fallback()
	if name == Auto {
		name = detect(raw)
	}
	if decoded, err := charset.Reader(name, bytes.NewReader(raw)); err == nil {
		if text, err := io.ReadAll(decoded); err == nil && utf8.Valid(text) {
			return string(text), name
		}
	}
	if name != LastResort {
		if decoded, err := charset.Reader(LastResort, bytes.NewReader(raw)); err == nil {
			if text, err := io.ReadAll(decoded); err == nil && utf8.Valid(text) {
				return string(text), LastResort
			}
		}
	}
	return strings.ToValidUTF8(string(raw), replacement), LastResort
}

// Text is Decode over a string, for text that has already been read but may
// never have been converted.
func Text(s string) (string, string) {
	if utf8.ValidString(s) {
		return s, ""
	}
	return Decode([]byte(s))
}

// Valid returns s with anything that is not valid UTF-8 replaced. It is the
// last line before storage: whatever a guess concludes, a wrong character is
// recoverable and a broken byte is not, because the database and the search
// index both refuse to work with one.
func Valid(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.ToValidUTF8(s, replacement)
}

// detect asks the ICU detector what the bytes look like. It returns the last
// resort when the answer is missing, too weak to act on, or names something no
// table here can decode.
func detect(raw []byte) string {
	result, err := chardet.NewTextDetector().DetectBest(raw)
	if err != nil || result == nil || result.Confidence < minConfidence {
		return LastResort
	}
	name := strings.ToLower(result.Charset)
	// latin-1 and windows-1252 differ only in the c1 range, where latin-1 has
	// control characters and windows-1252 has quotes and dashes. Mail that
	// reaches this path and uses those bytes means the punctuation, so the
	// detector's latin-1 is read as the superset. Browsers do the same.
	if name == "iso-8859-1" {
		return LastResort
	}
	if _, err := charset.Reader(name, strings.NewReader("")); err != nil {
		return LastResort
	}
	return name
}
