package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	// logFileName is the active log inside the log directory. Rotated copies
	// get a .1, .2, ... suffix, oldest highest.
	logFileName = "pelton.log"
	// maxFileBytes is how large the active log grows before it rotates, and
	// keptFiles is how many rotated copies survive. Together they cap the log
	// directory at maxFileBytes * (keptFiles + 1), so logging left on cannot
	// quietly fill the disk.
	maxFileBytes = 2 << 20
	keptFiles    = 3
)

// rotator is an append-only file writer that rolls over at maxFileBytes.
type rotator struct {
	mu   sync.Mutex
	dir  string
	file *os.File
	size int64
}

// openRotator opens (or creates) the active log file in dir, appending to
// whatever is already there.
func openRotator(dir string) (*rotator, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	r := &rotator{dir: dir}
	if err := r.open(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *rotator) open() error {
	f, err := os.OpenFile(r.path(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	r.file = f
	r.size = info.Size()
	return nil
}

func (r *rotator) path() string {
	return filepath.Join(r.dir, logFileName)
}

// Write appends p, rotating first if it would push the active file past the
// size cap. A single write larger than the cap still lands whole rather than
// being split across files.
func (r *rotator) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file == nil {
		return 0, os.ErrClosed
	}
	if r.size > 0 && r.size+int64(len(p)) > maxFileBytes {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := r.file.Write(p)
	r.size += int64(n)
	return n, err
}

// rotate closes the active file, shifts the kept copies down one and starts a
// fresh file. The caller holds mu.
func (r *rotator) rotate() error {
	if err := r.file.Close(); err != nil {
		return err
	}
	r.file = nil
	base := r.path()
	// drop the oldest, then walk backwards so no rename overwrites a file it
	// still has to move.
	os.Remove(fmt.Sprintf("%s.%d", base, keptFiles))
	for i := keptFiles - 1; i >= 1; i-- {
		os.Rename(fmt.Sprintf("%s.%d", base, i), fmt.Sprintf("%s.%d", base, i+1))
	}
	if keptFiles > 0 {
		if err := os.Rename(base, base+".1"); err != nil {
			return err
		}
	} else {
		os.Remove(base)
	}
	return r.open()
}

// Close closes the active file. Writes after Close return os.ErrClosed.
func (r *rotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file == nil {
		return nil
	}
	err := r.file.Close()
	r.file = nil
	return err
}

// removeLogs deletes the active log and every rotated copy in dir, leaving the
// directory itself in place.
func removeLogs(dir string) error {
	base := filepath.Join(dir, logFileName)
	if err := os.Remove(base); err != nil && !os.IsNotExist(err) {
		return err
	}
	for i := 1; i <= keptFiles; i++ {
		if err := os.Remove(fmt.Sprintf("%s.%d", base, i)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
