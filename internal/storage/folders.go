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

// folderColumns is the select list every folder query shares, in the order
// scanFolder reads them.
const folderColumns = `id, account_id, name, imap_path, delimiter, parent_id,
       attributes, uid_validity, position, pinned_position`

// folderOrder sorts folders for display: reordered groups first in the order the
// user chose, then everything untouched in discovery (id) order. In sqlite
// `position = 0` is 1 for the untouched rows, which is why they sort last.
const folderOrder = `ORDER BY position = 0, position, id`

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
	// Position is the folder's rank among its siblings in the sidebar, or 0 when
	// the user has never reordered that group. Unpositioned folders sort after
	// positioned ones by id, so discovery order survives and a new folder lands
	// at the end rather than the front.
	Position int
	// PinnedPosition is the folder's rank in the sidebar's Pinned group, or 0
	// when it is not pinned. Pinning mirrors the folder into that group; it stays
	// in its account's tree either way.
	PinnedPosition int
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
	const query = `SELECT ` + folderColumns + ` FROM folders WHERE id = ?`
	f, err := scanFolder(d.sql.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrFolderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("storage: get folder %d: %w", id, err)
	}
	return f, nil
}

// ListFolders returns every folder for an account in sidebar order: groups the
// user has reordered first, in that order, then the rest by id.
func (d *DB) ListFolders(ctx context.Context, accountID int64) ([]Folder, error) {
	const query = `SELECT ` + folderColumns + ` FROM folders WHERE account_id = ? ` + folderOrder
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
	const query = `SELECT ` + folderColumns + `
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

// ListPinnedFolders returns every pinned folder across all accounts in the order
// the user arranged the Pinned group, shallowest rank first.
func (d *DB) ListPinnedFolders(ctx context.Context) ([]Folder, error) {
	const query = `SELECT ` + folderColumns + `
FROM folders WHERE pinned_position > 0 ORDER BY pinned_position, id`
	rows, err := d.sql.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("storage: list pinned folders: %w", err)
	}
	defer rows.Close()

	var folders []Folder
	for rows.Next() {
		f, err := scanFolder(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan pinned folder: %w", err)
		}
		folders = append(folders, *f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: iterate pinned folders: %w", err)
	}
	return folders, nil
}

// SetFolderPositions rewrites the sidebar order of a folder group in one
// transaction, so a drag reorder persists atomically. Callers pass the ids of
// one sibling group; folders not listed keep their current position. Positions
// start at 1, since 0 means "never reordered" (see folderOrder).
func (d *DB) SetFolderPositions(ctx context.Context, orderedIDs []int64) error {
	return d.setPositions(ctx, `UPDATE folders SET position = ? WHERE id = ?`, orderedIDs, "folders")
}

// SetPinnedFolderPositions rewrites the order of the Pinned group in one
// transaction. Every id passed must already be pinned; this only reorders, it
// does not pin (see SetFolderPinned).
func (d *DB) SetPinnedFolderPositions(ctx context.Context, orderedIDs []int64) error {
	return d.setPositions(ctx,
		`UPDATE folders SET pinned_position = ? WHERE id = ? AND pinned_position > 0`,
		orderedIDs, "pinned folders")
}

// SetFolderPinned pins or unpins a folder. Pinning appends it to the end of the
// Pinned group; unpinning clears its rank and leaves the gap, which the next
// reorder closes. Pinning an already-pinned folder (or unpinning an unpinned
// one) is a no-op rather than an error.
func (d *DB) SetFolderPinned(ctx context.Context, id int64, pinned bool) error {
	if !pinned {
		res, err := d.sql.ExecContext(ctx,
			`UPDATE folders SET pinned_position = 0 WHERE id = ?`, id)
		if err != nil {
			return fmt.Errorf("storage: unpin folder %d: %w", id, err)
		}
		return requireOneRow(res, ErrFolderNotFound)
	}
	// coalesce covers the empty-group case, where max() over no rows is null.
	const query = `
UPDATE folders
SET pinned_position = (SELECT coalesce(max(pinned_position), 0) + 1 FROM folders)
WHERE id = ? AND pinned_position = 0`
	res, err := d.sql.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("storage: pin folder %d: %w", id, err)
	}
	// zero rows means it was already pinned, which is not a failure. Only a
	// missing folder is, so confirm the row exists before letting it pass.
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("storage: pin folder rows affected: %w", err)
	}
	if n == 0 {
		if _, err := d.GetFolder(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// setPositions runs one position-rewriting statement per id inside a single
// transaction. The statement takes the new position and the id, in that order.
func (d *DB) setPositions(ctx context.Context, stmt string, orderedIDs []int64, what string) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("storage: begin reorder %s: %w", what, err)
	}
	defer tx.Rollback()

	for pos, id := range orderedIDs {
		if _, err := tx.ExecContext(ctx, stmt, pos+1, id); err != nil {
			return fmt.Errorf("storage: reorder %s, id %d: %w", what, id, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("storage: commit reorder %s: %w", what, err)
	}
	return nil
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
		&parent, &attrs, &f.UIDValidity, &f.Position, &f.PinnedPosition); err != nil {
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
