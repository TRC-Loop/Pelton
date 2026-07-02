package configsync

import (
	"context"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

// metadataSnapshot is the file format for message-metadata.json: locally-only
// message state (color labels, offline pins, snoozes) keyed by account email
// and Message-ID so it means the same thing on any device.
type metadataSnapshot struct {
	Messages []storage.MessageExtra `json:"messages"`
}

func (m *Manager) metadataPath(cfg Config) string {
	return joinPath(cfg.Path, metadataFileName)
}

// pullMetadata applies every remote message extra to whichever local message
// matches its account email and Message-ID. Messages not yet downloaded on
// this device are silently skipped; ApplyMessageExtra is a no-op for them.
func (m *Manager) pullMetadata(ctx context.Context, cfg Config) error {
	var snap metadataSnapshot
	found, err := readJSONFile(m.metadataPath(cfg), &snap)
	if err != nil || !found {
		return err
	}
	for _, e := range snap.Messages {
		if err := m.store.ApplyMessageExtra(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

// pushMetadata exports every message with any non-default local state.
func (m *Manager) pushMetadata(ctx context.Context, cfg Config) error {
	extras, err := m.store.ListMessageExtras(ctx)
	if err != nil {
		return err
	}
	return writeJSONFile(m.metadataPath(cfg), metadataSnapshot{Messages: extras})
}
