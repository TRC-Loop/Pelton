package desktop

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/TRC-Loop/Pelton/internal/logging"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

func TestDebugForced(t *testing.T) {
	tests := []struct {
		name string
		args []string
		env  string
		want bool
	}{
		{name: "nothing set", args: []string{"pelton"}},
		{name: "flag", args: []string{"--debug"}, want: true},
		{name: "env var", env: "1", want: true},
		{name: "another flag is not the debug flag", args: []string{"--potatoes-are-nice"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != "" {
				t.Setenv(debugEnvVar, tt.env)
			} else {
				t.Setenv(debugEnvVar, "")
			}
			if got := debugForced(tt.args); got != tt.want {
				t.Errorf("debugForced(%v) = %t, want %t", tt.args, got, tt.want)
			}
		})
	}
}

// TestLoggingDefaultsPerChannel pins the defaults the privacy constraint rests
// on: a stable build writes nothing until asked, and a nightly, which already
// warns that it is untested, writes both.
func TestLoggingDefaultsPerChannel(t *testing.T) {
	tests := []struct {
		name       string
		channel    string
		debug      bool
		wantLogs   bool
		wantCrash  bool
		wantLevel  slog.Level
		wantForced bool
	}{
		{name: "stable", wantLevel: slog.LevelInfo},
		{name: "nightly", channel: storage.ChannelNightly, wantLogs: true, wantCrash: true, wantLevel: slog.LevelInfo},
		{name: "stable with --debug", debug: true, wantLogs: true, wantCrash: true, wantLevel: slog.LevelDebug},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no store, so every setting falls back to its default, which is
			// what this test is about.
			a := &App{channel: tt.channel, debug: tt.debug, logWriter: logging.NewWriter()}
			if got := a.logsOn(); got != tt.wantLogs {
				t.Errorf("logsOn() = %t, want %t", got, tt.wantLogs)
			}
			if got := a.crashLogsOn(); got != tt.wantCrash {
				t.Errorf("crashLogsOn() = %t, want %t", got, tt.wantCrash)
			}
			if got := a.logLevel(); got != tt.wantLevel {
				t.Errorf("logLevel() = %v, want %v", got, tt.wantLevel)
			}
		})
	}
}

func TestApplyLogSettingsWritesOnlyWhenOn(t *testing.T) {
	dir := t.TempDir()
	a := &App{dataDir: dir, version: "test", logWriter: logging.NewWriter()}
	a.log = a.logWriter.Logger()
	t.Cleanup(func() {
		a.logWriter.Disable()
		logging.ConfigureCrashes("", "", "", false)
	})

	a.applyLogSettings()
	if a.logWriter.Enabled() {
		t.Error("file logging is on with the setting off")
	}
	if _, err := os.Stat(logging.Dir(dir)); !os.IsNotExist(err) {
		t.Error("log directory was created with logging off")
	}

	a.debug = true
	a.applyLogSettings()
	if !a.logWriter.Enabled() {
		t.Fatal("--debug did not turn file logging on")
	}
	a.log.Info("hello from the test")
	if _, err := os.Stat(filepath.Join(logging.Dir(dir), "pelton.log")); err != nil {
		t.Errorf("no log file after logging a line: %v", err)
	}

	a.debug = false
	a.applyLogSettings()
	if a.logWriter.Enabled() {
		t.Error("file logging stayed on after the override was dropped")
	}
}

func TestGetLogStatusReportsThePendingCrash(t *testing.T) {
	dir := t.TempDir()
	a := &App{dataDir: dir, version: "test", channel: storage.ChannelNightly, logWriter: logging.NewWriter()}
	a.log = a.logWriter.Logger()
	t.Cleanup(func() {
		a.logWriter.Disable()
		logging.ConfigureCrashes("", "", "", false)
	})
	a.applyLogSettings()

	if got := a.GetLogStatus(); got.CrashName != "" {
		t.Errorf("GetLogStatus() reports crash %q before anything crashed", got.CrashName)
	}

	if _, err := logging.WriteCrash("syncing inbox", "boom", []byte("stack")); err != nil {
		t.Fatalf("WriteCrash() = %v", err)
	}
	status := a.GetLogStatus()
	if status.CrashName == "" {
		t.Fatal("GetLogStatus() found no crash after one was written")
	}
	if status.Dir != logging.Dir(dir) {
		t.Errorf("GetLogStatus().Dir = %q, want %q", status.Dir, logging.Dir(dir))
	}
	if status.SizeBytes == 0 {
		t.Error("GetLogStatus().SizeBytes = 0 with a crash file on disk")
	}

	if err := a.DismissCrashReport(); err != nil {
		t.Fatalf("DismissCrashReport() = %v", err)
	}
	if got := a.GetLogStatus(); got.CrashName != "" {
		t.Errorf("GetLogStatus() still reports %q after dismissal", got.CrashName)
	}

	if err := a.DeleteLogs(); err != nil {
		t.Fatalf("DeleteLogs() = %v", err)
	}
	if got := a.GetLogStatus(); got.SizeBytes != 0 {
		t.Errorf("GetLogStatus().SizeBytes = %d after DeleteLogs", got.SizeBytes)
	}
	if !a.logWriter.Enabled() {
		t.Error("DeleteLogs turned logging off instead of starting a fresh file")
	}
}
