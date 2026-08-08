package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrFolderNotFound is returned when a folder id has no row.
var ErrFolderNotFound = errors.New("storage: folder not found")

// Well known mailbox names. Servers may differ, but these are the common
// special use folders callers reach for when one is not reported explicitly.
const (
	FolderInbox   = "INBOX"
	FolderSent    = "Sent"
	FolderDrafts  = "Drafts"
	FolderTrash   = "Trash"
	FolderJunk    = "Junk"
	FolderArchive = "Archive"
)

// attributeSeparator joins folder attributes into the single text column.
const attributeSeparator = " "

// Folder is one mailbox in an account's hierarchy.
type Folder struct {
	ID        int64
	AccountID int64
	Name      string
	// IMAPPath is the raw mailbox name the server returned. Hierarchy is encoded
	// with Delimiter, which varies per server, so neither is ever assumed.
	IMAPPath    string
	Delimiter   string
	ParentID    *int64
	Attributes  []string
	UIDValidity uint32
}

// CreateFolder inserts a folder and returns its new id.
func (d *DB) CreateFolder(ctx context.Context, f *Folder) (int64, error) {
	const query = `
INSERT INTO folders (account_id, name, imap_path, delimiter, parent_id, attributes, uid_validity)
VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := d.sql.ExecContext(ctx, query,
		f.AccountID, f.Name, f.IMAPPath, f.Delimiter, nullableID(f.ParentID),
		joinAttributes(f.Attributes), f.UIDValidity)
	if err != nil {
		return 0, fmt.Errorf("storage: insert folder %q: %w", f.IMAPPath, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("storage: folder insert id: %w", err)
	}
	f.ID = id
	return id, nil
}

// GetFolder returns one folder by id, or ErrFolderNotFound.
func (d *DB) GetFolder(ctx context.Context, id int64) (*Folder, error) {
	const query = `
SELECT id, account_id, name, imap_path, delimiter, parent_id, attributes, uid_validity
FROM folders WHERE id = ?`
	f, err := scanFolder(d.sql.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrFolderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("storage: get folder %d: %w", id, err)
	}
	return f, nil
}

// ListFolders returns every folder for an account ordered by id.
func (d *DB) ListFolders(ctx context.Context, accountID int64) ([]Folder, error) {
	const query = `
SELECT id, account_id, name, imap_path, delimiter, parent_id, attributes, uid_validity
FROM folders WHERE account_id = ? ORDER BY id`
	rows, err := d.sql.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("storage: list folders for account %d: %w", accountID, err)
	}
	defer rows.Close()

	var folders []Folder
	for rows.Next() {
		f, err := scanFolder(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan folder: %w", err)
		}
		folders = append(folders, *f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: iterate folders: %w", err)
	}
	return folders, nil
}

// SetFolderUIDValidity updates the stored UIDVALIDITY for a folder. A change
// means the server reset the mailbox and the cache for it is stale.
func (d *DB) SetFolderUIDValidity(ctx context.Context, id int64, uidValidity uint32) error {
	res, err := d.sql.ExecContext(ctx,
		`UPDATE folders SET uid_validity = ? WHERE id = ?`, uidValidity, id)
	if err != nil {
		return fmt.Errorf("storage: update uid_validity for folder %d: %w", id, err)
	}
	return requireOneRow(res, ErrFolderNotFound)
}

// RenameFolder updates a folder's display name and imap path. It does not touch
// the folder's children: the caller renames the subtree, since only it knows the
// server's delimiter (see RenameFolderSubtree).
func (d *DB) RenameFolder(ctx context.Context, id int64, name, imapPath string) error {
	res, err := d.sql.ExecContext(ctx,
		`UPDATE folders SET name = ?, imap_path = ? WHERE id = ?`, name, imapPath, id)
	if err != nil {
		return fmt.Errorf("storage: rename folder %d: %w", id, err)
	}
	return requireOneRow(res, ErrFolderNotFound)
}

// RenameFolderSubtree rewrites the imap path prefix of every descendant of a
// renamed folder, which an imap RENAME moves along with their parent. oldPrefix
// and newPrefix are the parent paths with the server's delimiter already
// appended, so the match cannot catch a sibling whose name merely starts with
// the same text. Returns the number of rows rewritten.
func (d *DB) RenameFolderSubtree(ctx context.Context, accountID int64, oldPrefix, newPrefix string) (int, error) {
	// escape the like wildcards in the stored path so a folder legitimately
	// containing % or _ does not match half the account.
	const query = `
UPDATE folders
SET imap_path = ? || substr(imap_path, ?)
WHERE account_id = ? AND imap_path LIKE ? ESCAPE '\'`
	res, err := d.sql.ExecContext(ctx, query,
		newPrefix, len(oldPrefix)+1, accountID, escapeLike(oldPrefix)+"%")
	if err != nil {
		return 0, fmt.Errorf("storage: rename folder subtree %q: %w", oldPrefix, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("storage: rename subtree rows affected: %w", err)
	}
	return int(n), nil
}

// FolderDescendants returns every folder nested under a folder, deepest last,
// so a caller can tear a subtree down parent-last or build it parent-first. The
// folder itself is not included.
func (d *DB) FolderDescendants(ctx context.Context, accountID int64, prefix string) ([]Folder, error) {
	const query = `
SELECT id, account_id, name, imap_path, delimiter, parent_id, attributes, uid_validity
FROM folders
WHERE account_id = ? AND imap_path LIKE ? ESCAPE '\'
ORDER BY length(imap_path)`
	rows, err := d.sql.QueryContext(ctx, query, accountID, escapeLike(prefix)+"%")
	if err != nil {
		return nil, fmt.Errorf("storage: list folder descendants of %q: %w", prefix, err)
	}
	defer rows.Close()

	var folders []Folder
	for rows.Next() {
		f, err := scanFolder(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan folder descendant: %w", err)
		}
		folders = append(folders, *f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: iterate folder descendants: %w", err)
	}
	return folders, nil
}

// DeleteFolder removes a folder; its messages and their attachment rows cascade.
func (d *DB) DeleteFolder(ctx context.Context, id int64) error {
	res, err := d.sql.ExecContext(ctx, `DELETE FROM folders WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("storage: delete folder %d: %w", id, err)
	}
	return requireOneRow(res, ErrFolderNotFound)
}

func scanFolder(row rowScanner) (*Folder, error) {
	var (
		f      Folder
		parent sql.NullInt64
		attrs  string
	)
	if err := row.Scan(&f.ID, &f.AccountID, &f.Name, &f.IMAPPath, &f.Delimiter,
		&parent, &attrs, &f.UIDValidity); err != nil {
		return nil, err
	}
	if parent.Valid {
		f.ParentID = &parent.Int64
	}
	f.Attributes = splitAttributes(attrs)
	return &f, nil
}

func joinAttributes(attrs []string) string {
	return strings.Join(attrs, attributeSeparator)
}

func splitAttributes(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, attributeSeparator)
}

// nullableID maps an optional id to a value sql can store as NULL.
func nullableID(id *int64) any {
	if id == nil {
		return nil
	}
	return *id
}
