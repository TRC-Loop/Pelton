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
	"sync"

	"github.com/TRC-Loop/Pelton/internal/outbox"
	"github.com/TRC-Loop/Pelton/internal/search"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// App is the bound application object. Its exported methods form the api the
// frontend calls.
type App struct {
	ctx   context.Context
	log   *slog.Logger
	store *storage.DB
	index *search.Index
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

	store, err := openStore(ctx)
	if err != nil {
		a.log.Error("open store", "err", err)
		return
	}
	a.store = store
	a.queue = outbox.NewQueue(store)

	// open the search index and bring it up to date in the background so startup
	// is not blocked by a large backfill. a failure here only disables search.
	if idx, err := openSearchIndex(); err != nil {
		a.log.Error("open search index", "err", err)
	} else {
		a.index = idx
		go a.backfillSearch()
	}

	a.startBackgroundServices()
}

// shutdown is the wails OnShutdown hook. It closes the store so the sqlite wal
// is checkpointed cleanly.
func (a *App) shutdown(ctx context.Context) {
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

// openStore opens the default database location and applies migrations. It is
// the same path the cli tools use, so accounts they created are visible here.
func openStore(ctx context.Context) (*storage.DB, error) {
	path, err := storage.DefaultPath()
	if err != nil {
		return nil, err
	}
	store, err := storage.Open(path)
	if err != nil {
		return nil, err
	}
	if err := store.RunMigrations(ctx); err != nil {
		store.Close()
		return nil, err
	}
	return store, nil
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
