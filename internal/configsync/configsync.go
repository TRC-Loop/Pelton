// Package configsync mirrors settings and locally-only message state through
// a folder the user points at (a Nextcloud/Dropbox/iCloud Drive sync
// directory, or anything else their OS keeps in sync across devices). Pelton
// itself never talks to a cloud provider; it only reads and writes plain
// files in that folder and lets whatever already syncs it do the transport.
//
// Two modes:
//   - Mirror: the device's own app-support directory stays put. Every sync
//     pass pulls in anything newer from the folder (per key, last write
//     wins by timestamp) and then pushes the resulting local state back out,
//     so the folder always reflects the merge. Safe with several devices
//     open at once; you pick what's included (settings, message metadata,
//     or the full offline cache).
//   - InPlace: the device's entire data directory (database, attachments,
//     settings - everything) simply IS the chosen folder; see inplace.go.
//     There is no scope and no periodic sync pass, since it is all just one
//     set of files that your cloud tool already keeps in sync. Only one
//     device may have Pelton open against the folder at a time - concurrent
//     writers to one live sqlite file can corrupt it.
//
// Credentials never enter this package: they live in the OS keyring and are
// not part of the settings table or the message extras this syncs.
package configsync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

// Mode selects the sync direction.
type Mode string

const (
	ModeMirror  Mode = "mirror"
	ModeInPlace Mode = "inplace"
)

// EmailScope selects how much local message state is included.
type EmailScope string

const (
	EmailScopeOff      EmailScope = "off"
	EmailScopeMetadata EmailScope = "metadata"
	EmailScopeFull     EmailScope = "full"
)

// Config is the user's chosen sync setup, persisted as a single JSON setting.
type Config struct {
	Enabled      bool       `json:"enabled"`
	Mode         Mode       `json:"mode"`
	Path         string     `json:"path"`
	SyncSettings bool       `json:"syncSettings"`
	EmailScope   EmailScope `json:"emailScope"`
	// LastSyncUnix is when TriggerSync last completed successfully.
	LastSyncUnix int64 `json:"lastSyncUnix"`
	// LastError is the most recent sync failure, cleared on success, shown in
	// the settings UI so a broken folder path is not silently ignored.
	LastError string `json:"lastError"`
}

const settingKey = "configsync_config"

const (
	settingsFileName   = "settings.json"
	metadataFileName   = "message-metadata.json"
	dbSnapshotName     = "mail-cache.sqlite"
	attachmentsDir     = "attachments"
	pendingRestoreFile = "configsync_pending_full_restore"
)

// Manager owns the persisted config, the folder watcher, and every sync pass.
type Manager struct {
	store      *storage.DB
	log        *slog.Logger
	stateDir   string // default app-support directory; holds the pending-full-restore and in-place markers
	dbFileName string

	mu      sync.Mutex
	cfg     Config
	watcher *fsnotify.Watcher
	stopCh  chan struct{}
	// debounce coalesces a burst of filesystem events (many cloud clients
	// write several files in quick succession) into one sync pass.
	debounce *time.Timer
	onSync   func(Config)
}

// New creates a Manager. stateDir is the device's default (non-in-place)
// app-support directory, where the pending-full-restore and in-place markers
// live across restarts. dbFileName is the sqlite file name.
func New(store *storage.DB, stateDir, dbFileName string, log *slog.Logger) *Manager {
	return &Manager{store: store, stateDir: stateDir, dbFileName: dbFileName, log: log}
}

// OnSync registers a callback fired after every sync pass (success or
// failure) with the current config, so the UI can reflect status live.
func (m *Manager) OnSync(fn func(Config)) {
	m.mu.Lock()
	m.onSync = fn
	m.mu.Unlock()
}

// Start loads the persisted config and, if enabled, begins watching. Call
// once at app startup.
func (m *Manager) Start(ctx context.Context) error {
	cfg, err := m.load(ctx)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.cfg = cfg
	m.mu.Unlock()
	if cfg.Enabled && cfg.Mode != ModeInPlace {
		return m.startWatching(ctx)
	}
	return nil
}

// Status returns the current config.
func (m *Manager) Status() Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg
}

