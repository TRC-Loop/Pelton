package imap

import (
	"fmt"

	"github.com/emersion/go-imap/v2"
)

// MarkDeleted sets the \Deleted flag on the given uids in a single STORE.
// nothing is actually removed until Expunge runs.
func (c *Client) MarkDeleted(uids ...imap.UID) error {
	if len(uids) == 0 {
		return nil
	}
	storeFlags := &imap.StoreFlags{
		Op:     imap.StoreFlagsAdd,
		Flags:  []imap.Flag{imap.FlagDeleted},
		Silent: true,
	}
	if err := c.raw.Store(imap.UIDSetNum(uids...), storeFlags, nil).Close(); err != nil {
		return fmt.Errorf("imap: mark deleted: %w", err)
	}
	return nil
}

// Expunge permanently removes \Deleted messages. when the server advertises
// UIDPLUS and uids are given it issues UID EXPUNGE, removing only those uids so
// a \Deleted message another client left behind is untouched. otherwise a plain
// EXPUNGE removes every \Deleted message in the mailbox.
//
// gmail diverges: \Deleted plus EXPUNGE inside an ordinary label only removes
// that label, real deletion happens in [Gmail]/Trash or All Mail. a later
// version may move to Trash on gmail instead, see push.go.
func (c *Client) Expunge(uids ...imap.UID) error {
	if len(uids) > 0 && c.raw.Caps().Has(imap.CapUIDPlus) {
		if _, err := c.raw.UIDExpunge(imap.UIDSetNum(uids...)).Collect(); err != nil {
			return fmt.Errorf("imap: uid expunge: %w", err)
		}
		return nil
	}
	if _, err := c.raw.Expunge().Collect(); err != nil {
		return fmt.Errorf("imap: expunge: %w", err)
	}
	return nil
}
