package imap

import (
	"context"
	"fmt"

	"github.com/emersion/go-imap/v2"
)

// IdleUntil parks the connection in IMAP IDLE and blocks until the server
// pushes a mailbox update, ctx is cancelled, or the connection drops. It always
// stops IDLE before returning, so on return the connection is free for the
// caller to issue commands (a FETCH resync, say). It reports whether an update
// arrived: true means new server activity to sync, false means ctx ended.
//
// This replaces the old pattern of holding IDLE open for the whole session and
// fetching on a separate goroutine: go-imap forbids sending any command while
// IDLE runs, so that fetch blocked until the 28-minute idle restart, which is
// what made new mail take minutes to appear.
func (c *Client) IdleUntil(ctx context.Context) (bool, error) {
	if !c.SupportsIdle() {
		return false, fmt.Errorf("imap: server does not advertise the IDLE capability")
	}
	if c.raw.State() != imap.ConnStateSelected {
		return false, fmt.Errorf("imap: a mailbox must be selected before idling")
	}

	cmd, err := c.raw.Idle()
	if err != nil {
		return false, fmt.Errorf("imap: start idle: %w", err)
	}

	// cmd.Wait blocks until IDLE terminates on its own, which is how a dropped
	// connection surfaces; watch it so a dead link is reported instead of
	// hanging until ctx is cancelled.
	idleDone := make(chan error, 1)
	go func() { idleDone <- cmd.Wait() }()

	select {
	case <-ctx.Done():
		return false, c.stopIdle(cmd, idleDone)
	case _, ok := <-c.updates:
		if !ok {
			return false, c.stopIdle(cmd, idleDone)
		}
		if err := c.stopIdle(cmd, idleDone); err != nil {
			return true, err
		}
		c.drainUpdates()
		return true, nil
	case err := <-idleDone:
		if err == nil {
			err = fmt.Errorf("imap: idle ended unexpectedly")
		}
		return false, fmt.Errorf("imap: idle terminated: %w", err)
	}
}

// stopIdle closes the IDLE command and reaps the Wait goroutine started in
// IdleUntil, returning the first error from either.
func (c *Client) stopIdle(cmd idleCloser, idleDone <-chan error) error {
	closeErr := cmd.Close()
	waitErr := <-idleDone
	if closeErr != nil {
		return fmt.Errorf("imap: stop idle: %w", closeErr)
	}
	if waitErr != nil {
		return fmt.Errorf("imap: idle terminated: %w", waitErr)
	}
	return nil
}

// idleCloser is the slice of *imapclient.IdleCommand IdleUntil needs, named so
// stopIdle stays testable without a live connection.
type idleCloser interface {
	Close() error
}

// drainUpdates clears any updates that coalesced while IDLE was stopping, so the
// next IdleUntil starts clean and a single resync covers them all.
func (c *Client) drainUpdates() {
	for {
		select {
		case <-c.updates:
		default:
			return
		}
	}
}
