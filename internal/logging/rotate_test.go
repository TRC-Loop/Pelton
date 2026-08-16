package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRotatorCapsTheDirectory(t *testing.T) {
	dir := t.TempDir()
	rot, err := openRotator(dir)
	if err != nil {
		t.Fatalf("openRotator() = %v", err)
	}
	t.Cleanup(func() { rot.Close() })

	line := []byte(strings.Repeat("x", 64*1024) + "\n")
	// enough to fill the active file and every kept copy several times over.
	for range (maxFileBytes / len(line)) * (keptFiles + 3) {
		if _, err := rot.Write(line); err != nil {
			t.Fatalf("Write() = %v", err)
		}
	}

	var total int64
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			t.Fatalf("stat %s: %v", e.Name(), err)
		}
		total += info.Size()
	}
	if len(entries) != keptFiles+1 {
		t.Errorf("got %d files, want %d", len(entries), keptFiles+1)
	}
	if cap := int64(maxFileBytes) * int64(keptFiles+1); total > cap {
		t.Errorf("log directory is %d bytes, over the %d cap", total, cap)
	}
}

func TestRotatorKeepsTheNewestLinesInTheActiveFile(t *testing.T) {
	dir := t.TempDir()
	rot, err := openRotator(dir)
	if err != nil {
		t.Fatalf("openRotator() = %v", err)
	}
	t.Cleanup(func() { rot.Close() })

	filler := []byte(strings.Repeat("y", 128*1024) + "\n")
	for range (maxFileBytes / len(filler)) + 1 {
		rot.Write(filler)
	}
	if _, err := rot.Write([]byte("the newest line\n")); err != nil {
		t.Fatalf("Write() = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, logFileName))
	if err != nil {
		t.Fatalf("read active log: %v", err)
	}
	if !strings.Contains(string(data), "the newest line") {
		t.Error("newest line is not in the active file after a rotation")
	}
	if _, err := os.Stat(filepath.Join(dir, logFileName+".1")); err != nil {
		t.Errorf("no rotated copy: %v", err)
	}
}

func TestRemoveLogsClearsEveryCopy(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, logFileName)
	names := []string{base}
	for i := 1; i <= keptFiles; i++ {
		names = append(names, fmt.Sprintf("%s.%d", base, i))
	}
	for _, name := range names {
		if err := os.WriteFile(name, []byte("old"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	if err := removeLogs(dir); err != nil {
		t.Fatalf("removeLogs() = %v", err)
	}
	for _, name := range names {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			t.Errorf("%s survived removeLogs", filepath.Base(name))
		}
	}
	// a second pass must not turn the missing files into an error.
	if err := removeLogs(dir); err != nil {
		t.Errorf("removeLogs() on an empty dir = %v", err)
	}
}
