package storage

import (
	"context"
	"fmt"
	"time"
)

// AccountSyncState is how an account's last sync went. LastOK is kept across a
// failure, so the ui can say how long a mailbox has been broken rather than
// only that it is.
type AccountSyncState struct {
	AccountID int64
	LastOK    time.Time
	FailedAt  time.Time
	// Reason is a coarse class the ui turns into a sentence: "auth", "network",
	// "credentials" or "other". Detail is the underlying error text, shown to
	// the reader who wants the server's own words.
	Reason string
	Detail string
}

// Failing reports whether this account's last sync attempt failed.
func (s AccountSyncState) Failing() bool { return !s.FailedAt.IsZero() }

// RecordSyncOK marks an account as having synced through to the end, clearing
// any failure it was carrying.
func (d *DB) RecordSyncOK(ctx context.Context, accountID int64) error {
	const query = `
INSERT INTO account_sync_state (account_id, last_ok_at, failed_at, reason, detail)
VALUES (?, ?, '', '', '')
ON CONFLICT(account_id) DO UPDATE SET last_ok_at = excluded.last_ok_at, failed_at = '', reason = '', detail = ''`
	if _, err := d.sql.ExecContext(ctx, query, accountID, formatTime(time.Now().UTC())); err != nil {
		return fmt.Errorf("storage: record sync ok for account %d: %w", accountID, err)
	}
	return nil
}

// RecordSyncFailure records that an account's sync failed, leaving its last
// successful time alone.
func (d *DB) RecordSyncFailure(ctx context.Context, accountID int64, reason, detail string) error {
	const query = `
INSERT INTO account_sync_state (account_id, last_ok_at, failed_at, reason, detail)
VALUES (?, '', ?, ?, ?)
ON CONFLICT(account_id) DO UPDATE SET failed_at = excluded.failed_at, reason = excluded.reason, detail = excluded.detail`
	if _, err := d.sql.ExecContext(ctx, query, accountID, formatTime(time.Now().UTC()), reason, detail); err != nil {
		return fmt.Errorf("storage: record sync failure for account %d: %w", accountID, err)
	}
	return nil
}

// AccountSyncStates returns the recorded state of every account that has one.
// Accounts that have never been synced are simply absent.
func (d *DB) AccountSyncStates(ctx context.Context) ([]AccountSyncState, error) {
	const query = `SELECT account_id, last_ok_at, failed_at, reason, detail FROM account_sync_state`
	rows, err := d.sql.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("storage: list account sync state: %w", err)
	}
	defer rows.Close()

	out := make([]AccountSyncState, 0, 4)
	for rows.Next() {
		var s AccountSyncState
		var lastOK, failedAt string
		if err := rows.Scan(&s.AccountID, &lastOK, &failedAt, &s.Reason, &s.Detail); err != nil {
			return nil, fmt.Errorf("storage: scan account sync state: %w", err)
		}
		if s.LastOK, err = parseTime(lastOK); err != nil {
			return nil, err
		}
		if s.FailedAt, err = parseTime(failedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
