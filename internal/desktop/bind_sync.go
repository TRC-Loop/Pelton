package desktop

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/TRC-Loop/Pelton/internal/outbox"
	psmtp "github.com/TRC-Loop/Pelton/internal/smtp"
	"github.com/TRC-Loop/Pelton/internal/storage"
	psync "github.com/TRC-Loop/Pelton/internal/sync"

	pimap "github.com/TRC-Loop/Pelton/internal/imap"
)

// syncMu serializes imap sessions per process so manual and background syncs do
// not open competing logins for the same account at once.
var syncMu sync.Mutex

// startBackgroundServices launches the outbox worker and the initial sync plus
// per-account idle loops. Credentials come from the keyring (added by the
// wizard) with an environment fallback for the legacy cli account.
func (a *App) startBackgroundServices() {
	go a.runOutboxWorker()
	go a.runInitialSyncAndIdle()
	go a.runSnoozePoller()
	go a.harvestAddressBook()
	go a.runAutoSyncLoop()
}

// runAutoSyncLoop periodically runs a full sync pass across every account, on
// top of the always-on imap idle push (which not every server supports, and
// which can silently drop on flaky networks). the interval is a user setting
// (0 disables it); a short base tick lets a changed interval or low-power
// toggle take effect promptly without needing its own change-notification
// channel. it does nothing while low-power mode is on.
func (a *App) runAutoSyncLoop() {
	const baseTick = 5 * time.Second
	ticker := time.NewTicker(baseTick)
	defer ticker.Stop()
	lastRun := time.Now()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			interval := a.intSetting(settingAutoSync, 900)
			if interval <= 0 || a.lowPowerMode() {
				continue
			}
			if time.Since(lastRun) < time.Duration(interval)*time.Second {
				continue
			}
			lastRun = time.Now()
			if err := a.TriggerSync(); err != nil && !errors.Is(err, errNoCredentials) {
				a.log.Error("auto sync", "err", err)
			}
		}
	}
}

// runOutboxWorker drains the outbox, resolving smtp credentials per message from
// the sending account. Messages whose account has no credentials stay queued and
// surface in the outbox view.
func (a *App) runOutboxWorker() {
	transmitter := &accountTransmitter{app: a}
	worker := outbox.NewWorker(a.queue, transmitter,
		outbox.WithLogger(a.log),
		// emit after every state change so the ui reflects sending -> sent/failed
		// promptly. without this the outbox banner stayed stuck on "sending".
		outbox.WithOnChange(func() { a.emit(EventOutboxChanged, nil) }),
	)
	if _, err := a.queue.RequeueStuck(a.ctx); err != nil {
		a.log.Error("requeue stuck outbox", "err", err)
	}
	if err := worker.Run(a.ctx); err != nil && a.ctx.Err() == nil {
		a.log.Error("outbox worker stopped", "err", err)
	}
}

// accountTransmitter sends a queued message using the credentials of its
// account, resolved fresh each attempt so refreshed oauth tokens are picked up.
type accountTransmitter struct {
	app *App
}

func (t *accountTransmitter) Transmit(ctx context.Context, m outbox.Message) error {
	// note: the worker emits EventOutboxChanged via WithOnChange after the state
	// is persisted, so we must not emit here (that fired before markSent and left
	// the ui stuck on "sending").
	account, err := t.app.store.GetAccount(ctx, m.AccountID)
	if err != nil {
		return err
	}
	cfg, err := t.app.resolveSMTP(*account)
	if err != nil {
		return err
	}
	sender := psmtp.NewSender(cfg, psmtp.WithLogger(t.app.log))
	return sender.Transmit(ctx, m)
}

// runInitialSyncAndIdle syncs every account once, then parks each on idle.
func (a *App) runInitialSyncAndIdle() {
	accounts, err := a.store.ListAccounts(a.ctx)
	if err != nil {
		a.log.Error("list accounts for sync", "err", err)
		return
	}
	for _, account := range accounts {
		if err := a.syncAccount(account); err != nil && !errors.Is(err, errNoCredentials) {
			a.log.Error("initial sync", "account", account.Email, "err", err)
		}
		go a.idleLoop(account)
	}
}

// TriggerSync syncs all accounts on demand (the ui refresh action). It returns a
// clear error only when no account could be synced for lack of credentials.
func (a *App) TriggerSync() error {
	if err := a.ready(); err != nil {
		return err
	}
	accounts, err := a.store.ListAccounts(a.ctx)
	if err != nil {
		return err
	}

	synced := 0
	for _, account := range accounts {
		if err := a.syncAccount(account); err != nil {
			if errors.Is(err, errNoCredentials) {
				continue
			}
			a.log.Error("sync account", "account", account.Email, "err", err)
			continue
		}
		synced++
	}
	if synced == 0 && len(accounts) > 0 {
		return errNoCredentials
	}
	return nil
}

