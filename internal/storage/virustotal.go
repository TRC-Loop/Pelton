package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// VerdictKind distinguishes the two things that get scanned, since a url and a
// file digest could otherwise collide in one key space.
type VerdictKind string

const (
	// VerdictKindURL is a link found in a message body.
	VerdictKindURL VerdictKind = "url"
	// VerdictKindFile is an attachment, keyed by its sha-256.
	VerdictKindFile VerdictKind = "file"
)

// ErrVerdictNotCached is returned when nothing usable is cached for a target,
// either because it was never looked up or because the entry aged out.
var ErrVerdictNotCached = errors.New("storage: no cached verdict")

// Verdict is one cached scan result. Status, the engine counts and Permalink
// are stored exactly as the scanner reported them.
type Verdict struct {
	Status     string
	Malicious  int
	Suspicious int
	Total      int
	Permalink  string
	CheckedAt  time.Time
}

// CachedVerdict returns the stored verdict for a target if it was recorded
// within maxAge, or ErrVerdictNotCached. A zero or negative maxAge treats every
// entry as expired, which is how a forced rescan skips the cache.
func (d *DB) CachedVerdict(ctx context.Context, kind VerdictKind, target string, maxAge time.Duration) (*Verdict, error) {
	if maxAge <= 0 {
		return nil, ErrVerdictNotCached
	}

	const query = `
SELECT status, malicious, suspicious, total, permalink, checked_at
FROM virustotal_verdicts WHERE kind = ? AND target = ?`

	var (
		v       Verdict
		checked string
	)
	err := d.sql.QueryRowContext(ctx, query, string(kind), target).
		Scan(&v.Status, &v.Malicious, &v.Suspicious, &v.Total, &v.Permalink, &checked)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVerdictNotCached
	}
	if err != nil {
		return nil, fmt.Errorf("storage: get cached verdict: %w", err)
	}

	t, err := parseTime(checked)
	if err != nil {
		return nil, fmt.Errorf("storage: cached verdict time: %w", err)
	}
	if time.Since(t) > maxAge {
		return nil, ErrVerdictNotCached
	}
	v.CheckedAt = t
	return &v, nil
}

// CacheVerdict records a scan result, replacing any previous one for the same
// target. CheckedAt is set to now when the caller left it zero.
func (d *DB) CacheVerdict(ctx context.Context, kind VerdictKind, target string, v Verdict) error {
	checked := v.CheckedAt
	if checked.IsZero() {
		checked = time.Now().UTC()
	}

	const query = `
INSERT INTO virustotal_verdicts (kind, target, status, malicious, suspicious, total, permalink, checked_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(kind, target) DO UPDATE SET
    status = excluded.status,
    malicious = excluded.malicious,
    suspicious = excluded.suspicious,
    total = excluded.total,
    permalink = excluded.permalink,
    checked_at = excluded.checked_at`
	_, err := d.sql.ExecContext(ctx, query, string(kind), target,
		v.Status, v.Malicious, v.Suspicious, v.Total, v.Permalink, formatTime(checked))
	if err != nil {
		return fmt.Errorf("storage: cache verdict: %w", err)
	}
	return nil
}

// ClearVerdicts deletes every cached verdict. Turning the integration off calls
// it, so the record of which links and files were scanned does not outlive the
// feature that created it.
func (d *DB) ClearVerdicts(ctx context.Context) error {
	if _, err := d.sql.ExecContext(ctx, `DELETE FROM virustotal_verdicts`); err != nil {
		return fmt.Errorf("storage: clear verdicts: %w", err)
	}
	return nil
}
