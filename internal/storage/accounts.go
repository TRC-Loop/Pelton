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
	Username  string
	IMAPHost  string
	IMAPPort  int
	SMTPHost  string
	SMTPPort  int
	CreatedAt time.Time
}

// CreateAccount inserts an account and returns its new id. CreatedAt is set to
// now if the caller left it zero.
func (d *DB) CreateAccount(ctx context.Context, a *Account) (int64, error) {
	created := a.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}

	const query = `
INSERT INTO accounts (email, display_name, username, imap_host, imap_port, smtp_host, smtp_port, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := d.sql.ExecContext(ctx, query,
		a.Email, a.DisplayName, a.Username, a.IMAPHost, a.IMAPPort, a.SMTPHost, a.SMTPPort, formatTime(created))
	if err != nil {
		return 0, fmt.Errorf("storage: insert account %q: %w", a.Email, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("storage: account insert id: %w", err)
	}
	a.ID = id
	a.CreatedAt = created
	return id, nil
}

// GetAccount returns one account by id, or ErrAccountNotFound.
func (d *DB) GetAccount(ctx context.Context, id int64) (*Account, error) {
	const query = `
SELECT id, email, display_name, username, imap_host, imap_port, smtp_host, smtp_port, created_at
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

// ListAccounts returns all accounts ordered by id.
func (d *DB) ListAccounts(ctx context.Context) ([]Account, error) {
	const query = `
SELECT id, email, display_name, username, imap_host, imap_port, smtp_host, smtp_port, created_at
FROM accounts ORDER BY id`
	rows, err := d.sql.QueryContext(ctx, query)
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
SET email = ?, display_name = ?, username = ?, imap_host = ?, imap_port = ?, smtp_host = ?, smtp_port = ?
WHERE id = ?`
	res, err := d.sql.ExecContext(ctx, query,
		a.Email, a.DisplayName, a.Username, a.IMAPHost, a.IMAPPort, a.SMTPHost, a.SMTPPort, a.ID)
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

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanAccount(row rowScanner) (*Account, error) {
	var (
		a       Account
		created string
	)
	if err := row.Scan(&a.ID, &a.Email, &a.DisplayName, &a.Username, &a.IMAPHost, &a.IMAPPort,
		&a.SMTPHost, &a.SMTPPort, &created); err != nil {
		return nil, err
	}
	t, err := parseTime(created)
	if err != nil {
		return nil, err
	}
	a.CreatedAt = t
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
