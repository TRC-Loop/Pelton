package sync

import (
	"context"
	"slices"
	"testing"

	"github.com/emersion/go-imap/v2"
)

// TestSyncOnlyDeletesWhatTheUserDeleted is the guard on the whole destructive
// path: a folder sync must hand the imap layer exactly the uids the user marked
// for deletion, and never widen that set because the cache and the server
// disagree about anything else.
func TestSyncOnlyDeletesWhatTheUserDeleted(t *testing.T) {
	ctx := context.Background()
	db, folder := newSyncTestFolder(t)
	client := &fakeClient{uids: []uint32{1, 2, 3, 4, 5}}
	engine := NewEngine(client, db, nil)

	if _, err := engine.SyncFolder(ctx, folder); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	states, err := db.ListMessageStates(ctx, folder.ID)
	if err != nil {
		t.Fatalf("list states: %v", err)
	}
	if len(states) != 5 {
		t.Fatalf("cached %d messages, want 5", len(states))
	}
	// the user deletes exactly one of them.
	var target uint32
	for _, s := range states {
		if s.UID == 3 {
			target = s.UID
			if err := db.MarkDeletePending(ctx, s.ID); err != nil {
				t.Fatalf("mark pending: %v", err)
			}
		}
	}
	if target == 0 {
		t.Fatal("uid 3 was not cached")
	}

	if _, err := engine.SyncFolder(ctx, folder); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	if !slices.Equal(client.deleted, []imap.UID{3}) {
		t.Errorf("sync asked the server to delete %v, want just [3]", client.deleted)
	}
}

// TestSyncDeletesNothingOnTheServerWhenTheCacheIsStale covers the case that
// would be worst: the server dropped messages the cache still has. That is a
// local cleanup, and it must never turn into a server-side delete.
func TestSyncDeletesNothingOnTheServerWhenTheCacheIsStale(t *testing.T) {
	ctx := context.Background()
	db, folder := newSyncTestFolder(t)
	client := &fakeClient{uids: []uint32{1, 2, 3, 4, 5}}
	engine := NewEngine(client, db, nil)

	if _, err := engine.SyncFolder(ctx, folder); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	// the server now reports an empty mailbox, the shape a bad SELECT or another
	// client emptying the folder would produce.
	client.uids = nil
	res, err := engine.SyncFolder(ctx, folder)
	if err != nil {
		t.Fatalf("second sync: %v", err)
	}
	if res.Deleted != 5 {
		t.Errorf("dropped %d cached messages, want 5", res.Deleted)
	}
	if len(client.deleted) != 0 {
		t.Errorf("sync deleted %v on the server, want nothing", client.deleted)
	}
}

// TestSyncDoesNotPushADeleteTheServerAlreadyApplied: the message is gone
// upstream, so there is nothing to expunge and the local row just goes.
func TestSyncDoesNotPushADeleteTheServerAlreadyApplied(t *testing.T) {
	ctx := context.Background()
	db, folder := newSyncTestFolder(t)
	client := &fakeClient{uids: []uint32{1, 2}}
	engine := NewEngine(client, db, nil)

	if _, err := engine.SyncFolder(ctx, folder); err != nil {
		t.Fatalf("initial sync: %v", err)
	}
	states, err := db.ListMessageStates(ctx, folder.ID)
	if err != nil {
		t.Fatalf("list states: %v", err)
	}
	for _, s := range states {
		if s.UID == 2 {
			if err := db.MarkDeletePending(ctx, s.ID); err != nil {
				t.Fatalf("mark pending: %v", err)
			}
		}
	}

	client.uids = []uint32{1}
	if _, err := engine.SyncFolder(ctx, folder); err != nil {
		t.Fatalf("second sync: %v", err)
	}
	if len(client.deleted) != 0 {
		t.Errorf("sync deleted %v on the server, want nothing", client.deleted)
	}
}