// Configure validates and persists a new setup (or a change to path/mode/
// scope on an existing one). Switching into or out of InPlace mode only
// takes effect on the next app start (see inplace.go); everything else
// (re)starts the watcher and runs an initial sync pass immediately.
//
// mergeOnJoin only matters when cfg.Mode is InPlace and cfg.Path already
// holds another device's data: true merges this device's accounts and
// settings into it (see merge.go), false discards this device's local state
// and simply adopts what's there.
func (m *Manager) Configure(ctx context.Context, cfg Config, mergeOnJoin bool) error {
	if cfg.Path == "" {
		return fmt.Errorf("configsync: a folder is required")
	}
	if cfg.Mode != ModeMirror && cfg.Mode != ModeInPlace {
		return fmt.Errorf("configsync: unknown mode %q", cfg.Mode)
	}

	previous := m.Status()
	if previous.Enabled && previous.Mode == ModeInPlace && cfg.Mode != ModeInPlace {
		if err := DisableInPlace(ctx, m.store, m.stateDir, m.dbFileName); err != nil {
			return err
		}
	}
	if cfg.Mode == ModeInPlace {
		joining, err := PeekInPlaceFolder(cfg.Path, m.dbFileName)
		if err != nil {
			return err
		}
		if joining && mergeOnJoin {
			if err := MergeIntoInPlaceFolder(ctx, m.store, m.stateDir, cfg.Path, m.dbFileName); err != nil {
				return err
			}
		} else if err := EnableInPlace(ctx, m.store, m.stateDir, cfg.Path, m.dbFileName); err != nil {
			return err
		}
	} else if err := os.MkdirAll(cfg.Path, 0o755); err != nil {
		return fmt.Errorf("configsync: folder %q is not usable: %w", cfg.Path, err)
	}
	cfg.Enabled = true

	m.stopWatchingLocked()
	m.mu.Lock()
	m.cfg = cfg
	m.mu.Unlock()
	if err := m.persist(ctx, cfg); err != nil {
		return err
	}
	if cfg.Mode == ModeInPlace {
		return nil
	}
	if err := m.startWatching(ctx); err != nil {
		return err
	}
	return m.TriggerSync(ctx)
}

// PeekFolder reports whether path already holds another device's in-place
// data and, if so, a summary of what's there, for the setup ui to show
// before the user commits to joining it.
func (m *Manager) PeekFolder(ctx context.Context, path string) (bool, FolderSummary, error) {
	exists, err := PeekInPlaceFolder(path, m.dbFileName)
	if err != nil || !exists {
		return exists, FolderSummary{}, err
	}
	summary, err := SummarizeInPlaceFolder(ctx, path, m.dbFileName)
	return true, summary, err
}

// Close stops watching without changing the persisted config, so a normal
// app shutdown does not disable sync for next launch.
func (m *Manager) Close() {
	m.stopWatchingLocked()
}

// Disable stops watching and turns sync off. In Mirror mode the folder's
// contents are left untouched (another device may still be using them); in
// InPlace mode the live data is first copied back to the default directory
// so it is not stranded there, and the switch back takes effect next start.
func (m *Manager) Disable(ctx context.Context) error {
	m.stopWatchingLocked()
	m.mu.Lock()
	cfg := m.cfg
	m.mu.Unlock()

	if cfg.Enabled && cfg.Mode == ModeInPlace {
		if err := DisableInPlace(ctx, m.store, m.stateDir, m.dbFileName); err != nil {
			return err
		}
	}

	m.mu.Lock()
	m.cfg.Enabled = false
	cfg = m.cfg
	m.mu.Unlock()
	return m.persist(ctx, cfg)
}

func (m *Manager) load(ctx context.Context) (Config, error) {
	var cfg Config
	err := m.store.GetJSON(ctx, settingKey, &cfg)
	if err != nil {
		if err == storage.ErrSettingNotFound {
			return Config{Mode: ModeMirror, EmailScope: EmailScopeOff}, nil
		}
		return Config{}, err
	}
	return cfg, nil
}

func (m *Manager) persist(ctx context.Context, cfg Config) error {
	return m.store.SetJSON(ctx, settingKey, cfg)
}

func (m *Manager) startWatching(ctx context.Context) error {
	m.mu.Lock()
	path := m.cfg.Path
	m.mu.Unlock()

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("configsync: start watcher: %w", err)
	}
	if err := w.Add(path); err != nil {
		w.Close()
		return fmt.Errorf("configsync: watch %q: %w", path, err)
	}

	m.mu.Lock()
	m.watcher = w
	m.stopCh = make(chan struct{})
	stopCh := m.stopCh
	m.mu.Unlock()

	go m.watchLoop(ctx, w, stopCh)
	return nil
}

