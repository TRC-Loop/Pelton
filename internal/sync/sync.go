package sync

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/emersion/go-imap/v2"

	pimap "github.com/TRC-Loop/Pelton/internal/imap"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// mailClient is the slice of the imap client's public surface the sync engine
// uses. depending on an interface keeps the boundary explicit and lets tests
// substitute a fake without a live server.
type mailClient interface {
	Select(mailbox string) (*pimap.Mailbox, error)
	FetchAllFlags() ([]pimap.MessageHeader, error)
	FetchMessage(uid imap.UID) (*pimap.Message, error)
	AddFlags(uid imap.UID, flags ...imap.Flag) error
	DeleteMessages(uids ...imap.UID) error
	MoveMessages(uids []imap.UID, mailbox string) error
}

// Engine orchestrates one account's imap connection and the local store. It is
// created per connected account and is not safe for concurrent use, matching
// the imap client.
type Engine struct {
	client mailClient
	store  *storage.DB
	log    *slog.Logger
	// ColorSync, when true, adopts server-side flag colors (Thunderbird $LabelN
	// keywords) into the local cache during each folder sync.
	ColorSync bool
	// TrashPath and TrashFolderID identify the account's trash folder, which is
	// where a deleted message goes. Deleting from the trash itself, or from an
	// account with no trash folder (TrashPath empty), is the permanent delete
	// instead. The caller sets these because folder roles are resolved above
	// this layer.
	TrashPath     string
	TrashFolderID int64
	// InitialLimit caps how many of a folder's newest messages the first sync
	// fetches; the rest wait behind the folder's sync floor until a backfill asks
	// for them (see window.go). 0 means no cap, the pre-#175 behavior. It only
	// applies to a folder's first sync: once a folder is cached, lowering this
	// never discards anything.
	InitialLimit int
}

// NewEngine wires an imap client and the store together. A nil logger is
// replaced with a discarding default so callers need not pass one.
func NewEngine(client mailClient, store *storage.DB, log *slog.Logger) *Engine {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Engine{client: client, store: store, log: log}
}

// FolderSyncResult summarises one folder sync for logging and the cli.
type FolderSyncResult struct {
	New              int     // fetched from server into the cache
	NewIDs           []int64 // storage ids of the messages fetched this sync
	Deleted          int     // removed from the cache (server side or pushed delete)
	FlagUpdated      int     // server flag changes adopted locally
	Conflicts        int     // messages changed on both sides
	Pushed           int     // local flag or delete operations sent to the server
	UIDValidityReset bool    // the cache for the folder was dropped and refetched
	HasOlder         bool    // the server still holds messages below the sync window
	Repaired         int     // messages refetched because their cached text was broken
	RepairedIDs      []int64 // storage ids of those messages, for reindexing
}

// SyncAccount syncs every cached folder for an account. A failure on one folder
// is logged and does not stop the others, so one broken mailbox cannot block a
// whole account.
func (e *Engine) SyncAccount(ctx context.Context, accountID int64) error {
	folders, err := e.store.ListFolders(ctx, accountID)
	if err != nil {
		return fmt.Errorf("sync: list folders for account %d: %w", accountID, err)
	}
	for _, folder := range folders {
		if err := ctx.Err(); err != nil {
			return err
		}
		// a folder the user unchecked is never opened at all (#173).
		if folder.SyncExcluded {
			continue
		}
		res, err := e.SyncFolder(ctx, folder)
		if err != nil {
			e.log.Error("folder sync failed", "folder", folder.IMAPPath, "err", err)
			continue
		}
		e.log.Info("folder synced",
			"folder", folder.IMAPPath,
			"new", res.New, "deleted", res.Deleted, "flag_updated", res.FlagUpdated,
			"conflicts", res.Conflicts, "pushed", res.Pushed, "uidvalidity_reset", res.UIDValidityReset)
	}
	return nil
}

// SyncFolder runs a full bidirectional sync of one folder and returns a summary.
func (e *Engine) SyncFolder(ctx context.Context, folder storage.Folder) (FolderSyncResult, error) {
	return e.syncFolder(ctx, folder, 0)
}

// BackfillFolder lowers the folder's sync window by batch messages and syncs,
// fetching that many older messages from the server. It is a no-op returning a
// zero result when the folder has no floor, i.e. it is already cached in full.
// Callers use HasOlder on the result to decide whether more remain.
func (e *Engine) BackfillFolder(ctx context.Context, folder storage.Folder, batch int) (FolderSyncResult, error) {
	return e.syncFolder(ctx, folder, batch)
}

