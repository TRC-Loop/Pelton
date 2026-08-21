package charsetguess_test

import (
	"io"
	"strings"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"

	"github.com/TRC-Loop/Pelton/internal/charsetguess"
)

func TestDecodeAlwaysProducesValidUTF8(t *testing.T) {
	cases := []struct {
		name string
		raw  []byte
		want string
	}{
		{"already utf-8", []byte("café"), "café"},
		{"latin-1 bytes", latin1(t, "Ich möchte einen größeren Kaffee, bitte, und danach gehen wir spazieren."), "Ich möchte einen größeren Kaffee, bitte, und danach gehen wir spazieren."},
		{"shift_jis bytes", shiftJIS(t, "こんにちは世界。これは日本語のメールです。文字化けしないことを確認します。"), "こんにちは世界。これは日本語のメールです。文字化けしないことを確認します。"},
		{"empty", nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, _ := charsetguess.Decode(c.raw)
			if !utf8.ValidString(got) {
				t.Fatalf("decoded text is not valid utf-8: %q", got)
			}
			if got != c.want {
				t.Errorf("Decode() = %q, want %q", got, c.want)
			}
		})
	}
}

// a body of random bytes decodes to something rather than failing, and whatever
// it decodes to is storable. This is the guarantee the search index depends on.
func TestDecodeGarbageIsStillValidUTF8(t *testing.T) {
	raw := []byte{0xff, 0xfe, 0x00, 0x81, 0x8d, 0x90, 0xc0, 0xc1}
	got, name := charsetguess.Decode(raw)
	if !utf8.ValidString(got) {
		t.Fatalf("decoded text is not valid utf-8: %q", got)
	}
	if name == "" {
		t.Error("Decode() reported no guess for undecodable bytes")
	}
}

func TestDecodeReportsGuessOnlyWhenItGuessed(t *testing.T) {
	if _, name := charsetguess.Decode([]byte("plain ascii")); name != "" {
		t.Errorf("Decode() named %q for text that needed no guess", name)
	}
	if _, name := charsetguess.Decode(latin1(t, "Grüße aus Köln, die Tür war offen und der Kaffee war größer.")); name == "" {
		t.Error("Decode() named nothing for text it had to guess at")
	}
}

func TestFallbackOverridesDetection(t *testing.T) {
	t.Cleanup(func() { charsetguess.SetFallback(charsetguess.Auto) })

	raw := latin1(t, "Grüße")
	charsetguess.SetFallback("iso-8859-5")
	got, name := charsetguess.Decode(raw)
	if name != "iso-8859-5" {
		t.Fatalf("Decode() used %q, want the configured iso-8859-5", name)
	}
	if got == "Grüße" {
		t.Error("Decode() detected anyway; the setting was ignored")
	}
	if !utf8.ValidString(got) {
		t.Errorf("decoded text is not valid utf-8: %q", got)
	}
}

func TestSetFallbackRejectsNonsense(t *testing.T) {
	t.Cleanup(func() { charsetguess.SetFallback(charsetguess.Auto) })

	for _, name := range []string{"", "x-nonsense", "auto"} {
		charsetguess.SetFallback(name)
		if got := charsetguess.Fallback(); got != charsetguess.Auto {
			t.Errorf("SetFallback(%q) left the fallback at %q, want auto", name, got)
		}
	}
}

// the hook go-message calls. An unknown name must not come back as an error,
// because an error is what made the caller keep the raw bytes.
func TestReaderNeverFails(t *testing.T) {
	for _, name := range []string{"iso-8859-1", "x-nonsense", "", "cp1252"} {
		r, err := charsetguess.Reader(name, strings.NewReader("caf\xe9"))
		if err != nil {
			t.Fatalf("Reader(%q) errored: %v", name, err)
		}
		text, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("Reader(%q) read: %v", name, err)
		}
		if !utf8.Valid(text) {
			t.Errorf("Reader(%q) produced invalid utf-8: %q", name, text)
		}
	}
}

func TestKnown(t *testing.T) {
	for _, name := range []string{"utf-8", "us-ascii", "iso-8859-1", "windows-1252", "koi8-r"} {
		if !charsetguess.Known(name) {
			t.Errorf("Known(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"", "x-nonsense", "not a charset"} {
		if charsetguess.Known(name) {
			t.Errorf("Known(%q) = true, want false", name)
		}
	}
}

func latin1(t *testing.T, s string) []byte {
	t.Helper()
	b, err := charmap.Windows1252.NewEncoder().Bytes([]byte(s))
	if err != nil {
		t.Fatalf("encode latin-1: %v", err)
	}
	return b
}

func shiftJIS(t *testing.T, s string) []byte {
	t.Helper()
	b, err := japanese.ShiftJIS.NewEncoder().Bytes([]byte(s))
	if err != nil {
		t.Fatalf("encode shift_jis: %v", err)
	}
	return b
}
