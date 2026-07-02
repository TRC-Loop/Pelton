package configsync

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

// MergeIntoInPlaceFolder merges this device's accounts and settings into an
// already-populated in-place folder's database, then switches this device to
// use that folder next start. Its own mail cache is left on disk unused, not
// merged - see mergeAccountsAndSettings.
func MergeIntoInPlaceFolder(ctx context.Context, local *storage.DB, stateDir, target, dbFileName string) error {
	targetDB := filepath.Join(target, dbFileName)
	exists, err := PeekInPlaceFolder(target, dbFileName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("configsync: %q has no existing pelton data to merge into", target)
	}

	remote, err := storage.Open(targetDB)
	if err != nil {
		return err
	}
	defer remote.Close()
	if err := remote.RunMigrations(ctx); err != nil {
		return err
	}
	if err := mergeAccountsAndSettings(ctx, local, remote); err != nil {
		return err
	}
	return writeJSONFile(filepath.Join(stateDir, inPlaceMarkerFile), inPlaceMarker{Path: target})
}

// mergeAccountsAndSettings merges local into target: settings last-write-wins
// per key (this device's own configsync setting is never carried over, same
// as pullSettings/pushSettings), and accounts are unioned by email. A merged
// account has no keyring entry on the target's original device - credentials
// are OS-keyring only by design and never travel between devices, so a
// merged-in account needs its password re-entered here, same as any newly
// added account. Cached mail is not merged; the target folder's cache is
// what this device uses going forward.
func mergeAccountsAndSettings(ctx context.Context, local, target *storage.DB) error {
	settings, err := local.AllSettings(ctx)
	if err != nil {
		return err
	}
	for _, s := range settings {
		if s.Key == settingKey {
			continue
		}
		if err := target.SetIfNewer(ctx, s.Key, s.Value, s.UpdatedAt); err != nil {
			return err
		}
	}

	localAccounts, err := local.ListAccounts(ctx)
	if err != nil {
		return err
	}
	targetAccounts, err := target.ListAccounts(ctx)
	if err != nil {
		return err
	}
	have := make(map[string]bool, len(targetAccounts))
	for _, a := range targetAccounts {
		have[strings.ToLower(a.Email)] = true
	}
	for _, a := range localAccounts {
		if have[strings.ToLower(a.Email)] {
			continue
		}
		a.ID = 0
		if _, err := target.CreateAccount(ctx, &a); err != nil {
			return err
		}
	}
	return nil
}
