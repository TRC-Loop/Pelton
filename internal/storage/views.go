package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrViewNotFound is returned when a view id has no row.
var ErrViewNotFound = errors.New("storage: view not found")

// View is a user-defined saved search ("preset search"). The query fields mirror
// search.Query; the scope fields narrow the result further. AccountID is zero for
// "all accounts". Position drives sidebar ordering.
type View struct {
	ID    int64
	Name  string
	Icon  string
	Color string

	QueryText    string
	QueryFrom    string
	QueryTo      string
	QuerySubject string
	WithinDays   int

	UnreadOnly    bool
	FlaggedOnly   bool
	HasAttachment bool
	AccountID     int64

	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

const selectViewColumns = `
SELECT id, name, icon, color,
       query_text, query_from, query_to, query_subject, within_days,
       unread_only, flagged_only, has_attachment, account_id,
       position, created_at, updated_at
FROM views`

// ListViews returns every view ordered by position then name.
func (d *DB) ListViews(ctx context.Context) ([]View, error) {
	rows, err := d.sql.QueryContext(ctx, selectViewColumns+` ORDER BY position, name`)
	if err != nil {
		return nil, fmt.Errorf("storage: list views: %w", err)
	}
	defer rows.Close()

	var out []View
	for rows.Next() {
		v, err := scanView(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan view: %w", err)
		}
		out = append(out, *v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: iterate views: %w", err)
	}
	return out, nil
}

// GetView returns one view by id.
func (d *DB) GetView(ctx context.Context, id int64) (*View, error) {
	v, err := scanView(d.sql.QueryRowContext(ctx, selectViewColumns+` WHERE id = ?`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrViewNotFound
		}
		return nil, fmt.Errorf("storage: get view %d: %w", id, err)
	}
	return v, nil
}

// CreateView inserts a view and returns its new id. When Position is zero it is
// placed at the end so new views append to the list.
func (d *DB) CreateView(ctx context.Context, v *View) (int64, error) {
	now := time.Now().UTC()
	if v.Position == 0 {
		var maxPos sql.NullInt64
		if err := d.sql.QueryRowContext(ctx, `SELECT MAX(position) FROM views`).Scan(&maxPos); err != nil {
			return 0, fmt.Errorf("storage: view max position: %w", err)
		}
		v.Position = int(maxPos.Int64) + 1
	}
	const query = `
INSERT INTO views (name, icon, color, query_text, query_from, query_to, query_subject,
                   within_days, unread_only, flagged_only, has_attachment, account_id,
                   position, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := d.sql.ExecContext(ctx, query,
		v.Name, v.Icon, v.Color, v.QueryText, v.QueryFrom, v.QueryTo, v.QuerySubject,
		v.WithinDays, boolToInt(v.UnreadOnly), boolToInt(v.FlaggedOnly), boolToInt(v.HasAttachment),
		zeroAsNull(v.AccountID), v.Position, formatTime(now), formatTime(now))
	if err != nil {
		return 0, fmt.Errorf("storage: insert view: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("storage: view insert id: %w", err)
	}
	v.ID = id
	v.CreatedAt = now
	v.UpdatedAt = now
	return id, nil
}

// UpdateView updates the mutable fields of an existing view.
func (d *DB) UpdateView(ctx context.Context, v *View) error {
	now := time.Now().UTC()
	const query = `
UPDATE views SET name = ?, icon = ?, color = ?,
                 query_text = ?, query_from = ?, query_to = ?, query_subject = ?,
                 within_days = ?, unread_only = ?, flagged_only = ?, has_attachment = ?,
                 account_id = ?, position = ?, updated_at = ?
WHERE id = ?`
	res, err := d.sql.ExecContext(ctx, query,
		v.Name, v.Icon, v.Color, v.QueryText, v.QueryFrom, v.QueryTo, v.QuerySubject,
		v.WithinDays, boolToInt(v.UnreadOnly), boolToInt(v.FlaggedOnly), boolToInt(v.HasAttachment),
		zeroAsNull(v.AccountID), v.Position, formatTime(now), v.ID)
	if err != nil {
		return fmt.Errorf("storage: update view %d: %w", v.ID, err)
	}
	v.UpdatedAt = now
	return requireOneRow(res, ErrViewNotFound)
}

// DeleteView removes a view.
func (d *DB) DeleteView(ctx context.Context, id int64) error {
	res, err := d.sql.ExecContext(ctx, `DELETE FROM views WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("storage: delete view %d: %w", id, err)
	}
	return requireOneRow(res, ErrViewNotFound)
}

// SetViewPositions rewrites the ordering of views in one transaction, so a drag
// reorder persists atomically. ids not present keep their current position.
func (d *DB) SetViewPositions(ctx context.Context, orderedIDs []int64) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("storage: begin reorder views: %w", err)
	}
	defer tx.Rollback()

	now := formatTime(time.Now().UTC())
	for pos, id := range orderedIDs {
		if _, err := tx.ExecContext(ctx,
			`UPDATE views SET position = ?, updated_at = ? WHERE id = ?`, pos+1, now, id); err != nil {
			return fmt.Errorf("storage: reorder view %d: %w", id, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("storage: commit reorder views: %w", err)
	}
	return nil
}

func scanView(row rowScanner) (*View, error) {
	var (
		v                       View
		accountID               sql.NullInt64
		unread, flagged, hasAtt int
		created, updated        string
	)
	if err := row.Scan(&v.ID, &v.Name, &v.Icon, &v.Color,
		&v.QueryText, &v.QueryFrom, &v.QueryTo, &v.QuerySubject, &v.WithinDays,
		&unread, &flagged, &hasAtt, &accountID,
		&v.Position, &created, &updated); err != nil {
		return nil, err
	}
	v.UnreadOnly = unread != 0
	v.FlaggedOnly = flagged != 0
	v.HasAttachment = hasAtt != 0
	v.AccountID = accountID.Int64
	ct, err := parseTime(created)
	if err != nil {
		return nil, err
	}
	ut, err := parseTime(updated)
	if err != nil {
		return nil, err
	}
	v.CreatedAt = ct
	v.UpdatedAt = ut
	return &v, nil
}
