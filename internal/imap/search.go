package imap

import (
	"fmt"
	"time"

	"github.com/emersion/go-imap/v2"
)

// SearchByMessageID returns the UIDs in the selected mailbox whose Message-ID
// header matches messageID. It backs undo-archive: after a move the message has a
// new UID in the destination, so we relocate it by its stable rfc Message-ID.
func (c *Client) SearchByMessageID(messageID string) ([]imap.UID, error) {
	if c.raw.Mailbox() == nil {
		return nil, fmt.Errorf("imap: no mailbox selected for search")
	}
	if messageID == "" {
		return nil, fmt.Errorf("imap: empty message-id")
	}
	criteria := &imap.SearchCriteria{
		Header: []imap.SearchCriteriaHeaderField{{Key: "Message-Id", Value: messageID}},
	}
	data, err := c.raw.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("imap: search message-id %q: %w", messageID, err)
	}
	return data.AllUIDs(), nil
}

// SearchSince returns the UIDs of messages in the selected mailbox whose
// internal date is on or after since. It backs the bulk offline download, which
// walks a date range and fetches whatever is not cached yet. A mailbox must be
// selected first.
func (c *Client) SearchSince(since time.Time) ([]imap.UID, error) {
	if c.raw.Mailbox() == nil {
		return nil, fmt.Errorf("imap: no mailbox selected for search")
	}
	criteria := &imap.SearchCriteria{Since: since}
	data, err := c.raw.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("imap: search since %s: %w", since.Format("2006-01-02"), err)
	}
	return data.AllUIDs(), nil
}

// FetchSizes returns the RFC822 byte size of each given uid, without fetching
// any body content. It backs the offline-download size estimate: the raw
// message size is what travels over the wire either way, whether or not
// attachments are kept afterward.
func (c *Client) FetchSizes(uids []imap.UID) (map[imap.UID]int64, error) {
	if c.raw.Mailbox() == nil {
		return nil, fmt.Errorf("imap: no mailbox selected for fetch")
	}
	if len(uids) == 0 {
		return map[imap.UID]int64{}, nil
	}
	var set imap.UIDSet
	set.AddNum(uids...)
	options := &imap.FetchOptions{UID: true, RFC822Size: true}
	buffers, err := c.raw.Fetch(set, options).Collect()
	if err != nil {
		return nil, fmt.Errorf("imap: fetch sizes: %w", err)
	}
	sizes := make(map[imap.UID]int64, len(buffers))
	for _, b := range buffers {
		sizes[b.UID] = b.RFC822Size
	}
	return sizes, nil
}