// syncFolder is the shared body of SyncFolder and BackfillFolder. backfill is
// how many older messages to admit into the window before planning; 0 keeps the
// window where it is, which is every ordinary sync.
func (e *Engine) syncFolder(ctx context.Context, folder storage.Folder, backfill int) (FolderSyncResult, error) {
	var res FolderSyncResult

	mbox, err := e.client.Select(folder.IMAPPath)
	if err != nil {
		return res, fmt.Errorf("sync: select folder %q: %w", folder.IMAPPath, err)
	}

	state, err := loadFolderSyncState(ctx, e.store, folder)
	if err != nil {
		return res, err
	}

	folder, reset, err := e.handleUIDValidity(ctx, folder, state.StoredUIDValidity, mbox.UIDValidity)
	if err != nil {
		return res, err
	}
	res.UIDValidityReset = reset
	// storedFloor is what the folder row actually holds, which the reset below
	// deliberately stops tracking; the write at the end compares against this so
	// a reset back to no floor still reaches the database.
	storedFloor := state.SyncFloorUID
	if reset {
		// the cache was dropped, so this is a first sync again and the window has
		// to be recomputed from scratch rather than kept at a floor that refers to
		// uids from the old UIDVALIDITY generation.
		state.SyncFloorUID = 0
		state.LastSeenUID = 0
	}

	localStates, err := e.store.ListMessageStates(ctx, folder.ID)
	if err != nil {
		return res, fmt.Errorf("sync: load local states for folder %q: %w", folder.IMAPPath, err)
	}
	locals, localByUID := localView(localStates)

	servers, serverColors, err := loadServerView(e.client)
	if err != nil {
		return res, err
	}

	floor := e.windowFloor(state, servers, len(localStates), backfill)
	if floor != storedFloor {
		if err := e.store.SetFolderSyncFloorUID(ctx, folder.ID, floor); err != nil {
			return res, err
		}
	}
	res.HasOlder = floor > 0

	plan := BuildPlan(locals, servers, floor)
	e.executePlan(ctx, folder, plan, localByUID, &res)

	// mail cached before charset detection existed is stored with bytes that
	// are not valid utf-8 and cannot be fixed locally, so it comes from the
	// server again. A few per sync, after the plan, so it never delays new mail.
	e.repairMangled(ctx, folder, &res)

	// adopt server-side color labels when color syncing is on. this runs after the
	// plan so newly fetched messages already have local rows to color.
	if e.ColorSync {
		e.adoptColors(ctx, folder, serverColors)
	}

	newHigh := max(highestUID(servers), state.LastSeenUID)
	if err := e.store.SetFolderLastSeenUID(ctx, folder.ID, newHigh); err != nil {
		return res, err
	}
	return res, nil
}

// windowFloor decides the folder's sync floor for this run: a backfill lowers
// the existing one, a first sync of a folder bigger than InitialLimit
// establishes one, and everything else keeps what is stored (minus a floor that
// no longer holds anything back).
//
// "First sync" is both an empty high water mark and an empty cache. Checking
// only the high water mark would re-cap a folder whose messages were all
// deleted server side, and checking only the cache would re-cap a folder the
// user had legitimately emptied.
func (e *Engine) windowFloor(state FolderSyncState, servers []ServerMessage, cached, backfill int) uint32 {
	if backfill > 0 {
		return lowerFloor(servers, state.SyncFloorUID, backfill)
	}
	if state.SyncFloorUID == 0 && state.LastSeenUID == 0 && cached == 0 {
		return floorForLimit(servers, e.InitialLimit)
	}
	return normalizeFloor(servers, state.SyncFloorUID)
}

// handleUIDValidity drops and refetches the folder cache if the server's
// UIDVALIDITY changed, which means every cached uid for the folder is stale.
// this is destructive but correct, so it is logged loudly. returns the folder
// with its updated uid_validity.
func (e *Engine) handleUIDValidity(ctx context.Context, folder storage.Folder, stored, server uint32) (storage.Folder, bool, error) {
	if stored == server {
		return folder, false, nil
	}

	reset := false
	// stored == 0 is a first sync, there is nothing cached to drop.
	if stored != 0 {
		e.log.Warn("uidvalidity changed, dropping stale cache for folder",
			"folder", folder.IMAPPath, "stored", stored, "server", server)
		n, err := e.store.PurgeFolderMessages(ctx, folder.AccountID, folder.ID)
		if err != nil {
			return folder, false, err
		}
		e.log.Warn("purged stale cached messages", "folder", folder.IMAPPath, "count", n)
		reset = true
	}

	if err := e.store.SetFolderUIDValidity(ctx, folder.ID, server); err != nil {
		return folder, false, err
	}
	if err := e.store.SetFolderLastSeenUID(ctx, folder.ID, 0); err != nil {
		return folder, false, err
	}
	folder.UIDValidity = server
	return folder, reset, nil
}

