package desktop

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TRC-Loop/Pelton/internal/logging"
)

func newDevToolsApp(t *testing.T) *App {
	t.Helper()
	w := logging.NewWriter()
	return &App{logWriter: w, log: w.Logger(), startedAt: time.Now()}
}

// The overlays read log lines, which carry mailbox names and addresses. Off is
// the default and no setting can change it, so the only thing to pin is that a
// plain run really is off and every binding refuses.
func TestDevToolsOffWithoutTheEnvironment(t *testing.T) {
	t.Setenv("PELTON_DEV", "")
	t.Setenv(devToolsEnvVar, "")
	a := newDevToolsApp(t)

	if a.DevToolsEnabled() {
		t.Fatal("developer tools are enabled on a plain run")
	}
	if _, err := a.DevActivity(0); !errors.Is(err, errDevToolsDisabled) {
		t.Errorf("DevActivity err = %v, want errDevToolsDisabled", err)
	}
	if _, err := a.DevProcessStats(); !errors.Is(err, errDevToolsDisabled) {
		t.Errorf("DevProcessStats err = %v, want errDevToolsDisabled", err)
	}
	if err := a.ClearDevActivity(); !errors.Is(err, errDevToolsDisabled) {
		t.Errorf("ClearDevActivity err = %v, want errDevToolsDisabled", err)
	}
}

func TestDevToolsEnabledByEitherVariable(t *testing.T) {
	for _, name := range []string{"PELTON_DEV", devToolsEnvVar} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("PELTON_DEV", "")
			t.Setenv(devToolsEnvVar, "")
			t.Setenv(name, "1")

			if !newDevToolsApp(t).DevToolsEnabled() {
				t.Errorf("%s did not enable the developer tools", name)
			}
		})
	}
}

// Nothing is buffered until the overlay asks, and once it does the app's own
// log lines are what it sees.
func TestDevActivityReadsTheAppLog(t *testing.T) {
	t.Setenv(devToolsEnvVar, "1")
	a := newDevToolsApp(t)

	a.log.Info("syncing", "folder", "INBOX")

	first, err := a.DevActivity(0)
	if err != nil {
		t.Fatalf("DevActivity: %v", err)
	}
	// the line above was written before anything asked for a buffer, so it is
	// gone; what matters is that lines written from here on arrive.
	a.log.Info("fetched", "count", 3)

	second, err := a.DevActivity(first.Next)
	if err != nil {
		t.Fatalf("DevActivity: %v", err)
	}
	if len(second.Lines) != 1 {
		t.Fatalf("got %d new lines, want 1: %v", len(second.Lines), second.Lines)
	}
	if !strings.Contains(second.Lines[0].Text, "fetched") {
		t.Errorf("line = %q, want the fetched line", second.Lines[0].Text)
	}
	if second.Level == "" {
		t.Error("no log level reported")
	}
}

func TestClearDevActivityEmptiesTheBuffer(t *testing.T) {
	t.Setenv(devToolsEnvVar, "1")
	a := newDevToolsApp(t)

	if _, err := a.DevActivity(0); err != nil {
		t.Fatalf("DevActivity: %v", err)
	}
	a.log.Info("before the clear")
	if err := a.ClearDevActivity(); err != nil {
		t.Fatalf("ClearDevActivity: %v", err)
	}

	got, err := a.DevActivity(0)
	if err != nil {
		t.Fatalf("DevActivity: %v", err)
	}
	if len(got.Lines) != 0 {
		t.Errorf("lines = %v, want none after the clear", got.Lines)
	}
}

func TestDevProcessStatsReportsTheRuntime(t *testing.T) {
	t.Setenv(devToolsEnvVar, "1")
	a := newDevToolsApp(t)

	stats, err := a.DevProcessStats()
	if err != nil {
		t.Fatalf("DevProcessStats: %v", err)
	}
	if stats.Goroutines < 1 {
		t.Errorf("Goroutines = %d, want at least the one running this test", stats.Goroutines)
	}
	if stats.HeapBytes == 0 {
		t.Error("HeapBytes = 0, want the live heap")
	}
}

// The database is more than the .db file in WAL mode, and reporting only that
// would understate it.
func TestDatabaseSizeCountsTheWalFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pelton.db")
	for suffix, size := range map[string]int{"": 100, "-wal": 40, "-shm": 10} {
		if err := os.WriteFile(path+suffix, make([]byte, size), 0o600); err != nil {
			t.Fatalf("write %s: %v", suffix, err)
		}
	}

	if got := databaseSize(path); got != 150 {
		t.Errorf("databaseSize = %d, want 150", got)
	}
	if got := databaseSize(filepath.Join(dir, "missing.db")); got != 0 {
		t.Errorf("databaseSize of a missing file = %d, want 0", got)
	}
}

// treeSize walks the whole tree, unlike dirSize, because attachments are nested
// per account and per message.
func TestTreeSizeWalksSubdirectories(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "1", "2")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "a.bin"), make([]byte, 64), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	if got := treeSize(dir); got != 64 {
		t.Errorf("treeSize = %d, want 64", got)
	}
	if got := treeSize(filepath.Join(dir, "nope")); got != 0 {
		t.Errorf("treeSize of a missing dir = %d, want 0", got)
	}
}
