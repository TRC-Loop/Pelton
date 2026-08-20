package imap

import (
	"slices"
	"testing"

	"github.com/emersion/go-imap/v2"
)

// TestMoveWithoutMoveExtensionLeavesForeignDeletedMail is the move-path twin of
// TestDeleteMessagesLeavesAnotherClientsDeletedMail. Without the MOVE
// extension a move is a copy plus a delete, and go-imap's own fallback ends in
// a plain EXPUNGE, which would take a message another client had only flagged.
func TestMoveWithoutMoveExtensionLeavesForeignDeletedMail(t *testing.T) {
	c, wire := newTestServer(t, testOptions{subjects: []string{"theirs", "ours", "keep"}})
	if err := createMailbox(c, "Archive"); err != nil {
		t.Fatalf("create Archive: %v", err)
	}

	theirs := uidOf(t, c, "theirs")
	ours := uidOf(t, c, "ours")
	if err := c.storeMany([]imap.UID{theirs}, imap.StoreFlagsAdd, imap.FlagDeleted); err != nil {
		t.Fatalf("mark theirs: %v", err)
	}
	wire.reset()

	if err := c.MoveMessages([]imap.UID{ours}, "Archive"); err != nil {
		t.Fatalf("MoveMessages: %v", err)
	}

	if got := mailboxSubjects(t, c, "INBOX"); !slices.Equal(got, []string{"theirs", "keep"}) {
		t.Fatalf("source holds %v, want [theirs keep]", got)
	}
	deleted, err := c.searchDeleted()
	if err != nil {
		t.Fatalf("search deleted: %v", err)
	}
	if !slices.Equal(deleted, []imap.UID{theirs}) {
		t.Errorf("\\Deleted uids = %v, want just %v", deleted, theirs)
	}

	// and the message really did arrive at the destination.
	if got := mailboxSubjects(t, c, "Archive"); !slices.Equal(got, []string{"ours"}) {
		t.Errorf("destination holds %v, want [ours]", got)
	}
}

// TestMoveWithMoveExtension covers the ordinary path, where the server does it
// all in one command.
func TestMoveWithMoveExtension(t *testing.T) {
	caps := imap.CapSet{imap.CapIMAP4rev1: {}, imap.CapMove: {}}
	c, wire := newTestServer(t, testOptions{caps: caps, subjects: []string{"theirs", "ours", "keep"}})
	if err := createMailbox(c, "Archive"); err != nil {
		t.Fatalf("create Archive: %v", err)
	}

	theirs := uidOf(t, c, "theirs")
	if err := c.storeMany([]imap.UID{theirs}, imap.StoreFlagsAdd, imap.FlagDeleted); err != nil {
		t.Fatalf("mark theirs: %v", err)
	}
	wire.reset()

	if err := c.MoveMessages([]imap.UID{uidOf(t, c, "ours")}, "Archive"); err != nil {
		t.Fatalf("MoveMessages: %v", err)
	}
	if wire.hasBareExpunge() {
		t.Errorf("a move on a MOVE-capable server sent an EXPUNGE\n%s", wire.String())
	}
	if got := mailboxSubjects(t, c, "INBOX"); !slices.Equal(got, []string{"theirs", "keep"}) {
		t.Errorf("source holds %v, want [theirs keep]", got)
	}
}

// TestMoveBatchWithoutMoveExtension moves several at once, which is what a
// delete of a multi-selection turns into.
func TestMoveBatchWithoutMoveExtension(t *testing.T) {
	c, _ := newTestServer(t, testOptions{subjects: []string{"a", "b", "c", "keep"}})
	if err := createMailbox(c, "Trash"); err != nil {
		t.Fatalf("create Trash: %v", err)
	}

	uids := []imap.UID{uidOf(t, c, "a"), uidOf(t, c, "b"), uidOf(t, c, "c")}
	if err := c.MoveMessages(uids, "Trash"); err != nil {
		t.Fatalf("MoveMessages: %v", err)
	}
	if got := mailboxSubjects(t, c, "INBOX"); !slices.Equal(got, []string{"keep"}) {
		t.Fatalf("source holds %v, want just [keep]", got)
	}
	if got := mailboxSubjects(t, c, "Trash"); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("destination holds %v, want [a b c]", got)
	}
}

// TestMoveNothingIsANoOp: an empty set must not reach the copy or the delete.
func TestMoveNothingIsANoOp(t *testing.T) {
	c, wire := newTestServer(t, testOptions{subjects: []string{"theirs", "keep"}})
	if err := createMailbox(c, "Trash"); err != nil {
		t.Fatalf("create Trash: %v", err)
	}
	theirs := uidOf(t, c, "theirs")
	if err := c.storeMany([]imap.UID{theirs}, imap.StoreFlagsAdd, imap.FlagDeleted); err != nil {
		t.Fatalf("mark theirs: %v", err)
	}
	wire.reset()

	if err := c.MoveMessages(nil, "Trash"); err != nil {
		t.Fatalf("MoveMessages: %v", err)
	}
	if wire.hasBareExpunge() {
		t.Errorf("an empty move reached EXPUNGE\n%s", wire.String())
	}
	if got := mailboxSubjects(t, c, "INBOX"); !slices.Equal(got, []string{"theirs", "keep"}) {
		t.Errorf("source holds %v, want both messages", got)
	}
}

// TestMoveToAMissingMailboxKeepsTheSource: the copy fails, so nothing may be
// removed from where it is.
func TestMoveToAMissingMailboxKeepsTheSource(t *testing.T) {
	c, _ := newTestServer(t, testOptions{subjects: []string{"ours", "keep"}})

	if err := c.MoveMessages([]imap.UID{uidOf(t, c, "ours")}, "NoSuchMailbox"); err == nil {
		t.Fatal("MoveMessages to a missing mailbox returned no error")
	}
	if got := mailboxSubjects(t, c, "INBOX"); !slices.Equal(got, []string{"ours", "keep"}) {
		t.Errorf("source holds %v, want both messages", got)
	}
}