// adoptColors makes the server authoritative for flag colors: for each cached
// message whose stored color differs from the server keyword, it writes the
// server's color (0 clears). It only writes on a difference, so a steady state
// costs no writes.
func (e *Engine) adoptColors(ctx context.Context, folder storage.Folder, serverColors map[uint32]int) {
	states, err := e.store.ListMessageStates(ctx, folder.ID)
	if err != nil {
		e.log.Error("color sync: list states", "folder", folder.IMAPPath, "err", err)
		return
	}
	current, err := e.store.FolderFlagColors(ctx, folder.ID)
	if err != nil {
		e.log.Error("color sync: current colors", "folder", folder.IMAPPath, "err", err)
		return
	}
	for _, s := range states {
		desired, ok := serverColors[s.UID]
		if !ok {
			continue
		}
		if current[s.UID] != desired {
			if err := e.store.SetFlagColor(ctx, s.ID, desired); err != nil {
				e.log.Error("color sync: set color", "uid", s.UID, "err", err)
			}
		}
	}
}

// executePlan applies a reconciled plan. pull actions and flag pushes run
// inline; local deletions are batched so they cost two server round trips total.
// a failure on one message is logged and skipped so it cannot corrupt the cache
// or block the rest of the folder.
func (e *Engine) executePlan(ctx context.Context, folder storage.Folder, plan []Decision, localByUID map[uint32]storage.MessageState, res *FolderSyncResult) {
	var pendingDeletes []storage.MessageState

	// walked back to front, because BuildPlan orders by ascending uid and a full
	// body fetch is the expensive step here. Front to back means the oldest mail
	// downloads first and a large mailbox shows years-old messages for as long as
	// it takes to reach the present (#175); back to front puts the newest mail on
	// screen within seconds. Nothing else in the plan is order-sensitive: the
	// decisions are independent per uid and deletes are batched at the end.
	for i := len(plan) - 1; i >= 0; i-- {
		d := plan[i]
		if d.Conflict {
			res.Conflicts++
		}
		if err := ctx.Err(); err != nil {
			e.log.Warn("sync cancelled mid-folder", "folder", folder.IMAPPath)
			return
		}

		switch d.Action {
		case ActionNone:
			// already in agreement

		case ActionFetchNew:
			id, err := e.fetchAndStore(ctx, folder, d.UID)
			if err != nil {
				e.log.Error("fetch new message failed", "uid", d.UID, "err", err)
				continue
			}
			res.New++
			res.NewIDs = append(res.NewIDs, id)

		case ActionDeleteLocal:
			if err := e.deleteLocal(ctx, folder, localByUID[d.UID]); err != nil {
				e.log.Error("delete local message failed", "uid", d.UID, "err", err)
				continue
			}
			res.Deleted++

		case ActionAdoptServerFlags:
			if err := e.adoptServerFlags(ctx, localByUID[d.UID], d.Flags); err != nil {
				e.log.Error("adopt server flags failed", "uid", d.UID, "err", err)
				continue
			}
			res.FlagUpdated++

		case ActionPushFlags:
			if err := e.pushFlags(ctx, localByUID[d.UID], d.Flags); err != nil {
				e.log.Error("push flags failed", "uid", d.UID, "err", err)
				continue
			}
			res.Pushed++

		case ActionClearPending:
			if err := e.clearPending(ctx, localByUID[d.UID], d.Flags); err != nil {
				e.log.Error("clear pending flags failed", "uid", d.UID, "err", err)
				continue
			}

		case ActionPushDelete:
			pendingDeletes = append(pendingDeletes, localByUID[d.UID])
		}
	}

	if len(pendingDeletes) > 0 {
		if err := e.pushDeletes(ctx, folder, pendingDeletes); err != nil {
			e.log.Error("push deletes failed", "folder", folder.IMAPPath, "count", len(pendingDeletes), "err", err)
			return
		}
		res.Pushed += len(pendingDeletes)
		res.Deleted += len(pendingDeletes)
	}
}
