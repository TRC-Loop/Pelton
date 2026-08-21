package sync

import (
	"context"
	"errors"
	"fmt"

	"github.com/emersion/go-imap/v2"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

// repairBatch caps how many mangled messages one folder sync refetches. The
// mark is left on the rest, so a mailbox full of them is repaired over several
// syncs rather than turning one into a full redownload.
const repairBatch = 50

// repairMangled refetches messages that were cached before charset detection
// existed and whose stored text is not valid utf-8. Only the text is replaced:
// the message is the same message, so its flags, colour and attachments stay
// as they are.
//
// A message the server no longer has loses its mark instead, otherwise every
// sync from here on would try it again.
func (e *Engine) repairMangled(ctx context.Context, folder storage.Folder, res *FolderSyncResult) {
	broken, err := e.store.MessagesNeedingRefetch(ctx, folder.ID, repairBatch)
	if err != nil {
		e.log.Error("list messages needing refetch", "folder", folder.IMAPPath, "err", err)
		return
	}
	for _, m := range broken {
		if err := ctx.Err(); err != nil {
			return
		}
		if err := e.repairOne(ctx, m); err != nil {
			e.log.Error("refetch mangled message", "folder", folder.IMAPPath, "uid", m.UID, "err", err)
			continue
		}
		res.Repaired++
		res.RepairedIDs = append(res.RepairedIDs, m.ID)
	}
}

func (e *Engine) repairOne(ctx context.Context, m storage.MangledMessage) error {
	msg, err := e.client.FetchMessage(imap.UID(m.UID))
	if err != nil {
		// gone from the server, or unreadable: either way there is nothing to
		// repair it from, so stop asking.
		if clearErr := e.store.ClearRefetchMark(ctx, m.ID); clearErr != nil {
			return errors.Join(err, clearErr)
		}
		return fmt.Errorf("sync: refetch message uid %d: %w", m.UID, err)
	}
	if err := e.store.RepairMessageText(ctx, m.ID, msg.Subject, msg.Text, msg.HTML, msg.CharsetGuess); err != nil {
		return err
	}
	return nil
}
