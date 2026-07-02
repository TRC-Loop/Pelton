package imap

import (
	"fmt"

	"github.com/emersion/go-imap/v2"
)

// Move moves the message with the given UID to another mailbox. go-imap falls
// back to COPY + STORE \Deleted + EXPUNGE when the server lacks the MOVE
// extension, so this works on older servers too. A mailbox must be selected.
func (c *Client) Move(uid imap.UID, mailbox string) error {
	if c.raw.Mailbox() == nil {
		return fmt.Errorf("imap: no mailbox selected for move")
	}
	if _, err := c.raw.Move(imap.UIDSetNum(uid), mailbox).Wait(); err != nil {
		return fmt.Errorf("imap: move uid %d to %q: %w", uid, mailbox, err)
	}
	return nil
}
