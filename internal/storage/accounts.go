package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrAccountNotFound is returned when an account id has no row.
var ErrAccountNotFound = errors.New("storage: account not found")

// Account is non sensitive account metadata. Passwords and tokens are never
// stored here, they live in the os keyring keyed by this row's ID.
type Account struct {
	ID          int64
	Email       string
	DisplayName string
	// Username is the login name when it differs from the email address. Empty
	// means authenticate with Email.
	Username string
	IMAPHost string
	IMAPPort int
	SMTPHost string
	SMTPPort int
	// IMAPTLS and SMTPTLS pin the connection security: "ssl" for implicit TLS,
	// "starttls" to upgrade a cleartext connection. Empty derives it from the
	// port, which is what every account created before this was storable does.
	IMAPTLS   string
	SMTPTLS   string
	CreatedAt time.Time
	// Position is the account section's rank in the sidebar, or 0 until the user
	// reorders them. Unpositioned accounts sort after positioned ones by id, so
	// an install that never reordered keeps creation order.
	Position int
	// ExportOnArchive writes a local .eml copy of every message archived from
	// this account. ExportDir is where they go and must be set for the export to
	// run; ExportSubfolders is one of the mailexport subfolder modes and
	// ExportNameTemplate its file name pattern, both empty meaning the default.
	ExportOnArchive    bool
	ExportDir          string
	ExportSubfolders   string
	ExportNameTemplate string
	// PGPDefault is how this account starts a new message: '' unprotected,
	// 'sign' to sign when the account has a key, 'auto' to sign and encrypt
	// whenever every recipient has one.
	PGPDefault string
	// PasswordPromptDismissed records that the user asked the missing-password
	// prompt to stop interrupting for this account. The account still cannot
	// sync; the ui marks it instead of asking. Storing a password clears it.
	PasswordPromptDismissed bool
	// Local marks the Local Folders account, which holds imported mail and has
	// no server behind it. Sync, idle and the mailbox backup all skip it, and
	// its Email is LocalAccountEmail rather than a real address.
	Local bool
}

// LocalAccountEmail is the reserved address of the Local Folders account. It is
// not a routable address; it exists so the account row has the stable, unique
// identifier every other lookup in the app keys off.
const LocalAccountEmail = "local@pelton.invalid"

// LocalAccountName is the Local Folders account's stored display name. The ui
// localizes the label it shows, this is only the fallback.
const LocalAccountName = "Local Folders"

// accountColumns is the select list every account query shares, in the order
// scanAccount reads them.
const accountColumns = `id, email, display_name, username, imap_host, imap_port,
       smtp_host, smtp_port, imap_tls, smtp_tls, created_at, position, is_local,
       export_on_archive, export_dir, export_subfolders, export_name_template,
       pgp_default, password_prompt_dismissed`

// CreateAccount inserts an account and returns its new id. CreatedAt is set to
// now if the caller left it zero.
func (d *DB) CreateAccount(ctx context.Context, a *Account) (int64, error) {
	created := a.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}

	const query = `
INSERT INTO accounts (email, display_name, username, imap_host, imap_port, smtp_host, smtp_port, imap_tls, smtp_tls, created_at, is_local)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := d.sql.ExecContext(ctx, query,
		a.Email, a.DisplayName, a.Username, a.IMAPHost, a.IMAPPort, a.SMTPHost, a.SMTPPort,
		a.IMAPTLS, a.SMTPTLS, formatTime(created), boolToInt(a.Local))
	if err != nil {
		return 0, fmt.Errorf("storage: insert account %q: %w", a.Email, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("storage: account insert id: %w", err)
	}
	a.ID = id
	a.CreatedAt = created
	// an account nobody can see would be invisible mail, so it joins the profile
	// that created it. Other profiles opt in from the profile editor.
	if err := d.AddAccountToProfiles(ctx, id, []int64{d.ScopedProfileID()}); err != nil {
		return 0, err
	}
	return id, nil
}

// GetAccount returns one account by id, or ErrAccountNotFound.
func (d *DB) GetAccount(ctx context.Context, id int64) (*Account, error) {
	const query = `
SELECT ` + accountColumns + `
FROM accounts WHERE id = ?`
	a, err := scanAccount(d.sql.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("storage: get account %d: %w", id, err)
	}
	return a, nil
}

// ListAccounts returns all accounts in sidebar order: the ones the user
// reordered first, in that order, then the rest by id.
func (d *DB) ListAccounts(ctx context.Context) ([]Account, error) {
	const query = `
SELECT ` + accountColumns + `
FROM accounts
WHERE id IN (SELECT account_id FROM profile_accounts WHERE profile_id = ?)
ORDER BY position = 0, position, id`
	return d.listAccounts(ctx, query, d.ScopedProfileID())
}

// ListAllAccounts returns every account on the install, whether or not the
// current profile shows it. Only the profile editor and account deletion want
// this; everything else works with what the profile can see.
func (d *DB) ListAllAccounts(ctx context.Context) ([]Account, error) {
	const query = `
