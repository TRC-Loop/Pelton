package desktop

import (
	"errors"

	"github.com/TRC-Loop/Pelton/internal/credentials"
	pimap "github.com/TRC-Loop/Pelton/internal/imap"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// errEmptyPassword rejects a blank password rather than storing one that can
// never authenticate.
var errEmptyPassword = errors.New("pelton: enter a password")

// errAccountUsesOAuth means the account signs in with a provider token, so a
// password would be the wrong credential entirely.
var errAccountUsesOAuth = errors.New("pelton: this mailbox signs in with your provider, not a password")

// UpdateAccountRequest carries the editable fields of an existing account. The
// email address is intentionally not editable here: it keys folder/message
// ownership and the display identity, so changing it belongs to a re-add rather
// than an in-place edit.
type UpdateAccountRequest struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"displayName"`
	// Username is the login name when it differs from the email; empty logs in
	// with the account's email.
	Username string `json:"username"`
	IMAPHost string `json:"imapHost"`
	IMAPPort int    `json:"imapPort"`
	SMTPHost string `json:"smtpHost"`
	SMTPPort int    `json:"smtpPort"`
	// Password sets a new login password. Empty leaves whatever is stored
	// alone, so an edit that only moves the server ports does not have to
	// re-enter it. An account imported from another client has no stored
	// password at all, and this is how it gets one.
	Password string `json:"password"`
	// IMAPTLS and SMTPTLS pin the connection security: "ssl", "starttls", or
	// empty to derive it from the port.
	IMAPTLS string `json:"imapTls"`
	SMTPTLS string `json:"smtpTls"`
	// ExportOnArchive and the fields under it configure the local .eml copy
	// written when a message is archived. An empty ExportDir turns the export
	// off whatever the toggle says, since there would be nowhere to write.
	ExportOnArchive    bool   `json:"exportOnArchive"`
	ExportDir          string `json:"exportDir"`
	ExportSubfolders   string `json:"exportSubfolders"`
	ExportNameTemplate string `json:"exportNameTemplate"`
	// PGPDefault is the account's starting point for protecting outgoing mail:
	// '' unprotected, 'sign', or 'auto'. An unknown value is treated as ''.
	PGPDefault string `json:"pgpDefault"`
}

// UpdateAccount persists edits to an account's display name and server settings.
// It loads the row first so the immutable email is preserved regardless of the
// request. Server changes take effect on the next connection (existing idle
// sessions keep their current settings until they reconnect).
func (a *App) UpdateAccount(req UpdateAccountRequest) (AccountDTO, error) {
	if err := a.ready(); err != nil {
		return AccountDTO{}, err
	}
	if !validTLSMode(req.IMAPTLS) || !validTLSMode(req.SMTPTLS) {
		return AccountDTO{}, errUnknownTLSMode
	}
	account, err := a.store.GetAccount(a.ctx, req.ID)
	if err != nil {
		return AccountDTO{}, err
	}
	account.DisplayName = req.DisplayName
	account.Username = req.Username
	account.IMAPHost = req.IMAPHost
	account.IMAPPort = req.IMAPPort
	account.SMTPHost = req.SMTPHost
	account.SMTPPort = req.SMTPPort
	account.IMAPTLS = req.IMAPTLS
	account.SMTPTLS = req.SMTPTLS
	if err := a.store.UpdateAccount(a.ctx, account); err != nil {
		return AccountDTO{}, err
	}
	account.ExportOnArchive = req.ExportOnArchive && req.ExportDir != ""
	account.ExportDir = req.ExportDir
	account.ExportSubfolders = validSubfolderMode(req.ExportSubfolders)
	account.ExportNameTemplate = req.ExportNameTemplate
	if err := a.store.SetAccountArchiveExport(a.ctx, account.ID, account.ExportOnArchive,
		account.ExportDir, account.ExportSubfolders, account.ExportNameTemplate); err != nil {
		return AccountDTO{}, err
	}
	account.PGPDefault = validPGPDefault(req.PGPDefault)
	if err := a.store.SetAccountPGPDefault(a.ctx, account.ID, account.PGPDefault); err != nil {
		return AccountDTO{}, err
	}
	if req.Password != "" {
		if err := a.SetAccountPassword(req.ID, req.Password); err != nil {
			return AccountDTO{}, err
		}
	}
	return toAccountDTO(*account), nil
}

// SetAccountPassword stores a login password for an account, replacing whatever
// was there. It refuses to overwrite an OAuth secret with a password, since
// that would silently downgrade a working Gmail or Outlook login to one the
// provider will reject.
func (a *App) SetAccountPassword(accountID int64, password string) error {
	if err := a.ready(); err != nil {
		return err
	}
	if password == "" {
		return errEmptyPassword
	}
	if existing, err := credentials.Load(accountID); err == nil && existing.Method == credentials.MethodOAuth {
		return errAccountUsesOAuth
	}
	return credentials.Store(accountID, credentials.Secret{
		Method:   credentials.MethodPassword,
		Password: password,
	})
}

// needsPassword decides whether one account should be prompted for, given the
// error its keyring lookup returned. It is separate from the loop so it can be
// tested without reading the developer's real keyring.
//
// Only a definite "nothing stored" counts. A keyring that is locked or broken
// returns some other error, and prompting then would ask the user to retype a
// password they already have.
func (a *App) needsPassword(acct storage.Account, secretErr error) bool {
	if !errors.Is(secretErr, credentials.ErrNotFound) {
		return false
	}
	// the legacy cli account takes its password from the environment and is
	// not actually missing one.
	if _, err := a.imapFromEnv(pimap.Config{Username: loginName(acct)}); err == nil {
		return false
	}
	return true
}

// AccountsNeedingPassword lists the accounts that cannot connect because no
// password was ever stored for them. Importing from another mail client creates
// the account but cannot take its password, so those arrive here; the frontend
// prompts for one instead of letting every sync fail quietly.
func (a *App) AccountsNeedingPassword() ([]AccountDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	accounts, err := a.store.ListAccounts(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AccountDTO, 0)
	for _, acct := range accounts {
		_, secretErr := credentials.Load(acct.ID)
		if !a.needsPassword(acct, secretErr) {
			continue
		}
		out = append(out, toAccountDTO(acct))
	}
	return out, nil
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
