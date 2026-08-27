package mailexport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var when = time.Date(2026, 3, 7, 14, 5, 9, 0, time.UTC)

func TestFileName(t *testing.T) {
	meta := Meta{
		Date:      when,
		Subject:   "Invoice 42",
		From:      "billing@example.com",
		MessageID: "<abc123@example.com>",
	}
	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"default", "", "2026-03-07_Invoice 42"},
		{"date and time", "{date}-{time}", "2026-03-07-140509"},
		{"sender", "{from} {subject}", "billing@example.com Invoice 42"},
		{"message id, brackets dropped", "{id}", "abc123@example.com"},
		{"literal text is kept", "mail-{date}", "mail-2026-03-07"},
		{"unknown placeholder is literal", "{nope}-{date}", "{nope}-2026-03-07"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileName(meta, tt.template); got != tt.want {
				t.Errorf("FileName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFileNameCannotEscapeTheDirectory is the one that matters: the subject is
// attacker-controlled text arriving over the network.
func TestFileNameCannotEscapeTheDirectory(t *testing.T) {
	tests := []string{
		"../../etc/passwd",
		`..\..\windows\system32`,
		"a/b/c",
		"C:\\secrets",
	}
	for _, subject := range tests {
		got := FileName(Meta{Date: when, Subject: subject}, "{subject}")
		if strings.ContainsAny(got, `/\`) {
			t.Errorf("FileName(%q) = %q, still holds a separator", subject, got)
		}
		if strings.HasPrefix(got, ".") {
			t.Errorf("FileName(%q) = %q, starts with a dot", subject, got)
		}
	}
}

func TestFileNameEdgeCases(t *testing.T) {
	t.Run("no subject falls back to date and time", func(t *testing.T) {
		got := FileName(Meta{Date: when}, "{subject}")
		if got != "2026-03-07_140509" {
			t.Errorf("FileName() = %q, want the timestamp fallback", got)
		}
	})
	t.Run("no date uses now rather than year zero", func(t *testing.T) {
		got := FileName(Meta{Subject: "hi"}, "{date}")
		if strings.HasPrefix(got, "0000") || strings.HasPrefix(got, "0001") {
			t.Errorf("FileName() = %q, want today's date", got)
		}
	})
	t.Run("long subject is capped", func(t *testing.T) {
		got := FileName(Meta{Date: when, Subject: strings.Repeat("x", 500)}, "{subject}")
		if len(got) > maxNameBytes {
			t.Errorf("FileName() is %d bytes, want at most %d", len(got), maxNameBytes)
		}
	})
	t.Run("multibyte subject is not cut mid rune", func(t *testing.T) {
		got := FileName(Meta{Date: when, Subject: strings.Repeat("ü", 300)}, "{subject}")
		if strings.ContainsRune(got, '\uFFFD') {
			t.Errorf("FileName() = %q, holds a broken rune", got)
		}
	})
	t.Run("reserved windows name is stepped around", func(t *testing.T) {
		if got := FileName(Meta{Date: when, Subject: "NUL"}, "{subject}"); got == "NUL" {
			t.Error("FileName() returned the reserved name NUL unchanged")
		}
	})
	t.Run("newlines in a subject collapse", func(t *testing.T) {
		got := FileName(Meta{Date: when, Subject: "one\r\n two\tthree"}, "{subject}")
		if got != "one two three" {
			t.Errorf("FileName() = %q, want %q", got, "one two three")
		}
	})
}

func TestTargetDir(t *testing.T) {
	meta := Meta{Date: when}
	tests := []struct {
		mode string
		want string
	}{
		{SubfoldersNone, filepath.Join("base")},
		{SubfoldersYear, filepath.Join("base", "2026")},
		{SubfoldersMonth, filepath.Join("base", "2026", "03")},
		{"nonsense", filepath.Join("base")},
	}
	for _, tt := range tests {
		o := Options{Dir: "base", Subfolders: tt.mode}
		if got := o.TargetDir(meta); got != tt.want {
			t.Errorf("TargetDir(%q) = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	o := Options{Dir: dir, Subfolders: SubfoldersMonth}
	meta := Meta{Date: when, Subject: "Invoice 42"}

	path, err := o.Write(meta, []byte("From: a@b\r\n\r\nhi"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	want := filepath.Join(dir, "2026", "03", "2026-03-07_Invoice 42.eml")
	if path != want {
		t.Errorf("Write() path = %q, want %q", path, want)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(body) != "From: a@b\r\n\r\nhi" {
		t.Errorf("file holds %q, want the raw source unchanged", body)
	}
}

// TestWriteNeverOverwrites covers two messages that share a date and a subject,
// which is ordinary for receipts and notifications.
func TestWriteNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	o := Options{Dir: dir}
	meta := Meta{Date: when, Subject: "Receipt"}

	first, err := o.Write(meta, []byte("one"))
	if err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	second, err := o.Write(meta, []byte("two"))
	if err != nil {
		t.Fatalf("second Write() error = %v", err)
	}
	if first == second {
		t.Fatalf("both writes returned %q", first)
	}
	if filepath.Base(second) != "2026-03-07_Receipt-2.eml" {
		t.Errorf("second file is %q, want the -2 suffix", filepath.Base(second))
	}
	body, err := os.ReadFile(first)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(body) != "one" {
		t.Errorf("the first file now holds %q", body)
	}
}

func TestWriteRefusesWithoutADirectory(t *testing.T) {
	if _, err := (Options{}).Write(Meta{Date: when}, []byte("x")); err == nil {
		t.Error("Write() with no directory returned no error")
	}
}

func TestWriteRefusesEmptySource(t *testing.T) {
	if _, err := (Options{Dir: t.TempDir()}).Write(Meta{Date: when}, nil); err == nil {
		t.Error("Write() with no bytes returned no error")
	}
}
