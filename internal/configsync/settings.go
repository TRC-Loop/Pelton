package configsync

import (
	"context"
)

// settingsSnapshot is the file format for settings.json: a flat key/value map
// plus a timestamp per key so an importer can apply last-write-wins.
type settingsSnapshot struct {
	Settings map[string]settingEntry `json:"settings"`
}

type settingEntry struct {
	Value     string `json:"value"`
	UpdatedAt string `json:"updatedAt"`
}

func (m *Manager) settingsPath(cfg Config) string {
	return joinPath(cfg.Path, settingsFileName)
}

// pullSettings applies any remote setting whose updatedAt is not older than
// the local row's, using storage's SetIfNewer so a stale remote file can
// never roll back a more recent local change.
//
// settingKey (this device's own sync mode/path/scope) is never applied from a
// remote snapshot: it is per-device configuration, not a shared preference.
// Applying it would let one device's sync setup silently overwrite another's -
// including flipping a "read-only" device's mode away from read-only, or
// pointing it at a different folder, the moment it pulled.
func (m *Manager) pullSettings(ctx context.Context, cfg Config) error {
	var snap settingsSnapshot
	found, err := readJSONFile(m.settingsPath(cfg), &snap)
	if err != nil || !found {
		return err
	}
	for key, entry := range snap.Settings {
		if key == settingKey {
			continue
		}
		if err := m.store.SetIfNewer(ctx, key, entry.Value, entry.UpdatedAt); err != nil {
			return err
		}
	}
	return nil
}

// pushSettings writes every local setting out to the folder, except
// settingKey - see pullSettings for why this device's own sync configuration
// is never shared. It always exports the rest in full (dominated by small ui
// preference values, not mail content) so the file stays a complete,
// self-contained snapshot.
func (m *Manager) pushSettings(ctx context.Context, cfg Config) error {
	all, err := m.store.AllSettings(ctx)
	if err != nil {
		return err
	}
	snap := settingsSnapshot{Settings: make(map[string]settingEntry, len(all))}
	for _, s := range all {
		if s.Key == settingKey {
			continue
		}
		snap.Settings[s.Key] = settingEntry{Value: s.Value, UpdatedAt: s.UpdatedAt}
	}
	return writeJSONFile(m.settingsPath(cfg), snap)
}
