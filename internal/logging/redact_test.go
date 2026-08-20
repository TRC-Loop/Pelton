package logging

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRedactorRemovesRegisteredSecrets(t *testing.T) {
	tests := []struct {
		name    string
		secrets []string
		in      string
		want    string
	}{
		{
			name:    "plain value anywhere in the line",
			secrets: []string{"hunter2"},
			in:      "imap login failed for user: hunter2 rejected",
			want:    "imap login failed for user: [redacted] rejected",
		},
		{
			name:    "inside a url the server echoed back",
			secrets: []string{"s3cr3t"},
			in:      `dial imaps://arne:s3cr3t@mail.example.com:993`,
			want:    `dial imaps://arne:[redacted]@mail.example.com:993`,
		},
		{
			name:    "overlapping secrets do not chop each other",
			secrets: []string{"abc", "abcdef"},
			in:      "token abcdef end",
			want:    "token [redacted] end",
		},
		{
			name:    "empty registration is ignored",
			secrets: []string{""},
			in:      "nothing to remove",
			want:    "nothing to remove",
		},
		{
			name:    "escaped form matches too",
			secrets: []string{`pa"ss`},
			in:      `err="login failed: pa\"ss"`,
			want:    `err="login failed: [redacted]"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r Redactor
			for _, s := range tt.secrets {
				r.Add(s)
			}
			if got := r.Redact(tt.in); got != tt.want {
				t.Errorf("Redact() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestLoggedCredentialsNeverReachTheFile is the test the issue asks for: push
// known credentials through the real logger and assert the file does not hold
// them, including when they arrive inside an error rather than as a field.
func TestLoggedCredentialsNeverReachTheFile(t *testing.T) {
	secrets.Reset()
	t.Cleanup(secrets.Reset)

	const (
		password    = "correct-horse-battery"
		oauthToken  = "ya29.a0AfB_by-Fake-Token"
		appPassword = "abcd efgh ijkl mnop"
	)
	for _, s := range []string{password, oauthToken, appPassword} {
		Register(s)
	}

	dir := t.TempDir()
	w := NewWriter()
	w.stderr = nil
	if err := w.Enable(dir); err != nil {
		t.Fatalf("Enable() = %v", err)
	}
	log := w.Logger()

	log.Info("account added", "password", password)
	log.Error("oauth refresh", "err", errors.New("server said token "+oauthToken+" is invalid"))
	log.Info("app password stored", "value", appPassword)
	if err := w.Disable(); err != nil {
		t.Fatalf("Disable() = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, logFileName))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	body := string(data)
	for _, s := range []string{password, oauthToken, appPassword} {
		if strings.Contains(body, s) {
			t.Errorf("log file contains the secret %q:\n%s", s, body)
		}
	}
	if !strings.Contains(body, placeholder) {
		t.Errorf("log file has no redaction marker, so nothing was matched:\n%s", body)
	}
}

func TestWriterOffByDefaultWritesNoFile(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter()
	w.stderr = nil
	w.Logger().Info("hello")

	if w.Enabled() {
		t.Error("Enabled() = true on a fresh writer")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("log directory is not empty: %v", entries)
	}
}

func TestWriterLevelAppliesToExistingLoggers(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter()
	w.stderr = nil
	if err := w.Enable(dir); err != nil {
		t.Fatalf("Enable() = %v", err)
	}
	log := w.Logger()

	log.Debug("quiet")
	w.SetLevel(slog.LevelDebug)
	log.Debug("loud")
	w.Disable()

	data, err := os.ReadFile(filepath.Join(dir, logFileName))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	body := string(data)
	if strings.Contains(body, "quiet") {
		t.Error("debug line was written before the level allowed it")
	}
	if !strings.Contains(body, "loud") {
		t.Error("debug line missing after raising the level on the shared writer")
	}
}

func TestParseLevelRoundTrip(t *testing.T) {
	for _, name := range []string{"debug", "info", "warn", "error"} {
		if got := LevelName(ParseLevel(name)); got != name {
			t.Errorf("LevelName(ParseLevel(%q)) = %q", name, got)
		}
	}
	if got := ParseLevel("nonsense"); got != slog.LevelInfo {
		t.Errorf("ParseLevel(nonsense) = %v, want info", got)
	}
}
