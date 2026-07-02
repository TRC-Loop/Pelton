package configsync

// In-place mode: the chosen folder IS the app's data directory. No scope, no
// json snapshots, no periodic sync pass - the folder is just synced by
// whatever cloud tool the user already has. Unsafe with two devices open at
// once (concurrent writers to one live sqlite file); the ui warns about this.

import (
	"context"
	"os"
	"path/filepath"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

const inPlaceMarkerFile = "configsync_inplace_dir"

// must match settingSearchWatermark in internal/desktop/bind_search.go.
const searchWatermarkKey = "search_indexed_max_id"

type inPlaceMarker struct {
	Path string `json:"path"`
}

// ActiveDataDir resolves which directory this launch should use: the
// in-place folder if configured and reachable, otherwise defaultDir. The
// marker lives in stateDir (the normal app-support dir) so it is found
// before the database itself can be opened.
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

// EnableInPlace points future launches at target. If target already has a
// database (another device set it up first) it is adopted as-is; otherwise
// the current live data is snapshotted into it. Takes effect on next start.
func EnableInPlace(ctx context.Context, store *storage.DB, stateDir, target, dbFileName string) error {
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}
	targetDB := filepath.Join(target, dbFileName)
	if _, err := os.Stat(targetDB); err != nil {
		if err := store.Snapshot(ctx, targetDB); err != nil {
			return err
		}
		if err := copyDir(store.AttachmentsDir(), filepath.Join(target, attachmentsDir)); err != nil {
			return err
		}
	}
	if err := store.SetInt(ctx, searchWatermarkKey, 0); err != nil {
		return err
	}
	return writeJSONFile(filepath.Join(stateDir, inPlaceMarkerFile), inPlaceMarker{Path: target})
}

// DisableInPlace points future launches back at stateDir, snapshotting the
// current live data there first. Takes effect on next start.
func DisableInPlace(ctx context.Context, store *storage.DB, stateDir, dbFileName string) error {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}
	if err := store.Snapshot(ctx, filepath.Join(stateDir, dbFileName)); err != nil {
		return err
	}
	if err := copyDir(store.AttachmentsDir(), filepath.Join(stateDir, attachmentsDir)); err != nil {
		return err
	}
	if err := store.SetInt(ctx, searchWatermarkKey, 0); err != nil {
		return err
	}
	marker := filepath.Join(stateDir, inPlaceMarkerFile)
	if err := os.Remove(marker); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// PeekInPlaceFolder reports whether target already holds another device's
// in-place data, without changing anything.
func PeekInPlaceFolder(target, dbFileName string) (bool, error) {
	if _, err := os.Stat(filepath.Join(target, dbFileName)); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// FolderSummary describes an existing in-place folder's contents, so the
// setup ui can show what joining it would bring in before the user commits.
type FolderSummary struct {
	AccountEmails []string
	ModifiedUnix  int64
}

// SummarizeInPlaceFolder opens target's database read-only-in-spirit (a
// throwaway connection, closed before returning) and reports what's in it.
// Callers should check PeekInPlaceFolder first; this errors if there is
// nothing there.
func SummarizeInPlaceFolder(ctx context.Context, target, dbFileName string) (FolderSummary, error) {
	dbPath := filepath.Join(target, dbFileName)
	info, err := os.Stat(dbPath)
	if err != nil {
		return FolderSummary{}, err
	}
	db, err := storage.Open(dbPath)
	if err != nil {
		return FolderSummary{}, err
	}
	defer db.Close()

	accounts, err := db.ListAccounts(ctx)
	if err != nil {
		return FolderSummary{}, err
	}
	emails := make([]string, len(accounts))
	for i, a := range accounts {
		emails[i] = a.Email
	}
	return FolderSummary{AccountEmails: emails, ModifiedUnix: info.ModTime().Unix()}, nil
}
