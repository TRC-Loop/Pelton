package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// AgentAction is one write an agent made, for the log the user can read.
type AgentAction struct {
	ID        int64
	Tool      string
	MessageID int64
	Summary   string
	Error     string
	CreatedAt time.Time
}

// AgentProposal is a message an agent wants sent, waiting on the user.
type AgentProposal struct {
	ID        int64
	AccountID int64
	To        string
	Cc        string
	Bcc       string
	Subject   string
	Body      string
	CreatedAt time.Time
}

// ErrProposalNotFound means the proposal is gone, most likely already answered.
var ErrProposalNotFound = errors.New("storage: agent proposal not found")

// RecordAgentAction appends to the log.
func (d *DB) RecordAgentAction(ctx context.Context, a AgentAction) error {
	const query = `
INSERT INTO agent_actions (tool, message_id, summary, error, created_at)
VALUES (?, ?, ?, ?, ?)`
	_, err := d.sql.ExecContext(ctx, query, a.Tool, a.MessageID, a.Summary, a.Error,
		formatTime(nowOr(a.CreatedAt)))
	if err != nil {
		return fmt.Errorf("storage: record agent action %s: %w", a.Tool, err)
	}
	return nil
}

// ListAgentActions returns the most recent actions, newest first.
func (d *DB) ListAgentActions(ctx context.Context, limit int) ([]AgentAction, error) {
	const query = `
SELECT id, tool, message_id, summary, error, created_at
FROM agent_actions ORDER BY id DESC LIMIT ?`
	rows, err := d.sql.QueryContext(ctx, query, normalizeLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("storage: list agent actions: %w", err)
	}
	defer rows.Close()

	var out []AgentAction
	for rows.Next() {
		var (
			a       AgentAction
			created string
		)
		if err := rows.Scan(&a.ID, &a.Tool, &a.MessageID, &a.Summary, &a.Error, &created); err != nil {
			return nil, fmt.Errorf("storage: scan agent action: %w", err)
		}
		a.CreatedAt, _ = parseTime(created)
		out = append(out, a)
	}
	return out, rows.Err()
}

// ClearAgentActions empties the log.
func (d *DB) ClearAgentActions(ctx context.Context) error {
	if _, err := d.sql.ExecContext(ctx, `DELETE FROM agent_actions`); err != nil {
		return fmt.Errorf("storage: clear agent actions: %w", err)
	}
	return nil
}

// CreateAgentProposal queues a message for the user to approve and returns its
// id.
func (d *DB) CreateAgentProposal(ctx context.Context, p AgentProposal) (int64, error) {
	const query = `
INSERT INTO agent_proposals (account_id, to_addrs, cc_addrs, bcc_addrs, subject, body, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := d.sql.ExecContext(ctx, query, p.AccountID, p.To, p.Cc, p.Bcc,
		p.Subject, p.Body, formatTime(nowOr(p.CreatedAt)))
	if err != nil {
		return 0, fmt.Errorf("storage: create agent proposal: %w", err)
	}
	return res.LastInsertId()
}

// ListAgentProposals returns everything waiting on the user, oldest first.
func (d *DB) ListAgentProposals(ctx context.Context) ([]AgentProposal, error) {
	const query = selectProposalColumns + ` FROM agent_proposals ORDER BY id`
	rows, err := d.sql.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("storage: list agent proposals: %w", err)
	}
	defer rows.Close()

	var out []AgentProposal
	for rows.Next() {
		p, err := scanProposal(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// GetAgentProposal returns one proposal, or ErrProposalNotFound.
func (d *DB) GetAgentProposal(ctx context.Context, id int64) (*AgentProposal, error) {
	const query = selectProposalColumns + ` FROM agent_proposals WHERE id = ?`
	p, err := scanProposal(d.sql.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProposalNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("storage: get agent proposal %d: %w", id, err)
	}
	return p, nil
}

// DeleteAgentProposal removes a proposal, once it has been sent or discarded.
func (d *DB) DeleteAgentProposal(ctx context.Context, id int64) error {
	res, err := d.sql.ExecContext(ctx, `DELETE FROM agent_proposals WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("storage: delete agent proposal %d: %w", id, err)
	}
	return requireOneRow(res, ErrProposalNotFound)
}

const selectProposalColumns = `
SELECT id, account_id, to_addrs, cc_addrs, bcc_addrs, subject, body, created_at`

func scanProposal(row rowScanner) (*AgentProposal, error) {
	var (
		p       AgentProposal
		created string
	)
	if err := row.Scan(&p.ID, &p.AccountID, &p.To, &p.Cc, &p.Bcc, &p.Subject,
		&p.Body, &created); err != nil {
		return nil, err
	}
	p.CreatedAt, _ = parseTime(created)
	return &p, nil
}

// nowOr fills in the current time for a caller that left it zero, so every row
// carries one.
func nowOr(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}
