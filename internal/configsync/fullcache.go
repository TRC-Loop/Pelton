package configsync

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// openFileRetrying opens src for reading, retrying a few times with a short
// backoff. See readFileRetrying in configsync.go.
func openFileRetrying(src string) (*os.File, error) {
	const attempts = 5
	var f *os.File
	var err error
	for i := range attempts {
		f, err = os.Open(src)
		if err == nil || os.IsNotExist(err) {
			return f, err
		}
		time.Sleep(time.Duration(200*(i+1)) * time.Millisecond)
	}
	return nil, fmt.Errorf("%w (if this folder is a cloud-sync placeholder, it may still be downloading)", err)
}

// pushFullSnapshot writes a consistent point-in-time copy of the local mail
// database (via VACUUM INTO, safe against a live database) plus the
// attachments directory into the sync folder. It replaces whatever snapshot
// was there before, atomically at the file level.
func (m *Manager) pushFullSnapshot(ctx context.Context, cfg Config) error {
	dest := joinPath(cfg.Path, dbSnapshotName)
	tmp := dest + ".tmp"
	os.Remove(tmp)
	if err := m.store.Snapshot(ctx, tmp); err != nil {
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		return fmt.Errorf("configsync: place mail cache snapshot: %w", err)
	}
	return copyDir(m.store.AttachmentsDir(), joinPath(cfg.Path, attachmentsDir))
}

// checkPendingFullRestore does nothing at runtime beyond noticing that a
// newer remote snapshot exists. Swapping the live, open database file out
// from under the app is not safe, so a full-cache pull only ever arms a
// marker (writePendingFullRestore) that ApplyPendingFullRestore consumes at
// the next app startup, before the database is opened.
func (m *Manager) checkPendingFullRestore(cfg Config) error {
	remote := joinPath(cfg.Path, dbSnapshotName)
	info, err := os.Stat(remote)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	marker := filepath.Join(m.stateDir, pendingRestoreFile)
	existing, statErr := os.Stat(marker)
	if statErr == nil && !info.ModTime().After(existing.ModTime()) {
		// already armed for this (or a newer) snapshot.
		return nil
	}
	return os.WriteFile(marker, []byte(remote), 0o644)
}

// ApplyPendingFullRestore checks for a marker left by a previous run's full
// cache pull and, if present, replaces the local database and attachments
// with the remote snapshot before anything opens them. Call this before
// storage.Open. dbPath and attachmentsDir are the local database file and
// attachments directory it will overwrite.
func ApplyPendingFullRestore(stateDir, dbPath, attachmentsDirPath string) error {
	marker := filepath.Join(stateDir, pendingRestoreFile)
	raw, err := os.ReadFile(marker)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	remoteDB := string(raw)
	if _, err := os.Stat(remoteDB); err != nil {
		// the snapshot disappeared (folder unmounted, user cleared it); drop
		// the marker rather than fail startup.
		os.Remove(marker)
		return nil
	}

	if err := copyFile(remoteDB, dbPath); err != nil {
		return fmt.Errorf("configsync: restore mail cache: %w", err)
	}
	remoteAttachments := filepath.Join(filepath.Dir(remoteDB), attachmentsDir)
	if _, err := os.Stat(remoteAttachments); err == nil {
		if err := copyDir(remoteAttachments, attachmentsDirPath); err != nil {
			return fmt.Errorf("configsync: restore attachments: %w", err)
		}
	}
	return os.Remove(marker)
}

func copyFile(src, dst string) error {
	in, err := openFileRetrying(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

// copyDir mirrors src into dst, copying only files newer than their
// destination counterpart (or new files entirely), so repeated calls stay
// cheap once the two sides are close to in sync.
func copyDir(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		srcInfo, err := d.Info()
		if err != nil {
			return err
		}
		if dstInfo, err := os.Stat(target); err == nil && !srcInfo.ModTime().After(dstInfo.ModTime()) {
			return nil
		}
		return copyFile(path, target)
	})
}
