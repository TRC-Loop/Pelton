package imap

import (
	"errors"
	"fmt"

	"github.com/emersion/go-imap/v2"
)

// DeleteMessages permanently removes the given uids from the selected mailbox,
// and only those. It marks them \Deleted and expunges.
//
// The care is in the expunge. UID EXPUNGE (RFC 4315) removes exactly the uids
// asked for, but it needs UIDPLUS, and a plain EXPUNGE removes every message in
// the mailbox that carries \Deleted, including ones another client marked and
// has not expunged yet. Deleting one message in Pelton must never take a mail
// somebody's other client had merely flagged.
//
// So without UIDPLUS the foreign \Deleted marks are lifted for the length of
// the expunge and put back afterwards. It is the only way to scope a plain
// EXPUNGE, and the failure mode is a flag another client has to set again,
// rather than mail nobody can get back. When the mailbox holds no foreign marks
// at all, which is the ordinary case, a plain EXPUNGE is already exact and
// nothing is touched.
//
// gmail diverges from all of this: \Deleted plus EXPUNGE inside an ordinary
// label only removes that label, real deletion happens in [Gmail]/Trash or All
// Mail. See the note in push.go.
func (c *Client) DeleteMessages(uids ...imap.UID) error {
	if len(uids) == 0 {
		return nil
	}
	if c.raw.Caps().Has(imap.CapUIDPlus) {
		if err := c.markDeleted(uids); err != nil {
			return err
		}
		if _, err := c.raw.UIDExpunge(imap.UIDSetNum(uids...)).Collect(); err != nil {
			return fmt.Errorf("imap: uid expunge: %w", err)
		}
		return nil
	}

	// the search has to run before our own marks go on, or ours would come back
	// as foreign. a failure here is not a reason to expunge blindly: without
	// knowing what else is flagged there is no safe unscoped EXPUNGE.
	existing, err := c.searchDeleted()
	if err != nil {
		return fmt.Errorf("imap: refusing to expunge without UIDPLUS, cannot tell what else is marked deleted: %w", err)
	}
	foreign := foreignDeleted(existing, uids)

	if len(foreign) == 0 {
		if err := c.markDeleted(uids); err != nil {
			return err
		}
		return c.expungeAll()
	}

	if err := c.storeMany(foreign, imap.StoreFlagsDel, imap.FlagDeleted); err != nil {
		return fmt.Errorf("imap: cannot protect another client's deletions, not expunging: %w", err)
	}

	markErr := c.markDeleted(uids)
	var expungeErr error
	if markErr == nil {
		expungeErr = c.expungeAll()
	}

	// the marks go back on whatever happened above, and a failure to restore
	// them is reported rather than swallowed: the other client's pending
	// deletions are gone from its point of view until it marks them again.
	var restoreErr error
	if err := c.storeMany(foreign, imap.StoreFlagsAdd, imap.FlagDeleted); err != nil {
		restoreErr = fmt.Errorf("imap: could not restore %d \\Deleted mark(s) set by another client: %w", len(foreign), err)
	}
	return errors.Join(markErr, expungeErr, restoreErr)
}

// markDeleted sets \Deleted on the given uids in a single STORE. Nothing is
// removed until an expunge runs.
func (c *Client) markDeleted(uids []imap.UID) error {
	if err := c.storeMany(uids, imap.StoreFlagsAdd, imap.FlagDeleted); err != nil {
		return fmt.Errorf("imap: mark deleted: %w", err)
	}
	return nil
}

// expungeAll issues a plain EXPUNGE, which removes every \Deleted message in
// the mailbox. Only call it once the mailbox is known to hold no \Deleted
// message other than the ones being deleted.
func (c *Client) expungeAll() error {
	if _, err := c.raw.Expunge().Collect(); err != nil {
		return fmt.Errorf("imap: expunge: %w", err)
	}
	return nil
}

// searchDeleted returns the uids currently carrying \Deleted in the selected
// mailbox.
func (c *Client) searchDeleted() ([]imap.UID, error) {
	if c.raw.Mailbox() == nil {
		return nil, fmt.Errorf("imap: no mailbox selected")
	}
	criteria := &imap.SearchCriteria{Flag: []imap.Flag{imap.FlagDeleted}}
	data, err := c.raw.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("imap: search deleted: %w", err)
	}
	return data.AllUIDs(), nil
}

// storeMany issues one UID STORE for a set of uids. An empty set is a no-op
// rather than a STORE with no target.
func (c *Client) storeMany(uids []imap.UID, op imap.StoreFlagsOp, flags ...imap.Flag) error {
	if len(uids) == 0 {
		return nil
	}
	storeFlags := &imap.StoreFlags{Op: op, Flags: flags, Silent: true}
	if err := c.raw.Store(imap.UIDSetNum(uids...), storeFlags, nil).Close(); err != nil {
		return fmt.Errorf("imap: store flags on %d uid(s): %w", len(uids), err)
	}
	return nil
}

// foreignDeleted returns the uids that are flagged \Deleted but are not ours to
// delete: the messages a plain EXPUNGE would take with it.
func foreignDeleted(existing, ours []imap.UID) []imap.UID {
	if len(existing) == 0 {
		return nil
	}
	mine := make(map[imap.UID]struct{}, len(ours))
	for _, uid := range ours {
		mine[uid] = struct{}{}
	}
	var foreign []imap.UID
	for _, uid := range existing {
		if _, ok := mine[uid]; !ok {
			foreign = append(foreign, uid)
		}
	}
	return foreign
}
