// Package configsync mirrors settings and locally-only message state through
// a folder the user points at (a Nextcloud/Dropbox/iCloud Drive sync
// directory, or anything else their OS keeps in sync across devices). Pelton
// itself never talks to a cloud provider; it only reads and writes plain
// files in that folder and lets whatever already syncs it do the transport.
//
// Two modes:
//   - Copy: this device's local state is the source of truth. Every sync
//     pass pulls in anything newer from the folder (per key, last write
//     wins by timestamp) and then pushes the resulting local state back out,
//     so the folder always reflects the merge.
//   - ReadOnly: this device only ever pulls from the folder. It never writes
//     to it, so it cannot clobber what other, copy-mode devices have put
//     there.
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
	ModeCopy     Mode = "copy"
	ModeReadOnly Mode = "readonly"
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
	store    *storage.DB
	log      *slog.Logger
	stateDir string // directory holding the pending-full-restore marker

	mu      sync.Mutex
	cfg     Config
	watcher *fsnotify.Watcher
	stopCh  chan struct{}
	// debounce coalesces a burst of filesystem events (many cloud clients
	// write several files in quick succession) into one sync pass.
	debounce *time.Timer
	onSync   func(Config)
}

// New creates a Manager. stateDir is where the pending-full-restore marker
// lives across restarts (the same directory as the database file).
func New(store *storage.DB, stateDir string, log *slog.Logger) *Manager {
	return &Manager{store: store, stateDir: stateDir, log: log}
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
	if cfg.Enabled {
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
// scope on an existing one), then (re)starts the watcher and runs an initial
// sync pass.
func (m *Manager) Configure(ctx context.Context, cfg Config) error {
	if cfg.Path == "" {
		return fmt.Errorf("configsync: a folder is required")
	}
	if cfg.Mode != ModeCopy && cfg.Mode != ModeReadOnly {
		return fmt.Errorf("configsync: unknown mode %q", cfg.Mode)
	}
	if err := os.MkdirAll(cfg.Path, 0o755); err != nil {
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
	if err := m.startWatching(ctx); err != nil {
		return err
	}
	return m.TriggerSync(ctx)
}

// Close stops watching without changing the persisted config, so a normal
// app shutdown does not disable sync for next launch.
func (m *Manager) Close() {
	m.stopWatchingLocked()
}

// Disable stops watching and turns sync off, leaving the folder's contents
// untouched (the user may still want them, or another device may still be
// using them).
func (m *Manager) Disable(ctx context.Context) error {
	m.stopWatchingLocked()
	m.mu.Lock()
	m.cfg.Enabled = false
	cfg := m.cfg
	m.mu.Unlock()
	return m.persist(ctx, cfg)
}

func (m *Manager) load(ctx context.Context) (Config, error) {
	var cfg Config
	err := m.store.GetJSON(ctx, settingKey, &cfg)
	if err != nil {
		if err == storage.ErrSettingNotFound {
			return Config{Mode: ModeCopy, EmailScope: EmailScopeOff}, nil
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

	if cfg.Mode != ModeCopy {
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
	data, err := os.ReadFile(path)
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

func joinPath(dir string, parts ...string) string {
	all := append([]string{dir}, parts...)
	return filepath.Join(all...)
}
