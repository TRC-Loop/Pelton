// app.go wires the existing internal/* backend to the wails frontend. The App
// type is the single bound struct: every method here is callable from
// typescript through the generated bindings. App orchestrates only. It opens the
// store, owns the outbox queue and background services, and translates between
// internal package types and the flat dtos the ui consumes. No mail, crypto,
// sync or storage logic lives here; it all delegates to internal/*.
package desktop

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/TRC-Loop/Pelton/internal/configsync"
	"github.com/TRC-Loop/Pelton/internal/outbox"
	"github.com/TRC-Loop/Pelton/internal/search"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// App is the bound application object. Its exported methods form the api the
// frontend calls.
type App struct {
	ctx context.Context
	log *slog.Logger

	store *storage.DB
	index *search.Index
	sync  *configsync.Manager
	// defaultStateDir is the device's normal per-OS app-support directory,
	// independent of an active configsync in-place folder. It is where the
	// configsync markers live so they are discoverable regardless of which
	// directory storage actually opened from.
	defaultStateDir string
	// searchMu serializes index backfills so a startup pass and a post-sync pass
	// do not advance the watermark concurrently.
	searchMu sync.Mutex
	queue    *outbox.Queue
	version  string
	// embedded license data served to the about section on demand.
	licenseManifest string
	programLicense  string
}

// newApp creates the App with the build version. The heavy initialization
// happens in startup once wails has handed us a context we can emit runtime
// events on.
func newApp(version string) *App {
	return &App{
		log:     slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})),
		version: version,
	}
}

// startup is the wails OnStartup hook. It opens the store, runs migrations,
// builds the outbox queue and starts background services. A failure here is
// fatal to a useful app, so we log loudly; the ui surfaces the missing store via
// the bound methods returning errors.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	store, dataDir, err := openStore(ctx)
	if err != nil {
		a.log.Error("open store", "err", err)
		return
	}
	a.store = store
	a.queue = outbox.NewQueue(store)

	if defaultPath, pathErr := storage.DefaultPath(); pathErr == nil {
		a.defaultStateDir = filepath.Dir(defaultPath)
		a.sync = configsync.New(store, a.defaultStateDir, filepath.Base(defaultPath), a.log)
		a.sync.OnSync(func(cfg configsync.Config) {
			a.emit(EventConfigSync, cfg)
		})
		if err := a.sync.Start(ctx); err != nil {
			a.log.Warn("start config sync", "err", err)
		}
	}

	// open the search index and bring it up to date in the background so startup
	// is not blocked by a large backfill. a failure here only disables search.
	if idx, err := openSearchIndex(dataDir); err != nil {
		a.log.Error("open search index", "err", err)
	} else {
		a.index = idx
		go a.backfillSearch()
	}

	a.startBackgroundServices()

	// off by default; only runs at all if the user turned on a check
	// frequency in settings. backgrounded so a slow/unreachable network never
	// delays startup.
	go a.maybeAutoCheckForUpdates(ctx)

	// if a bulk offline download was still running when the app last closed,
	// pick it back up; planDownload skips anything already cached so this is
	// cheap when most of the range was already fetched.
	a.ResumePendingDownload()
}

// shutdown is the wails OnShutdown hook. It closes the store so the sqlite wal
// is checkpointed cleanly.
func (a *App) shutdown(ctx context.Context) {
	if a.sync != nil {
		a.sync.Close()
	}
	if a.index != nil {
		if err := a.index.Close(); err != nil {
			a.log.Error("close search index", "err", err)
		}
	}
	if a.store != nil {
		if err := a.store.Close(); err != nil {
			a.log.Error("close store", "err", err)
		}
	}
}

// openStore opens the database and applies migrations, returning the
// directory it opened from. That is normally the default per-OS app-support
// directory (the same path the cli tools use, so accounts they created are
// visible here), but configsync's in-place mode can redirect it to a folder
// the user chose instead - see configsync.ActiveDataDir.
func openStore(ctx context.Context) (*storage.DB, string, error) {
	defaultPath, err := storage.DefaultPath()
	if err != nil {
		return nil, "", err
	}
	defaultDir := filepath.Dir(defaultPath)
	dbFileName := filepath.Base(defaultPath)

	dataDir, err := configsync.ActiveDataDir(defaultDir, defaultDir)
	if err != nil {
		return nil, "", err
	}
	path := filepath.Join(dataDir, dbFileName)

	// if a previous run's full-cache config sync armed a restore (the live db
	// cannot be swapped out from under an open connection), apply it now,
	// before anything opens the database.
	attachmentsDir := filepath.Join(dataDir, "attachments")
	if err := configsync.ApplyPendingFullRestore(defaultDir, path, attachmentsDir); err != nil {
		return nil, "", err
	}
	store, err := storage.Open(path)
	if err != nil {
		return nil, "", err
	}
	if err := store.RunMigrations(ctx); err != nil {
		store.Close()
		return nil, "", err
	}
	return store, dataDir, nil
}

// AppVersion returns the build version string for the about section. It is set
// at build time via ldflags and defaults to "dev" in a plain build or dev run.
func (a *App) AppVersion() string {
	return a.version
}

// ready reports whether the store opened. Bound methods call this first so a
// failed startup yields a clear error instead of a nil pointer panic.
func (a *App) ready() error {
	if a.store == nil {
		return errStoreUnavailable
	}
	return nil
}
