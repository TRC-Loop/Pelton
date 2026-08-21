// bind_devtools.go backs the developer overlays (#188): a live view of what the
// app is doing, without adding log lines and rebuilding.
//
// Everything here stays on the machine. There is no reporter and no endpoint;
// the overlays read what the process already knows about itself. The activity
// overlay shows log lines, which carry mailbox names and addresses, so the
// whole surface is off unless the app was deliberately started for development:
// PELTON_DEV (what `make run` sets) or PELTON_DEVTOOLS. No setting can turn it
// on, so it cannot be reached by a stray click or by anything remote.
package desktop

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// devToolsEnvVar turns the overlays on for a build that is not running under
// PELTON_DEV, which is the only way to get them in a packaged binary.
const devToolsEnvVar = "PELTON_DEVTOOLS"

// devLogBufferLines is how many recent log lines the activity overlay keeps.
// Enough to cover a sync pass and the mistake before it, small enough that a
// session left running overnight costs nothing.
const devLogBufferLines = 500

// devDirSizeTTL is how long a measured directory size is reused. Walking the
// attachment tree is the one expensive thing on this screen and its size does
// not move between two ticks of a 2 second poll.
const devDirSizeTTL = 10 * time.Second

// errDevToolsDisabled is returned by every binding here when the app was not
// started for development. It is an error rather than empty data so a caller
// cannot mistake "off" for "nothing is happening".
var errDevToolsDisabled = errors.New("developer tools are not enabled")

// DevToolsEnabled reports whether the developer overlays are available. The
// frontend asks once at startup and does not bind the keys otherwise.
func (a *App) DevToolsEnabled() bool {
	return a.IsDevMode() || os.Getenv(devToolsEnvVar) != ""
}

// DevLogLineDTO is one buffered log line. Seq is the line's position in the
// whole session, so a reader can both ask for what is new and see that lines
// were dropped.
type DevLogLineDTO struct {
	Seq  uint64 `json:"seq"`
	Text string `json:"text"`
}

// DevActivityDTO is a page of the activity log. Next is the sequence to ask for
// on the following call.
type DevActivityDTO struct {
	Lines []DevLogLineDTO `json:"lines"`
	Next  uint64          `json:"next"`
	// Level is the current log threshold. Below debug, most of what the sync
	// path does is never written, so the overlay says so rather than looking
	// like nothing is happening.
	Level string `json:"level"`
}

// DevActivity returns the log lines buffered since the given sequence, starting
// the buffer on the first call. Pass 0 to get everything still held.
//
// The lines are the same redacted text that reaches stderr and the log file, so
// a registered secret is gone from all three.
func (a *App) DevActivity(after uint64) (DevActivityDTO, error) {
	if !a.DevToolsEnabled() {
		return DevActivityDTO{}, errDevToolsDisabled
	}
	if a.logWriter == nil {
		return DevActivityDTO{}, errors.New("desktop: no log writer")
	}

	lines, next := a.logWriter.Buffer(devLogBufferLines).Since(after)
	out := make([]DevLogLineDTO, 0, len(lines))
	for _, line := range lines {
		out = append(out, DevLogLineDTO{Seq: line.Seq, Text: line.Text})
	}
	return DevActivityDTO{
		Lines: out,
		Next:  next,
		Level: a.logWriter.Level().String(),
	}, nil
}

// ClearDevActivity empties the activity buffer, for starting a fresh read of
// one operation.
func (a *App) ClearDevActivity() error {
	if !a.DevToolsEnabled() {
		return errDevToolsDisabled
	}
	if a.logWriter == nil {
		return nil
	}
	a.logWriter.Buffer(devLogBufferLines).Clear()
	return nil
}

// DevProcessDTO is what the process overlay shows: the Go runtime's own numbers
// and what the app has put on disk.
type DevProcessDTO struct {
	Goroutines int `json:"goroutines"`
	// HeapBytes is live heap (HeapAlloc), HeapSysBytes what the runtime has
	// taken from the os. The gap between them is the runtime holding on to
	// memory it may reuse, which is normal and not a leak.
	HeapBytes    uint64 `json:"heapBytes"`
	HeapSysBytes uint64 `json:"heapSysBytes"`
	GCRuns       uint32 `json:"gcRuns"`
	// DatabaseBytes is the .db file only. WAL and shm files sit next to it and
	// are counted with it, since they are the same database.
	DatabaseBytes    int64 `json:"databaseBytes"`
	AttachmentsBytes int64 `json:"attachmentsBytes"`
	DataDirBytes     int64 `json:"dataDirBytes"`
	// UptimeSeconds is how long this process has been running.
	UptimeSeconds int64 `json:"uptimeSeconds"`
}

// DevProcessStats samples the runtime and measures the app's directories.
// Directory sizes are cached briefly, so polling this does not turn into a
// continuous walk of the attachment tree.
func (a *App) DevProcessStats() (DevProcessDTO, error) {
	if !a.DevToolsEnabled() {
		return DevProcessDTO{}, errDevToolsDisabled
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	out := DevProcessDTO{
		Goroutines:    runtime.NumGoroutine(),
		HeapBytes:     mem.HeapAlloc,
		HeapSysBytes:  mem.HeapSys,
		GCRuns:        mem.NumGC,
		UptimeSeconds: int64(time.Since(a.startedAt).Seconds()),
	}
	if a.store != nil {
		out.DatabaseBytes = databaseSize(a.store.Path())
		out.AttachmentsBytes = a.cachedDirSize(a.store.AttachmentsDir())
	}
	if a.dataDir != "" {
		out.DataDirBytes = a.cachedDirSize(a.dataDir)
	}
	return out, nil
}

// measuredDir is one measured directory and when it was measured.
type measuredDir struct {
	bytes int64
	at    time.Time
}

// cachedDirSize returns the size of dir, remeasuring only once the last figure
// has gone stale. A directory that cannot be read reports 0 rather than failing
// the whole overlay.
func (a *App) cachedDirSize(dir string) int64 {
	a.dirSizesMu.Lock()
	defer a.dirSizesMu.Unlock()

	if cached, ok := a.dirSizes[dir]; ok && time.Since(cached.at) < devDirSizeTTL {
		return cached.bytes
	}
	size := treeSize(dir)
	if a.dirSizes == nil {
		a.dirSizes = make(map[string]measuredDir)
	}
	a.dirSizes[dir] = measuredDir{bytes: size, at: time.Now()}
	return size
}

// treeSize adds up the regular files under dir and everything below it, unlike
// dirSize which only reads one level. Unreadable entries are skipped: an
// incomplete figure is more useful here than none.
func treeSize(dir string) int64 {
	var total int64
	_ = filepath.WalkDir(dir, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil //nolint:nilerr // a directory we cannot read is skipped, not fatal
		}
		if info, err := entry.Info(); err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

// databaseSize adds the WAL and shared-memory files to the main database file.
// In WAL mode a busy database can hold a lot of its recent size in the -wal
// file, and reporting only the .db would understate it.
func databaseSize(path string) int64 {
	if path == "" {
		return 0
	}
	var total int64
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if info, err := os.Stat(path + suffix); err == nil {
			total += info.Size()
		}
	}
	return total
}
