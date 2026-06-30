package sync

import (
	"context"
	"fmt"

	"github.com/emersion/go-imap/v2"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

// pushFlags stores the merged flags locally, adds them on the server and clears
// the pending marker.
//
// the server push is additive (STORE +FLAGS), never a full replace, so flags we
// do not model (\Answered, \Draft, keywords) are left untouched. combined with
// the union merge policy this means we only ever set flags on the server, never
// clear them, see the policy note in reconcile.go.
func (e *Engine) pushFlags(ctx context.Context, state storage.MessageState, flags storage.Flag) error {
	if err := e.store.SetMessageFlags(ctx, state.ID, flags); err != nil {
		return fmt.Errorf("sync: store merged flags for uid %d: %w", state.UID, err)
	}
	if imapFlags := storageFlagsToImap(flags); len(imapFlags) > 0 {
		if err := e.client.AddFlags(imap.UID(state.UID), imapFlags...); err != nil {
			return fmt.Errorf("sync: push flags for uid %d: %w", state.UID, err)
		}
	}
	if err := e.store.ClearFlagsPending(ctx, state.ID); err != nil {
		return fmt.Errorf("sync: clear pending flags for uid %d: %w", state.UID, err)
	}
	return nil
}

// clearPending stores the merged flags locally and clears the marker, for the
// case where the server already had everything the local change wanted.
func (e *Engine) clearPending(ctx context.Context, state storage.MessageState, flags storage.Flag) error {
	if err := e.store.SetMessageFlags(ctx, state.ID, flags); err != nil {
		return fmt.Errorf("sync: store merged flags for uid %d: %w", state.UID, err)
	}
	if err := e.store.ClearFlagsPending(ctx, state.ID); err != nil {
		return fmt.Errorf("sync: clear pending flags for uid %d: %w", state.UID, err)
	}
	return nil
}

// pushDeletes deletes the given messages on the server, then removes them from
// the cache. it marks all uids \Deleted in one STORE and expunges them in one
// call, so a folder with many local deletions costs two round trips, not two
// per message.
//
// decision: delete means \Deleted + EXPUNGE here, not move-to-Trash. it is the
// standard, provider neutral delete. the known divergence is gmail, where this
// only removes a label inside an ordinary mailbox, see Expunge in the imap
// package. moving to the account's Trash folder is the cleaner gmail behaviour
// and is a candidate for a later version.
func (e *Engine) pushDeletes(ctx context.Context, folder storage.Folder, states []storage.MessageState) error {
	if len(states) == 0 {
		return nil
	}

	uids := make([]imap.UID, 0, len(states))
	for _, s := range states {
		uids = append(uids, imap.UID(s.UID))
	}

	if err := e.client.MarkDeleted(uids...); err != nil {
		return fmt.Errorf("sync: mark deleted on server: %w", err)
	}
	if err := e.client.Expunge(uids...); err != nil {
		return fmt.Errorf("sync: expunge on server: %w", err)
	}

	// server delete succeeded, now drop the local rows and their files.
	for _, s := range states {
		if err := e.deleteLocal(ctx, folder, s); err != nil {
			return err
		}
	}
	return nil
}