// syncAccount connects with the account's resolved credentials, syncs every
// folder emitting progress and new-mail events, then logs out.
func (a *App) syncAccount(account storage.Account) error {
	cfg, err := a.resolveIMAP(account)
	if err != nil {
		return err
	}

	syncMu.Lock()
	defer syncMu.Unlock()

	a.emit(EventSyncState, SyncStateEvent{Running: true})
	defer a.emit(EventSyncState, SyncStateEvent{Running: false})

	client, err := pimap.Connect(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Login(); err != nil {
		return err
	}
	defer client.Logout()

	return a.syncFolders(client, account.ID)
}

// syncFolders runs the sync engine over each stored folder of an account,
// emitting a progress event per folder and a new-mail event when one gained
// messages.
func (a *App) syncFolders(client *pimap.Client, accountID int64) error {
	folders, err := a.store.ListFolders(a.ctx, accountID)
	if err != nil {
		return err
	}
	engine := psync.NewEngine(client, a.store, a.log)
	engine.ColorSync = a.boolSetting(settingFlagColorSync, false)

	newTotal := 0
	for i, f := range folders {
		a.emit(EventSyncProgress, SyncProgressEvent{
			AccountID: accountID, Folder: f.Name, Done: i, Total: len(folders),
		})
		res, err := engine.SyncFolder(a.ctx, f)
		if err != nil {
			a.log.Error("sync folder", "folder", f.Name, "err", err)
			continue
		}
		if res.New > 0 {
			newTotal += res.New
			a.emit(EventMailNew, MailNewEvent{AccountID: accountID, FolderID: f.ID, Count: res.New})
		}
	}
	a.emit(EventSyncProgress, SyncProgressEvent{
		AccountID: accountID, Done: len(folders), Total: len(folders),
	})

	// index the freshly synced mail so it becomes searchable. run it off the sync
	// path so the search backfill never holds up the next sync.
	if newTotal > 0 {
		go a.indexNewMessages()
		if !a.lowPowerMode() {
			go a.harvestAddressBook()
		}
	}
	return nil
}

// findInboxFolder returns the account's INBOX folder row. IMAP's INBOX is a
// case-insensitive special name, so the match ignores case.
func (a *App) findInboxFolder(accountID int64) (*storage.Folder, error) {
	folders, err := a.store.ListFolders(a.ctx, accountID)
	if err != nil {
		return nil, err
	}
	for i := range folders {
		if strings.EqualFold(folders[i].IMAPPath, "INBOX") {
			return &folders[i], nil
		}
	}
	return nil, fmt.Errorf("no inbox folder for account %d", accountID)
}

// syncOneFolder runs the sync engine over a single folder, emitting the same
// progress/new-mail events syncFolders would, without touching any other
// folder on the account. Used by the idle push handler so a single INBOX
// update does not pay for a full-account resync.
func (a *App) syncOneFolder(client *pimap.Client, folder storage.Folder) error {
	engine := psync.NewEngine(client, a.store, a.log)
	engine.ColorSync = a.boolSetting(settingFlagColorSync, false)

	res, err := engine.SyncFolder(a.ctx, folder)
	if err != nil {
		return err
	}
	if res.New > 0 {
		a.emit(EventMailNew, MailNewEvent{AccountID: folder.AccountID, FolderID: folder.ID, Count: res.New})
		go a.indexNewMessages()
		if !a.lowPowerMode() {
			go a.harvestAddressBook()
		}
	}
	return nil
}

// idleLoop parks one account on imap idle and re-syncs when the server reports
// activity, reconnecting with a short backoff and exiting on app shutdown.
func (a *App) idleLoop(account storage.Account) {
	for a.ctx.Err() == nil {
		if err := a.idleSession(account); err != nil && a.ctx.Err() == nil {
			if errors.Is(err, errNoCredentials) {
				return
			}
			a.log.Error("idle session", "account", account.Email, "err", err)
			select {
			case <-a.ctx.Done():
				return
			case <-time.After(15 * time.Second):
			}
		}
	}
}

// idleSession opens one connection for an account, idles, and re-syncs on each
// server update until the connection drops or the context is cancelled.
func (a *App) idleSession(account storage.Account) error {
	cfg, err := a.resolveIMAP(account)
	if err != nil {
		return err
	}

	client, err := pimap.Connect(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Login(); err != nil {
		return err
	}
	defer client.Logout()

	if !client.SupportsIdle() {
		<-a.ctx.Done()
		return nil
	}

	// IDLE requires a selected mailbox; the server reports unsolicited activity
	// for whichever mailbox is selected, so we monitor INBOX (where new mail
	// lands). without this SELECT the server rejects IDLE outright.
	inbox, err := a.findInboxFolder(account.ID)
	if err != nil {
		return fmt.Errorf("look up inbox folder: %w", err)
	}
	if _, err := client.Select(inbox.IMAPPath); err != nil {
		return fmt.Errorf("select inbox for idle: %w", err)
	}

	go func() {
		for range client.Updates() {
			syncMu.Lock()
			// idle only watches INBOX, so only resync INBOX here; a full
			// resync of every folder would make each push wait on folders
			// that did not change, delaying the new mail this update is
			// actually about. other folders still get picked up by the
			// periodic full sync (runAutoSyncLoop).
			if err := a.syncOneFolder(client, *inbox); err != nil {
				a.log.Error("idle resync", "err", err)
			}
			syncMu.Unlock()
		}
	}()

	return client.Idle(a.ctx)
}
