package sync

import (
	"bytes"
	"context"
	"fmt"

	"github.com/emersion/go-imap/v2"

	pimap "github.com/TRC-Loop/Pelton/internal/imap"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// imapFlagsToStorage maps the imap flag list to the storage bitmask, keeping
// only the flags Pelton models. unknown server flags (\Answered, \Draft, custom
// keywords) are intentionally ignored here.
func imapFlagsToStorage(flags []imap.Flag) storage.Flag {
	var out storage.Flag
	for _, f := range flags {
		switch f {
		case imap.FlagSeen:
			out |= storage.FlagSeen
		case imap.FlagFlagged:
			out |= storage.FlagFlagged
		case imap.FlagDeleted:
			out |= storage.FlagDeleted
		}
	}
	return out
}

// storageFlagsToImap maps the storage bitmask back to imap flags.
func storageFlagsToImap(f storage.Flag) []imap.Flag {
	var out []imap.Flag
	if f.Has(storage.FlagSeen) {
		out = append(out, imap.FlagSeen)
	}
	if f.Has(storage.FlagFlagged) {
		out = append(out, imap.FlagFlagged)
	}
	if f.Has(storage.FlagDeleted) {
		out = append(out, imap.FlagDeleted)
	}
	return out
}

// fetchAndStore pulls a full message by uid and inserts it with its
// attachments. body and attachments are fetched with BODY.PEEK so caching does
// not set \Seen on the server.
func (e *Engine) fetchAndStore(ctx context.Context, folder storage.Folder, uid uint32) error {
	msg, err := e.client.FetchMessage(imap.UID(uid))
	if err != nil {
		return fmt.Errorf("sync: fetch message uid %d: %w", uid, err)
	}

	stored := &storage.Message{
		AccountID: folder.AccountID,
		FolderID:  folder.ID,
		UID:       uint32(msg.UID),
		MessageID: msg.MessageID,
		Subject:   msg.Subject,
		// from/to are kept as formatted strings here. splitting name from address
		// is a later refinement, the storage columns already allow it.
		FromAddress: msg.From,
		ToAddresses: msg.To,
		CcAddresses: msg.Cc,
		Date:        msg.Date,
		Flags:       imapFlagsToStorage(msg.Flags),
		BodyPlain:   msg.Text,
		BodyHTML:    msg.HTML,
		SizeBytes:   msg.Size,
	}

	atts := make([]storage.IncomingAttachment, 0, len(msg.Attachments))
	for _, a := range msg.Attachments {
		atts = append(atts, storage.IncomingAttachment{
			Filename:    a.Filename,
			ContentType: a.ContentType,
			ContentID:   a.ContentID,
			Content:     bytes.NewReader(a.Content),
		})
	}

	if _, err := e.store.InsertMessageWithAttachments(ctx, stored, atts); err != nil {
		return fmt.Errorf("sync: store message uid %d: %w", uid, err)
	}
	return nil
}

// deleteLocal removes a cached message that the server no longer has, including
// its attachment files.
func (e *Engine) deleteLocal(ctx context.Context, folder storage.Folder, state storage.MessageState) error {
	if err := e.store.DeleteMessage(ctx, state.ID); err != nil {
		return fmt.Errorf("sync: delete local message uid %d: %w", state.UID, err)
	}
	if err := e.store.DeleteAttachmentFilesForMessage(folder.AccountID, state.ID); err != nil {
		return fmt.Errorf("sync: remove attachment files for uid %d: %w", state.UID, err)
	}
	return nil
}

// adoptServerFlags stores the server's flags for a message that changed on the
// server with no pending local change.
func (e *Engine) adoptServerFlags(ctx context.Context, state storage.MessageState, flags storage.Flag) error {
	if err := e.store.SetMessageFlags(ctx, state.ID, flags); err != nil {
		return fmt.Errorf("sync: adopt server flags for uid %d: %w", state.UID, err)
	}
	return nil
}

// compile-time check that the real imap client satisfies the interface sync
// depends on, so the public surface stays sufficient.
var _ mailClient = (*pimap.Client)(nil)
