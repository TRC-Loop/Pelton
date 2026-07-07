package desktop

import (
	"github.com/TRC-Loop/Pelton/internal/credentials"
)

// UpdateAccountRequest carries the editable fields of an existing account. The
// email address is intentionally not editable here: it keys folder/message
// ownership and the display identity, so changing it belongs to a re-add rather
// than an in-place edit.
type UpdateAccountRequest struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"displayName"`
	IMAPHost    string `json:"imapHost"`
	IMAPPort    int    `json:"imapPort"`
	SMTPHost    string `json:"smtpHost"`
	SMTPPort    int    `json:"smtpPort"`
}

// UpdateAccount persists edits to an account's display name and server settings.
// It loads the row first so the immutable email is preserved regardless of the
// request. Server changes take effect on the next connection (existing idle
// sessions keep their current settings until they reconnect).
func (a *App) UpdateAccount(req UpdateAccountRequest) (AccountDTO, error) {
	if err := a.ready(); err != nil {
		return AccountDTO{}, err
	}
	account, err := a.store.GetAccount(a.ctx, req.ID)
	if err != nil {
		return AccountDTO{}, err
	}
	account.DisplayName = req.DisplayName
	account.IMAPHost = req.IMAPHost
	account.IMAPPort = req.IMAPPort
	account.SMTPHost = req.SMTPHost
	account.SMTPPort = req.SMTPPort
	if err := a.store.UpdateAccount(a.ctx, account); err != nil {
		return AccountDTO{}, err
	}
	return toAccountDTO(*account), nil
}

// DeleteAccount removes an account entirely: its keyring secret, its cached mail
// (folders, messages and attachment rows cascade in the db) and its attachment
// files on disk. Deleting the keyring secret also lets any running idle loop for
// the account exit cleanly, since it stops on a missing-credentials error.
func (a *App) DeleteAccount(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	// drop the secret first so a still-running idle loop unwinds on its next
	// reconnect instead of retrying against a half-deleted account.
	if err := credentials.Delete(id); err != nil {
		a.log.Error("delete credentials", "account", id, "err", err)
	}
	if err := a.store.DeleteAccount(a.ctx, id); err != nil {
		return err
	}
	if err := a.store.DeleteAttachmentFilesForAccount(id); err != nil {
		a.log.Error("delete attachment files", "account", id, "err", err)
	}
	return nil
}
