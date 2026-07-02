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
	MarkDeleted(uids ...imap.UID) error
	Expunge(uids ...imap.UID) error
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
	New              int  // fetched from server into the cache
	Deleted          int  // removed from the cache (server side or pushed delete)
	FlagUpdated      int  // server flag changes adopted locally
	Conflicts        int  // messages changed on both sides
	Pushed           int  // local flag or delete operations sent to the server
	UIDValidityReset bool // the cache for the folder was dropped and refetched
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

	localStates, err := e.store.ListMessageStates(ctx, folder.ID)
	if err != nil {
		return res, fmt.Errorf("sync: load local states for folder %q: %w", folder.IMAPPath, err)
	}
	locals, localByUID := localView(localStates)

	servers, serverColors, err := loadServerView(e.client)
	if err != nil {
		return res, err
	}

	plan := BuildPlan(locals, servers)
	e.executePlan(ctx, folder, plan, localByUID, &res)

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

	for _, d := range plan {
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
			if err := e.fetchAndStore(ctx, folder, d.UID); err != nil {
				e.log.Error("fetch new message failed", "uid", d.UID, "err", err)
				continue
			}
			res.New++

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
