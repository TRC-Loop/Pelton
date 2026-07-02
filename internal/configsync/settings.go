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
func (m *Manager) pullSettings(ctx context.Context, cfg Config) error {
	var snap settingsSnapshot
	found, err := readJSONFile(m.settingsPath(cfg), &snap)
	if err != nil || !found {
		return err
	}
	for key, entry := range snap.Settings {
		if err := m.store.SetIfNewer(ctx, key, entry.Value, entry.UpdatedAt); err != nil {
			return err
		}
	}
	return nil
}

// pushSettings writes every local setting out to the folder. It always
// exports the full set (dominated by small UI preference values, not mail
// content) so the file stays a complete, self-contained snapshot.
func (m *Manager) pushSettings(ctx context.Context, cfg Config) error {
	all, err := m.store.AllSettings(ctx)
	if err != nil {
		return err
	}
	snap := settingsSnapshot{Settings: make(map[string]settingEntry, len(all))}
	for _, s := range all {
		snap.Settings[s.Key] = settingEntry{Value: s.Value, UpdatedAt: s.UpdatedAt}
	}
	return writeJSONFile(m.settingsPath(cfg), snap)
}