SELECT ` + accountColumns + `
FROM accounts ORDER BY position = 0, position, id`
	return d.listAccounts(ctx, query)
}

func (d *DB) listAccounts(ctx context.Context, query string, args ...any) ([]Account, error) {
	rows, err := d.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan account: %w", err)
		}
		accounts = append(accounts, *a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: iterate accounts: %w", err)
	}
	return accounts, nil
}

// UpdateAccount updates the mutable fields of an existing account.
func (d *DB) UpdateAccount(ctx context.Context, a *Account) error {
	const query = `
UPDATE accounts
SET email = ?, display_name = ?, username = ?, imap_host = ?, imap_port = ?, smtp_host = ?, smtp_port = ?,
    imap_tls = ?, smtp_tls = ?
WHERE id = ?`
	res, err := d.sql.ExecContext(ctx, query,
		a.Email, a.DisplayName, a.Username, a.IMAPHost, a.IMAPPort, a.SMTPHost, a.SMTPPort,
		a.IMAPTLS, a.SMTPTLS, a.ID)
	if err != nil {
		return fmt.Errorf("storage: update account %d: %w", a.ID, err)
	}
	return requireOneRow(res, ErrAccountNotFound)
}

// DeleteAccount removes an account. Its folders, messages and attachment rows
// cascade away; attachment files on disk are the caller's concern.
func (d *DB) DeleteAccount(ctx context.Context, id int64) error {
	res, err := d.sql.ExecContext(ctx, `DELETE FROM accounts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("storage: delete account %d: %w", id, err)
	}
	return requireOneRow(res, ErrAccountNotFound)
}

// SetAccountArchiveExport stores an account's export-on-archive configuration.
// Turning it on with an empty dir is allowed but exports nothing: the caller
// decides whether to insist on a directory.
func (d *DB) SetAccountArchiveExport(ctx context.Context, id int64, on bool, dir, subfolders, template string) error {
	const query = `
UPDATE accounts
SET export_on_archive = ?, export_dir = ?, export_subfolders = ?, export_name_template = ?
WHERE id = ?`
	res, err := d.sql.ExecContext(ctx, query, boolToInt(on), dir, subfolders, template, id)
	if err != nil {
		return fmt.Errorf("storage: set account %d archive export: %w", id, err)
	}
	return requireOneRow(res, ErrAccountNotFound)
}

// SetAccountPGPDefault stores how an account starts a new message: ” for
// unprotected, 'sign' or 'auto'. Validating the value is the caller's job.
func (d *DB) SetAccountPGPDefault(ctx context.Context, id int64, value string) error {
	res, err := d.sql.ExecContext(ctx, `UPDATE accounts SET pgp_default = ? WHERE id = ?`, value, id)
	if err != nil {
		return fmt.Errorf("storage: set account %d pgp default: %w", id, err)
	}
	return requireOneRow(res, ErrAccountNotFound)
}

// SetAccountPasswordPromptDismissed records whether the missing-password prompt
// should stay quiet for an account.
func (d *DB) SetAccountPasswordPromptDismissed(ctx context.Context, id int64, dismissed bool) error {
	res, err := d.sql.ExecContext(ctx, `UPDATE accounts SET password_prompt_dismissed = ? WHERE id = ?`, boolToInt(dismissed), id)
	if err != nil {
		return fmt.Errorf("storage: set account %d password prompt dismissed: %w", id, err)
	}
	return requireOneRow(res, ErrAccountNotFound)
}

// SetAccountPositions rewrites the sidebar order of the account sections in one
// transaction. Accounts not listed keep their current position. Positions start
// at 1, since 0 means "never reordered".
func (d *DB) SetAccountPositions(ctx context.Context, orderedIDs []int64) error {
	return d.setPositions(ctx, `UPDATE accounts SET position = ? WHERE id = ?`, orderedIDs, "accounts")
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanAccount(row rowScanner) (*Account, error) {
	var (
		a         Account
		created   string
		local     int
		exportOn  int
		dismissed int
	)
	if err := row.Scan(&a.ID, &a.Email, &a.DisplayName, &a.Username, &a.IMAPHost, &a.IMAPPort,
		&a.SMTPHost, &a.SMTPPort, &a.IMAPTLS, &a.SMTPTLS, &created, &a.Position, &local,
		&exportOn, &a.ExportDir, &a.ExportSubfolders, &a.ExportNameTemplate,
		&a.PGPDefault, &dismissed); err != nil {
		return nil, err
	}
	a.ExportOnArchive = exportOn != 0
	a.PasswordPromptDismissed = dismissed != 0
	t, err := parseTime(created)
	if err != nil {
		return nil, err
	}
	a.CreatedAt = t
	a.Local = local != 0
	return &a, nil
}

// requireOneRow turns a no rows affected result into notFound, so updates and
// deletes against a missing id report it instead of succeeding silently.
func requireOneRow(res sql.Result, notFound error) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("storage: rows affected: %w", err)
	}
	if n == 0 {
		return notFound
	}
	return nil
}
