package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrSignatureNotFound is returned when a signature id has no row.
var ErrSignatureNotFound = errors.New("storage: signature not found")

// Signature is a reusable header or footer block. Kind is "header" or "footer";
// Format is "markdown" or "html".
type Signature struct {
	ID        int64
	Name      string
	Kind      string
	Format    string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ListSignatures returns every signature ordered by kind then name.
func (d *DB) ListSignatures(ctx context.Context) ([]Signature, error) {
	const query = `
SELECT id, name, kind, format, content, created_at, updated_at
FROM signatures ORDER BY kind, name`
	rows, err := d.sql.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("storage: list signatures: %w", err)
	}
	defer rows.Close()

	var out []Signature
	for rows.Next() {
		s, err := scanSignature(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan signature: %w", err)
		}
		out = append(out, *s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: iterate signatures: %w", err)
	}
	return out, nil
}

// CreateSignature inserts a signature and returns its new id.
func (d *DB) CreateSignature(ctx context.Context, s *Signature) (int64, error) {
	now := time.Now().UTC()
	const query = `
INSERT INTO signatures (name, kind, format, content, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)`
	res, err := d.sql.ExecContext(ctx, query,
		s.Name, s.Kind, s.Format, s.Content, formatTime(now), formatTime(now))
	if err != nil {
		return 0, fmt.Errorf("storage: insert signature: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("storage: signature insert id: %w", err)
	}
	s.ID = id
	s.CreatedAt = now
	s.UpdatedAt = now
	return id, nil
}

// UpdateSignature updates the mutable fields of an existing signature.
func (d *DB) UpdateSignature(ctx context.Context, s *Signature) error {
	now := time.Now().UTC()
	const query = `
UPDATE signatures SET name = ?, kind = ?, format = ?, content = ?, updated_at = ?
WHERE id = ?`
	res, err := d.sql.ExecContext(ctx, query, s.Name, s.Kind, s.Format, s.Content, formatTime(now), s.ID)
	if err != nil {
		return fmt.Errorf("storage: update signature %d: %w", s.ID, err)
	}
	s.UpdatedAt = now
	return requireOneRow(res, ErrSignatureNotFound)
}

// DeleteSignature removes a signature. account_signatures references are set to
// null by the foreign key, so deleting never strands an account.
func (d *DB) DeleteSignature(ctx context.Context, id int64) error {
	res, err := d.sql.ExecContext(ctx, `DELETE FROM signatures WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("storage: delete signature %d: %w", id, err)
	}
	return requireOneRow(res, ErrSignatureNotFound)
}

// AccountSignatures is one account's default header/footer assignment. A zero id
// means no default is assigned for that slot.
type AccountSignatures struct {
	HeaderID int64
	FooterID int64
}

// GetAccountSignatures returns an account's default header/footer ids, with zero
// for any slot that has no row or no assignment.
func (d *DB) GetAccountSignatures(ctx context.Context, accountID int64) (AccountSignatures, error) {
	const query = `SELECT header_id, footer_id FROM account_signatures WHERE account_id = ?`
	var header, footer sql.NullInt64
	err := d.sql.QueryRowContext(ctx, query, accountID).Scan(&header, &footer)
	if errors.Is(err, sql.ErrNoRows) {
		return AccountSignatures{}, nil
	}
	if err != nil {
		return AccountSignatures{}, fmt.Errorf("storage: get account signatures %d: %w", accountID, err)
	}
	return AccountSignatures{HeaderID: header.Int64, FooterID: footer.Int64}, nil
}

// SetAccountSignatures upserts an account's default header/footer assignment. A
// zero id clears that slot.
func (d *DB) SetAccountSignatures(ctx context.Context, accountID int64, as AccountSignatures) error {
	const query = `
INSERT INTO account_signatures (account_id, header_id, footer_id)
VALUES (?, ?, ?)
ON CONFLICT(account_id) DO UPDATE SET header_id = excluded.header_id, footer_id = excluded.footer_id`
	_, err := d.sql.ExecContext(ctx, query, accountID, zeroAsNull(as.HeaderID), zeroAsNull(as.FooterID))
	if err != nil {
		return fmt.Errorf("storage: set account signatures %d: %w", accountID, err)
	}
	return nil
}

// zeroAsNull maps a zero id to SQL NULL so an unassigned slot stores as null.
func zeroAsNull(id int64) any {
	if id == 0 {
		return nil
	}
	return id
}

func scanSignature(row rowScanner) (*Signature, error) {
	var (
		s                Signature
		created, updated string
	)
	if err := row.Scan(&s.ID, &s.Name, &s.Kind, &s.Format, &s.Content, &created, &updated); err != nil {
		return nil, err
	}
	ct, err := parseTime(created)
	if err != nil {
		return nil, err
	}
	ut, err := parseTime(updated)
	if err != nil {
		return nil, err
	}
	s.CreatedAt = ct
	s.UpdatedAt = ut
	return &s, nil
}
