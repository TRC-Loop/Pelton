// Package configsync used to mirror settings and mail through a user-chosen
// folder. That feature was removed; what remains is the data-directory
// plumbing needed to migrate any device that had the old "in-place" mode
// (where the live data directory was redirected to the synced folder) back to
// the normal per-OS app-support directory on the next launch.
package configsync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

const (
	inPlaceMarkerFile = "configsync_inplace_dir"
	attachmentsDir    = "attachments"
	// searchWatermarkKey must match settingSearchWatermark in bind_search.go; the
	// migration resets it so the index rebuilds against the migrated database.
	searchWatermarkKey = "search_indexed_max_id"
)

// inPlaceMarker is the legacy marker file that redirected the data directory.
type inPlaceMarker struct {
	Path string `json:"path"`
}

// ActiveDataDir resolves which directory this launch should open the database
// from: the legacy in-place folder if a reachable marker still points at one,
// otherwise defaultDir. The marker lives in stateDir (the normal app-support
// dir) so it is found before the database itself can be opened.
func ActiveDataDir(stateDir, defaultDir string) (string, error) {
	var marker inPlaceMarker
	found, err := readJSONFile(filepath.Join(stateDir, inPlaceMarkerFile), &marker)
	if err != nil || !found || marker.Path == "" {
		return defaultDir, err
	}
	if _, statErr := os.Stat(marker.Path); statErr != nil {
		return defaultDir, nil
	}
	return marker.Path, nil
}

// MigrateInPlaceBack undoes the legacy in-place redirection: it snapshots the
// currently-open (in-place) database and its attachments back into stateDir and
// removes the marker, so the next launch opens from the normal directory again.
// It is a no-op (returns false) when no in-place marker is present. Reported so
// startup can log that a migration happened.
func MigrateInPlaceBack(ctx context.Context, store *storage.DB, stateDir, dbFileName string) (bool, error) {
	markerPath := filepath.Join(stateDir, inPlaceMarkerFile)
	var marker inPlaceMarker
	found, err := readJSONFile(markerPath, &marker)
	if err != nil || !found {
		return false, err
	}
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return false, err
	}
	if err := store.Snapshot(ctx, filepath.Join(stateDir, dbFileName)); err != nil {
		return false, err
	}
	if err := copyDir(store.AttachmentsDir(), filepath.Join(stateDir, attachmentsDir)); err != nil {
		return false, err
	}
	if err := store.SetInt(ctx, searchWatermarkKey, 0); err != nil {
		return false, err
	}
	if err := os.Remove(markerPath); err != nil && !os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

// readJSONFile decodes the JSON file at path into v, returning found=false when
// it does not exist.
func readJSONFile(path string, v any) (bool, error) {
	data, err := readFileRetrying(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return false, err
	}
	return true, nil
}

// readFileRetrying reads path, retrying a few times with a short backoff before
// giving up (a cloud-sync placeholder file can transiently fail while it is
// still hydrating).
func readFileRetrying(path string) ([]byte, error) {
	const attempts = 5
	var data []byte
	var err error
	for i := range attempts {
		data, err = os.ReadFile(path)
		if err == nil || os.IsNotExist(err) {
			return data, err
		}
		if i < attempts-1 {
			time.Sleep(time.Duration(200*(i+1)) * time.Millisecond)
		}
	}
	return nil, fmt.Errorf("%w (if this folder is a cloud-sync placeholder, it may still be downloading)", err)
}

// copyDir mirrors src into dst, copying only files newer than their destination
// counterpart (or new files entirely).
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

// copyFile copies a single file atomically via a temp file and rename.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
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
