package imap

import "fmt"

// Mailbox management (#132): CREATE, RENAME and DELETE. These change the
// server, unlike the read-only listing in folders.go, so each one is a single
// explicit call the desktop layer makes on the user's behalf and never as part
// of a sync.
//
// Paths are full mailbox names including the hierarchy, joined with the
// server's own delimiter. The delimiter varies per server (and can be absent on
// a flat one), so callers build the path from the parent they were given rather
// than assuming "/".

// CreateFolder creates a mailbox at the given full path. Creating a child under
// a parent that does not exist is up to the server: most create the missing
// levels, some refuse, so callers should create parents first when they mean
// to. It also subscribes to the new mailbox, since a server that filters LIST
// by subscription would otherwise hide it from every other client.
func (c *Client) CreateFolder(path string) error {
	if err := c.raw.Create(path, nil).Wait(); err != nil {
		return fmt.Errorf("imap: create folder %q: %w", path, err)
	}
	// best effort: the mailbox exists either way, and servers without the
	// concept of subscriptions still accept the command.
	if err := c.raw.Subscribe(path).Wait(); err != nil {
		return fmt.Errorf("imap: subscribe to new folder %q: %w", path, err)
	}
	return nil
}

// RenameFolder renames a mailbox. Per RFC 3501 the server moves the mailbox's
// children with it, so the caller has to rewrite the whole subtree locally and
// not just the one folder. Renaming INBOX is special-cased by the RFC (it moves
// the messages and leaves INBOX in place) and the desktop layer refuses it
// rather than relying on servers to agree on that behavior.
func (c *Client) RenameFolder(path, newPath string) error {
	if err := c.raw.Rename(path, newPath, nil).Wait(); err != nil {
		return fmt.Errorf("imap: rename folder %q to %q: %w", path, newPath, err)
	}
	return nil
}

// DeleteFolder deletes a mailbox and the messages in it. Servers differ on
// deleting a mailbox that still has children: some refuse, some keep the name
// as a non-selectable container. Either way the caller must not assume the
// children are gone.
func (c *Client) DeleteFolder(path string) error {
	if err := c.raw.Delete(path).Wait(); err != nil {
		return fmt.Errorf("imap: delete folder %q: %w", path, err)
	}
	return nil
}
