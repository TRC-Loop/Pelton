package sync

import "slices"

// The sync window (#175). A folder's cache does not have to cover the whole
// mailbox: sync_floor_uid records the lowest uid it does cover, and everything
// below that is left on the server until the user asks for it. Without this the
// first sync of a decade-old mailbox downloads every message body, oldest uid
// first, and the user watches 2016 arrive for half an hour before today's mail
// shows up.
//
// The floor is a uid rather than a count because uids are what reconcile works
// in and what the server can be asked about cheaply. 0 always means "no floor":
// the folder is cached in full, which is the state every folder synced by an
// older version is already in, so an upgrade changes nothing for them.

// descendingUIDs returns every server uid, newest (highest) first.
func descendingUIDs(servers []ServerMessage) []uint32 {
	uids := make([]uint32, 0, len(servers))
	for _, s := range servers {
		uids = append(uids, s.UID)
	}
	slices.Sort(uids)
	slices.Reverse(uids)
	return uids
}

// floorForLimit returns the floor that keeps exactly the limit newest messages
// in the window. It returns 0 when the limit is unlimited (<= 0) or the folder
// already has no more than limit messages, since there is then nothing to hold
// back.
func floorForLimit(servers []ServerMessage, limit int) uint32 {
	if limit <= 0 || len(servers) <= limit {
		return 0
	}
	return descendingUIDs(servers)[limit-1]
}

// lowerFloor returns the floor that admits the batch newest messages currently
// below current, for one "load older" step. It returns 0 once fewer than batch
// messages remain below the floor, which is how the folder reaches "fully
// cached" and stops offering to fetch more.
func lowerFloor(servers []ServerMessage, current uint32, batch int) uint32 {
	if current == 0 {
		return 0
	}
	var below []uint32
	for _, uid := range descendingUIDs(servers) {
		if uid < current {
			below = append(below, uid)
		}
	}
	if batch <= 0 || len(below) <= batch {
		return 0
	}
	return below[batch-1]
}

// normalizeFloor clears a floor that no longer holds anything back, which
// happens when the server-side messages below it were deleted. Without this the
// ui would keep offering a "load older" that fetches nothing.
func normalizeFloor(servers []ServerMessage, floor uint32) uint32 {
	if floor == 0 {
		return 0
	}
	for _, s := range servers {
		if s.UID < floor {
			return floor
		}
	}
	return 0
}