func (m *Manager) watchLoop(ctx context.Context, w *fsnotify.Watcher, stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}
			m.scheduleSync(ctx)
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			m.log.Warn("configsync watcher error", "err", err)
		}
	}
}

// scheduleSync debounces a burst of filesystem events into a single sync pass
// a couple seconds later, since a cloud client typically writes several files
// back to back for one logical change.
func (m *Manager) scheduleSync(ctx context.Context) {
	m.mu.Lock()
	if m.debounce != nil {
		m.debounce.Stop()
	}
	m.debounce = time.AfterFunc(2*time.Second, func() {
		if err := m.TriggerSync(ctx); err != nil {
			m.log.Warn("configsync auto sync failed", "err", err)
		}
	})
	m.mu.Unlock()
}

func (m *Manager) stopWatchingLocked() {
	m.mu.Lock()
	if m.stopCh != nil {
		close(m.stopCh)
		m.stopCh = nil
	}
	if m.watcher != nil {
		m.watcher.Close()
		m.watcher = nil
	}
	if m.debounce != nil {
		m.debounce.Stop()
		m.debounce = nil
	}
	m.mu.Unlock()
}

// TriggerSync runs one sync pass: pull first, then (copy mode only) push, so
// a device that just pulled a remote change immediately re-publishes the
// merged state rather than a stale local snapshot.
func (m *Manager) TriggerSync(ctx context.Context) error {
	m.mu.Lock()
	cfg := m.cfg
	m.mu.Unlock()

	if !cfg.Enabled {
		return nil
	}

	err := m.syncOnce(ctx, cfg)

	m.mu.Lock()
	if err != nil {
		m.cfg.LastError = err.Error()
	} else {
		m.cfg.LastError = ""
		m.cfg.LastSyncUnix = time.Now().Unix()
	}
	cfg = m.cfg
	onSync := m.onSync
	m.mu.Unlock()

	if persistErr := m.persist(ctx, cfg); persistErr != nil && err == nil {
		err = persistErr
	}
	if onSync != nil {
		onSync(cfg)
	}
	return err
}

func (m *Manager) syncOnce(ctx context.Context, cfg Config) error {
	if cfg.Mode == ModeInPlace {
		return nil
	}
	if cfg.SyncSettings {
		if err := m.pullSettings(ctx, cfg); err != nil {
			return fmt.Errorf("pull settings: %w", err)
		}
	}
	if cfg.EmailScope == EmailScopeMetadata {
		if err := m.pullMetadata(ctx, cfg); err != nil {
			return fmt.Errorf("pull message metadata: %w", err)
		}
	}
	if cfg.EmailScope == EmailScopeFull {
		if err := m.checkPendingFullRestore(cfg); err != nil {
			return fmt.Errorf("check full cache restore: %w", err)
		}
	}

	if cfg.Mode != ModeMirror {
		return nil
	}

	if cfg.SyncSettings {
		if err := m.pushSettings(ctx, cfg); err != nil {
			return fmt.Errorf("push settings: %w", err)
		}
	}
	if cfg.EmailScope == EmailScopeMetadata {
		if err := m.pushMetadata(ctx, cfg); err != nil {
			return fmt.Errorf("push message metadata: %w", err)
		}
	}
	if cfg.EmailScope == EmailScopeFull {
		if err := m.pushFullSnapshot(ctx, cfg); err != nil {
			return fmt.Errorf("push mail cache snapshot: %w", err)
		}
	}
	return nil
}

func writeJSONFile(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readJSONFile(path string, v any) (bool, error) {
	data, err := readFileRetrying(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return false, err
	}
	return true, nil
}

// readFileRetrying reads path, retrying a few times with a short backoff
// before giving up (cloud-sync placeholder files can transiently fail a read
// while still hydrating).
func readFileRetrying(path string) ([]byte, error) {
	const attempts = 5
	var data []byte
	var err error
	for i := range attempts {
		data, err = os.ReadFile(path)
		if err == nil || os.IsNotExist(err) {
			return data, err
		}
		time.Sleep(time.Duration(200*(i+1)) * time.Millisecond)
	}
	return nil, fmt.Errorf("%w (if this folder is a cloud-sync placeholder, it may still be downloading)", err)
}

func joinPath(dir string, parts ...string) string {
	all := append([]string{dir}, parts...)
	return filepath.Join(all...)
}
