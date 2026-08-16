package logging

import (
	"os"
	"strings"
	"testing"
)

func TestWriteCrashIsOffUntilConfigured(t *testing.T) {
	dir := t.TempDir()
	ConfigureCrashes(dir, "1.2.3", "", false)
	t.Cleanup(func() { ConfigureCrashes("", "", "", false) })

	path, err := WriteCrash("syncing inbox", "boom", []byte("stack"))
	if err != nil {
		t.Fatalf("WriteCrash() = %v", err)
	}
	if path != "" {
		t.Errorf("wrote %s with crash logging off", path)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("crash directory is not empty: %v", entries)
	}
}

func TestWriteCrashRecordsTheContextAndRedacts(t *testing.T) {
	secrets.Reset()
	t.Cleanup(secrets.Reset)
	Register("hunter2")

	dir := t.TempDir()
	ConfigureCrashes(dir, "2026.4", "nightly", true)
	t.Cleanup(func() { ConfigureCrashes("", "", "", false) })

	path, err := WriteCrash("syncing inbox", "runtime error: index out of range", []byte("goroutine 1 [running]:\npassword=hunter2\n"))
	if err != nil {
		t.Fatalf("WriteCrash() = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read crash: %v", err)
	}
	body := string(data)
	for _, want := range []string{"2026.4", "nightly", "syncing inbox", "index out of range", "goroutine 1"} {
		if !strings.Contains(body, want) {
			t.Errorf("crash file missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "hunter2") {
		t.Errorf("crash file contains a registered secret:\n%s", body)
	}
}

func TestPendingCrashStopsAfterAcknowledge(t *testing.T) {
	dir := t.TempDir()
	ConfigureCrashes(dir, "2026.4", "", true)
	t.Cleanup(func() { ConfigureCrashes("", "", "", false) })

	if _, ok := PendingCrash(dir); ok {
		t.Fatal("a fresh directory reports a pending crash")
	}

	// written by hand so two crashes get distinct timestamps without the test
	// having to wait a second between them.
	older := writeCrashFile(t, dir, "crash-20260816-101500.log")
	newer := writeCrashFile(t, dir, "crash-20260816-142530.log")

	got, ok := PendingCrash(dir)
	if !ok {
		t.Fatal("PendingCrash() found nothing")
	}
	if got.Name != "crash-20260816-142530.log" {
		t.Errorf("PendingCrash() = %s, want the newest file", got.Name)
	}
	if got.When.IsZero() {
		t.Error("PendingCrash() did not parse the timestamp out of the name")
	}

	if err := AcknowledgeCrashes(dir); err != nil {
		t.Fatalf("AcknowledgeCrashes() = %v", err)
	}
	if _, ok := PendingCrash(dir); ok {
		t.Error("PendingCrash() still reports an acknowledged crash")
	}
	for _, path := range []string{older, newer} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("acknowledging deleted %s: %v", path, err)
		}
	}

	// a crash after the acknowledgement is pending again.
	writeCrashFile(t, dir, "crash-20260817-090000.log")
	if _, ok := PendingCrash(dir); !ok {
		t.Error("a newer crash is not reported after an acknowledgement")
	}

	if err := RemoveCrashes(dir); err != nil {
		t.Fatalf("RemoveCrashes() = %v", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("RemoveCrashes() left %v", entries)
	}
}

func writeCrashFile(t *testing.T, dir, name string) string {
	t.Helper()
	path := dir + string(os.PathSeparator) + name
	if err := os.WriteFile(path, []byte("pelton crash report\n"), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}
