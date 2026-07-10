package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// MessageQuery selects a page of messages. It supports one or more folders (the
// unified views pass several folder ids at once) and an optional flag filter so
// the unified Flagged view can ask for only flagged rows. A zero RequireFlags
// means no flag filter. Limit and Offset drive list pagination; a non positive
// Limit means no cap.
type MessageQuery struct {
	FolderIDs    []int64
	RequireFlags Flag
	Limit        int
	Offset       int
}

// QueryMessages returns the page of messages matching q, newest first. It is the
// single read path the ui list uses for both per folder and unified views.
func (d *DB) QueryMessages(ctx context.Context, q MessageQuery) ([]Message, error) {
	if len(q.FolderIDs) == 0 {
		return nil, nil
	}

	where, args := messageWhere(q)
	query := selectMessageColumns + `
FROM messages
WHERE ` + where + `
ORDER BY date DESC, uid DESC
LIMIT ? OFFSET ?`
	args = append(args, normalizeLimit(q.Limit), q.Offset)

	rows, err := d.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: query messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("storage: scan message: %w", err)
		}
		messages = append(messages, *m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: iterate messages: %w", err)
	}
	return messages, nil
}

// CountMessages returns how many messages match q ignoring its limit and offset,
// so the ui can show totals and decide whether more pages exist.
func (d *DB) CountMessages(ctx context.Context, q MessageQuery) (int, error) {
	if len(q.FolderIDs) == 0 {
		return 0, nil
	}
	where, args := messageWhere(q)
	var n int
	err := d.sql.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM messages WHERE `+where, args...).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("storage: count messages: %w", err)
	}
	return n, nil
}

// FolderCounts returns the total and unread message counts for a single folder.
// Unread means the \Seen flag bit is not set.
func (d *DB) FolderCounts(ctx context.Context, folderID int64) (total, unread int, err error) {
	const query = `
SELECT COUNT(*), COALESCE(SUM(CASE WHEN flags & ? = 0 THEN 1 ELSE 0 END), 0)
FROM messages WHERE folder_id = ?`
	if err := d.sql.QueryRowContext(ctx, query, uint8(FlagSeen), folderID).Scan(&total, &unread); err != nil {
		return 0, 0, fmt.Errorf("storage: folder counts %d: %w", folderID, err)
	}
	return total, unread, nil
}

// UnreadCount returns the number of unread messages across the given folders,
// used for unified view badges.
func (d *DB) UnreadCount(ctx context.Context, folderIDs []int64) (int, error) {
	if len(folderIDs) == 0 {
		return 0, nil
	}
	placeholders, args := inClause(folderIDs)
	args = append([]any{uint8(FlagSeen)}, args...)
	query := `SELECT COUNT(*) FROM messages WHERE flags & ? = 0 AND folder_id IN (` + placeholders + `)`
	var n int
	if err := d.sql.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("storage: unread count: %w", err)
	}
	return n, nil
}

// LatestMessageFrom returns the most recent cached message whose sender matches
// value: an exact from-address when matchDomain is false, or any sender in the
// domain when it is true. Returns nil (no error) when nothing matches, so the
// image allowlist ui can show an example message for a trusted sender/domain.
func (d *DB) LatestMessageFrom(ctx context.Context, value string, matchDomain bool) (*Message, error) {
	cond := "LOWER(from_address) = ?"
	arg := strings.ToLower(value)
	if matchDomain {
		cond = "LOWER(from_address) LIKE ?"
		arg = "%@" + strings.ToLower(value)
	}
	query := selectMessageColumns + `
FROM messages
WHERE ` + cond + `
ORDER BY date DESC, uid DESC
LIMIT 1`
	m, err := scanMessage(d.sql.QueryRowContext(ctx, query, arg))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("storage: latest message from %q: %w", value, err)
	}
	return m, nil
}

// messageWhere builds the shared WHERE clause and its args for QueryMessages and
// CountMessages so the two never drift apart.
func messageWhere(q MessageQuery) (string, []any) {
	placeholders, args := inClause(q.FolderIDs)
	// pending_delete rows are awaiting server expunge; hide them from the list so
	// a local delete disappears immediately and reappears nowhere. snooze_hidden
	// rows are snoozed-and-hidden; they stay out of the list until the snooze fires
	// (the poller flips the bit back).
	where := "pending_delete = 0 AND snooze_hidden = 0 AND folder_id IN (" + placeholders + ")"
	if q.RequireFlags != 0 {
		// every requested flag bit must be set: (flags & mask) = mask.
		where += " AND (flags & ?) = ?"
		args = append(args, uint8(q.RequireFlags), uint8(q.RequireFlags))
	}
	return where, args
}

// inClause renders n "?" placeholders and the matching args slice for an IN list.
func inClause(ids []int64) (string, []any) {
	marks := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		marks[i] = "?"
		args[i] = id
	}
	return strings.Join(marks, ", "), args
}
