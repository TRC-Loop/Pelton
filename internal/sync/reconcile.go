// Package sync orchestrates Pelton's imap and storage layers to keep a local
// cache and a remote mailbox in agreement, in both directions. It only calls
// the public surface of those two packages and never reaches into their
// internals, so the dependency goes one way: sync depends on imap and storage,
// not the reverse.
//
// The decision logic lives in reconcile.go and is pure: it takes the local and
// server view of a folder and returns a plan, with no database or network
// calls, so every conflict case can be unit tested deterministically.
package sync

import (
	"slices"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

// Conflict policy, documented here so it is obvious and easy to change later:
//
//   - existence: the server wins. if the server no longer has a message it is
//     removed locally even if it was modified locally. a local delete that the
//     server still has is pushed up (the user's delete intent is honoured while
//     the message still exists on both sides).
//   - flags: union/merge. the result is the bitwise OR of local and server
//     flags, so if either side set \Seen or \Flagged it stays set. this means a
//     local "mark unread" (clearing \Seen) is NOT propagated in this version,
//     because union never clears a flag. that is a deliberate v1 limitation: it
//     guarantees we never lose a "seen" state, at the cost of not syncing
//     un-flagging. a future version can track per-flag change direction.
//
// reconcile is pure: same inputs always give the same Decision.

// LocalMessage is reconcile's view of a cached message.
type LocalMessage struct {
	UID           uint32
	Flags         storage.Flag
	PendingFlags  bool // local flag change not yet pushed
	PendingDelete bool // local delete not yet pushed
}

// ServerMessage is reconcile's view of a message currently on the server.
type ServerMessage struct {
	UID   uint32
	Flags storage.Flag
}

// Action is the single operation reconcile decides on for one message.
type Action int

const (
	// ActionNone means local and server already agree.
	ActionNone Action = iota
	// ActionFetchNew means the server has a message the cache does not: fetch
	// its body and attachments and insert it.
	ActionFetchNew
	// ActionDeleteLocal means the server no longer has a message the cache does:
	// remove it from the cache. no server call.
	ActionDeleteLocal
	// ActionAdoptServerFlags means the server's flags changed and there is no
	// pending local change, so store the server flags locally.
	ActionAdoptServerFlags
	// ActionPushFlags means push the merged flags to the server, store the merge
	// locally and clear the pending marker.
	ActionPushFlags
	// ActionPushDelete means delete the message on the server, then remove it
	// from the cache and clear the pending marker.
	ActionPushDelete
	// ActionClearPending means a pending local flag change is already satisfied
	// on the server, so just store the merge locally and clear the marker.
	ActionClearPending
)

// Decision is reconcile's output for one message.
type Decision struct {
	UID    uint32
	Action Action
	// Flags is the target flag set for the local store and any server push,
	// meaningful for the flag-related actions.
	Flags storage.Flag
	// Conflict is true when both sides changed the same message since the last
	// sync, recorded for reporting regardless of the action taken.
	Conflict bool
}

// mergeFlags applies the union policy.
func mergeFlags(local, server storage.Flag) storage.Flag {
	return local | server
}

// Reconcile decides what to do for a single message given the local and server
// views. Either side may be nil (absent). It performs no io.
func Reconcile(local *LocalMessage, server *ServerMessage) Decision {
	switch {
	case local == nil && server != nil:
		return Decision{UID: server.UID, Action: ActionFetchNew}

	case local != nil && server == nil:
		// server wins for existence. if we had a pending flag change it is a lost
		// update, flagged as a conflict; a pending delete simply agrees with the
		// server and is not a conflict.
		return Decision{
			UID:      local.UID,
			Action:   ActionDeleteLocal,
			Conflict: local.PendingFlags && !local.PendingDelete,
		}

	case local != nil && server != nil:
		return reconcileBoth(local, server)
	}

	// both nil cannot happen through BuildPlan, but stay total.
	return Decision{Action: ActionNone}
}

func reconcileBoth(local *LocalMessage, server *ServerMessage) Decision {
	// a local delete intent wins while the message still exists on both sides.
	if local.PendingDelete {
		return Decision{UID: local.UID, Action: ActionPushDelete}
	}

	if local.PendingFlags {
		merged := mergeFlags(local.Flags, server.Flags)
		// without a stored baseline we cannot prove both sides changed, so we
		// treat it as a conflict only when the server independently has a managed
		// flag the local copy lacks. a pending change the server simply has not
		// received yet (server flags are a subset of local) is not a conflict.
		conflict := server.Flags&^local.Flags != 0
		if merged == server.Flags {
			// server already has everything we wanted, nothing to push up.
			return Decision{UID: local.UID, Action: ActionClearPending, Flags: merged, Conflict: conflict}
		}
		return Decision{UID: local.UID, Action: ActionPushFlags, Flags: merged, Conflict: conflict}
	}

	// no pending local change: the server is authoritative for flags.
	if server.Flags != local.Flags {
		return Decision{UID: local.UID, Action: ActionAdoptServerFlags, Flags: server.Flags}
	}
	return Decision{UID: local.UID, Action: ActionNone}
}

// BuildPlan reconciles a whole folder. It returns one Decision per message
// across the union of local and server uids, in ascending uid order so the plan
// is deterministic and easy to test. Pure: no io.
func BuildPlan(locals []LocalMessage, servers []ServerMessage) []Decision {
	localByUID := make(map[uint32]LocalMessage, len(locals))
	for _, l := range locals {
		localByUID[l.UID] = l
	}
	serverByUID := make(map[uint32]ServerMessage, len(servers))
	for _, s := range servers {
		serverByUID[s.UID] = s
	}

	uids := unionUIDs(localByUID, serverByUID)
	plan := make([]Decision, 0, len(uids))
	for _, uid := range uids {
		local, hasLocal := localByUID[uid]
		server, hasServer := serverByUID[uid]
		var lp *LocalMessage
		var sp *ServerMessage
		if hasLocal {
			l := local
			lp = &l
		}
		if hasServer {
			s := server
			sp = &s
		}
		plan = append(plan, Reconcile(lp, sp))
	}
	return plan
}

func unionUIDs(local map[uint32]LocalMessage, server map[uint32]ServerMessage) []uint32 {
	set := make(map[uint32]struct{}, len(local)+len(server))
	for uid := range local {
		set[uid] = struct{}{}
	}
	for uid := range server {
		set[uid] = struct{}{}
	}
	uids := make([]uint32, 0, len(set))
	for uid := range set {
		uids = append(uids, uid)
	}
	slices.Sort(uids)
	return uids
}
