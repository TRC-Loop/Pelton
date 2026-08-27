package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// crashPrefix and crashSuffix bracket a crash file's timestamp:
	// crash-20260816-142530.log.
	crashPrefix = "crash-"
	crashSuffix = ".log"
	// ackFileName remembers the newest crash the user has already been shown,
	// so the next launch does not offer the same one again.
	ackFileName = "crashes-acknowledged"
	// crashTimeLayout is the filename timestamp: sortable, no separators that
	// need escaping on any platform.
	crashTimeLayout = "20060102-150405"
)

// crashConfig is what a crash file records about the build. It is set once at
// startup; a panic handler cannot go looking for it.
var crashConfig struct {
	mu      sync.RWMutex
	dir     string
	version string
	channel string
	enabled bool
}

// ConfigureCrashes tells the panic handlers where to write and what build they
// are running in. Until it is called with enabled true, a panic is left to the
// runtime exactly as before.
func ConfigureCrashes(dir, version, channel string, enabled bool) {
	crashConfig.mu.Lock()
	defer crashConfig.mu.Unlock()
	crashConfig.dir = dir
	crashConfig.version = version
	crashConfig.channel = channel
	crashConfig.enabled = enabled
}

// Guard recovers a panic, records it and exits.
//
// activity is what the app was doing, so a stack full of runtime frames comes
// with a plain sentence at the top of the file. Use it as the first statement
// of any goroutine the app spawns:
//
//	go func() { defer logging.Guard("sync inbox"); ... }()
//
// The process still ends. A panicking goroutine takes the program down anyway;
// the difference is that this way it leaves a file behind instead of vanishing.
func Guard(activity string) {
	r := recover()
	if r == nil {
		return
	}
	stack := debug.Stack()
	path, err := WriteCrash(activity, r, stack)
	if err == nil && path != "" {
		fmt.Fprintf(os.Stderr, "pelton: crash report written to %s\n", path)
	}
	fmt.Fprintf(os.Stderr, "panic: %v\n\n%s", r, stack)
	os.Exit(1)
}

// WriteCrash records one panic and returns the file it wrote. It returns an
// empty path with no error when crash logging is off, so callers can treat that
// as "nothing to show" rather than a failure.
func WriteCrash(activity string, value any, stack []byte) (string, error) {
	crashConfig.mu.RLock()
	dir, version, channel, enabled := crashConfig.dir, crashConfig.version, crashConfig.channel, crashConfig.enabled
	crashConfig.mu.RUnlock()
	if !enabled || dir == "" {
		return "", nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}

	now := time.Now()
	if channel == "" {
		channel = "stable"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "pelton crash report\n")
	fmt.Fprintf(&b, "time:     %s\n", now.Format(time.RFC3339))
	fmt.Fprintf(&b, "version:  %s (%s)\n", version, channel)
	fmt.Fprintf(&b, "os:       %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&b, "go:       %s\n", runtime.Version())
	fmt.Fprintf(&b, "activity: %s\n", activity)
	fmt.Fprintf(&b, "panic:    %v\n\n", value)
	b.Write(stack)

	path := filepath.Join(dir, crashPrefix+now.Format(crashTimeLayout)+crashSuffix)
	if err := os.WriteFile(path, []byte(secrets.Redact(b.String())), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// Crash is a crash file on disk.
type Crash struct {
	// Path is the file. Name is its base name, which is also what Acknowledge
	// records.
	Path string
	Name string
	When time.Time
}

// PendingCrash returns the newest crash file in dir that has not been
// acknowledged yet, and whether there was one. This is what makes the next
// launch able to say something happened.
func PendingCrash(dir string) (Crash, bool) {
	names := crashNames(dir)
	if len(names) == 0 {
		return Crash{}, false
	}
	newest := names[len(names)-1]
	if acked, err := os.ReadFile(filepath.Join(dir, ackFileName)); err == nil {
		if strings.TrimSpace(string(acked)) >= newest {
			return Crash{}, false
		}
	}
	when, err := time.ParseInLocation(crashTimeLayout, strings.TrimSuffix(strings.TrimPrefix(newest, crashPrefix), crashSuffix), time.Local)
	if err != nil {
		when = time.Time{}
	}
	return Crash{Path: filepath.Join(dir, newest), Name: newest, When: when}, true
}

// AcknowledgeCrashes marks every crash file currently in dir as seen, so
// PendingCrash stops reporting them. The files stay: the user may want to open
// one later, and deleting on dismiss would throw away the report the prompt was
// about.
func AcknowledgeCrashes(dir string) error {
	names := crashNames(dir)
	if len(names) == 0 {
		return nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ackFileName), []byte(names[len(names)-1]), 0o600)
}

// RemoveCrashes deletes every crash file in dir along with the acknowledgement
// marker.
func RemoveCrashes(dir string) error {
	for _, name := range crashNames(dir) {
		if err := os.Remove(filepath.Join(dir, name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.Remove(filepath.Join(dir, ackFileName)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// crashNames lists the crash files in dir, oldest first. The timestamp in the
// name sorts chronologically as text, so no stat calls are needed.
func crashNames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, crashPrefix) || !strings.HasSuffix(name, crashSuffix) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
