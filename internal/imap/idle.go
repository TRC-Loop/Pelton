package imap

import (
	"context"
	"fmt"

	"github.com/emersion/go-imap/v2"
)

// Idle blocks in IMAP IDLE until ctx is cancelled. Updates arrive on Updates().
func (c *Client) Idle(ctx context.Context) error {
	if !c.SupportsIdle() {
		return fmt.Errorf("imap: server does not advertise the IDLE capability")
	}
	if c.raw.State() != imap.ConnStateSelected {
		return fmt.Errorf("imap: a mailbox must be selected before idling")
	}

	cmd, err := c.raw.Idle()
	if err != nil {
		return fmt.Errorf("imap: start idle: %w", err)
	}

	<-ctx.Done()

	if err := cmd.Close(); err != nil {
		return fmt.Errorf("imap: stop idle: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("imap: idle terminated: %w", err)
	}
	return nil
}
