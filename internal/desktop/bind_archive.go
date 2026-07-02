package desktop

import (
	"fmt"

	"github.com/emersion/go-imap/v2"

	pimap "github.com/TRC-Loop/Pelton/internal/imap"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// ArchiveUndoDTO carries what undo-archive needs to move a message back: its
// stable rfc Message-ID (the moved copy has a new UID) and the folder it came
// from. MessageID is empty when the message had no Message-ID header, in which
// case undo is not possible.
type ArchiveUndoDTO struct {
	MessageID        string `json:"messageId"`
	OriginalFolderID int64  `json:"originalFolderId"`
}

// ArchiveMessage moves a message to its account's Archive folder on the server.
// It errors clearly when the account has no archive folder. It returns the info
// needed to undo the archive.
func (a *App) ArchiveMessage(id int64) (ArchiveUndoDTO, error) {
	if err := a.ready(); err != nil {
		return ArchiveUndoDTO{}, err
	}
	m, err := a.store.GetMessage(a.ctx, id)
	if err != nil {
		return ArchiveUndoDTO{}, err
	}
	archive, err := a.findArchiveFolder(m.AccountID)
	if err != nil {
		return ArchiveUndoDTO{}, err
	}
	return a.moveMessageTo(m, archive)
}

// MoveMessage moves a message to any folder (of its own account) on the server,
// dropping the local row so it leaves the current view immediately; the next sync
// re-fetches it under the destination. It returns the info needed to undo.
func (a *App) MoveMessage(id, destFolderID int64) (ArchiveUndoDTO, error) {
	if err := a.ready(); err != nil {
		return ArchiveUndoDTO{}, err
	}
	m, err := a.store.GetMessage(a.ctx, id)
	if err != nil {
		return ArchiveUndoDTO{}, err
	}
	dest, err := a.store.GetFolder(a.ctx, destFolderID)
	if err != nil {
		return ArchiveUndoDTO{}, err
	}
	if dest.AccountID != m.AccountID {
		return ArchiveUndoDTO{}, fmt.Errorf("pelton: cannot move a message to another account's folder")
	}
	return a.moveMessageTo(m, *dest)
}

// moveMessageTo performs the server move of a cached message to dest, then drops
// the local row and its files. It is the shared core of archive and move. Moving
// to the message's current folder is a no-op.
func (a *App) moveMessageTo(m *storage.Message, dest storage.Folder) (ArchiveUndoDTO, error) {
	source, err := a.store.GetFolder(a.ctx, m.FolderID)
	if err != nil {
		return ArchiveUndoDTO{}, err
	}
	if dest.ID == source.ID {
		return ArchiveUndoDTO{}, nil // already there
	}
	account, err := a.store.GetAccount(a.ctx, m.AccountID)
	if err != nil {
		return ArchiveUndoDTO{}, err
	}
	cfg, err := a.resolveIMAP(*account)
	if err != nil {
		return ArchiveUndoDTO{}, err
	}

	syncMu.Lock()
	defer syncMu.Unlock()

	client, err := pimap.Connect(cfg)
	if err != nil {
		return ArchiveUndoDTO{}, err
	}
	defer client.Close()
	if err := client.Login(); err != nil {
		return ArchiveUndoDTO{}, err
	}
	defer client.Logout()
	if _, err := client.Select(source.IMAPPath); err != nil {
		return ArchiveUndoDTO{}, fmt.Errorf("move: select %q: %w", source.IMAPPath, err)
	}
	if err := client.Move(imap.UID(m.UID), dest.IMAPPath); err != nil {
		return ArchiveUndoDTO{}, err
	}

	if err := a.store.DeleteMessage(a.ctx, m.ID); err != nil {
		return ArchiveUndoDTO{}, err
	}
	if err := a.store.DeleteAttachmentFilesForMessage(m.AccountID, m.ID); err != nil {
		a.log.Error("move: remove attachment files", "id", m.ID, "err", err)
	}
	a.emit(EventMailNew, MailNewEvent{AccountID: m.AccountID, FolderID: dest.ID, Count: 1})
	return ArchiveUndoDTO{MessageID: m.MessageID, OriginalFolderID: source.ID}, nil
}

// UnarchiveMessage moves an archived message back to originalFolderID, locating
// it by its rfc Message-ID (its UID changed when it was moved). It errors when
// the message has no Message-ID or can no longer be found in Archive.
func (a *App) UnarchiveMessage(rfcMessageID string, originalFolderID int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	if rfcMessageID == "" {
		return fmt.Errorf("pelton: this message cannot be un-archived (no Message-ID)")
	}
	dest, err := a.store.GetFolder(a.ctx, originalFolderID)
	if err != nil {
		return err
	}
	archive, err := a.findArchiveFolder(dest.AccountID)
	if err != nil {
		return err
	}
	account, err := a.store.GetAccount(a.ctx, dest.AccountID)
	if err != nil {
		return err
	}
	cfg, err := a.resolveIMAP(*account)
	if err != nil {
		return err
	}

	syncMu.Lock()
	defer syncMu.Unlock()

	client, err := pimap.Connect(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Login(); err != nil {
		return err
	}
	defer client.Logout()
	if _, err := client.Select(archive.IMAPPath); err != nil {
		return fmt.Errorf("unarchive: select %q: %w", archive.IMAPPath, err)
	}
	uids, err := client.SearchByMessageID(rfcMessageID)
	if err != nil {
		return err
	}
	if len(uids) == 0 {
		return fmt.Errorf("pelton: archived message not found to restore")
	}
	if err := client.Move(uids[len(uids)-1], dest.IMAPPath); err != nil {
		return err
	}
	a.emit(EventMailNew, MailNewEvent{AccountID: dest.AccountID, FolderID: dest.ID, Count: 1})
	return nil
}

// findArchiveFolder returns the account's archive-role folder, or an error when
// none exists.
func (a *App) findArchiveFolder(accountID int64) (storage.Folder, error) {
	folders, err := a.store.ListFolders(a.ctx, accountID)
	if err != nil {
		return storage.Folder{}, err
	}
	for _, f := range folders {
		if folderRole(f) == roleArchive {
			return f, nil
		}
	}
	return storage.Folder{}, fmt.Errorf("pelton: this account has no Archive folder")
}
