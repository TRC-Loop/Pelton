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
	"github.com/TRC-Loop/Pelton/internal/proxy"
	"github.com/TRC-Loop/Pelton/internal/search"
	"github.com/TRC-Loop/Pelton/internal/storage"
	"github.com/wailsapp/wails/v2/pkg/menu"
)

// App is the bound application object. Its exported methods form the api the
// frontend calls.
type App struct {
	ctx context.Context
	log *slog.Logger

	store *storage.DB
	// dataDir is the app data directory the store opened in; themes and the
	// search index live next to the database. Empty if the store failed to
	// open.
	dataDir string
	// storeReady closes once startup has finished assigning (or failing to
	// assign) store, giving domReady a happens-before edge before it reads
	// store on its own goroutine - without it, the read is a data race even
	// though a nil check keeps it from crashing.
	storeReady chan struct{}
	index      *search.Index
	// searchMu serializes index backfills so a startup pass and a post-sync pass
	// do not advance the watermark concurrently.
	searchMu sync.Mutex
	queue    *outbox.Queue
	version  string
	// embedded license data served to the about section on demand.
	licenseManifest string
	programLicense  string
	// trayIcon is the embedded .ico for the Windows notification-area icon
	// (see tray_windows.go); empty elsewhere.
	trayIcon []byte
	// mailMenuItems are the native Mail-menu items that act on the open message;
	// they start disabled and SetMailActionsEnabled toggles them as the frontend's
	// open message changes. mailActionsEnabled mirrors that same state so a menu
	// rebuild (RebuildMenu, on a language change) can restore it instead of
	// resetting every item back to disabled.
	mailMenuItems      []*menu.MenuItem
	mailActionsEnabled bool

	// dlMu guards dlCancel, the cancel function of the running bulk offline
	// download (nil when none is running). CancelDownload calls it to stop the
	// job without tearing down the whole app context.
	dlMu     sync.Mutex
	dlCancel context.CancelFunc

	// demoMode is the purely-cosmetic screenshot mode (the --potatoes-are-nice
	// flag): the frontend fills the ui with fixed sample data instead of reading
	// real accounts and mail. It never touches the store or the network.
	demoMode bool

	// proxyMu guards proxyCfg, the cached outbound proxy preference (with its
	// password from the keyring). It is loaded at startup and refreshed by
	// SetProxyConfig, so the mail and http paths read it without touching the
	// keyring on every connection.
	proxyMu  sync.RWMutex
	proxyCfg proxy.Config
}

// IsDemoMode reports whether the app was launched in the cosmetic demo mode. The
// frontend reads it once at startup to decide whether to render sample data.
func (a *App) IsDemoMode() bool {
	return a.demoMode
}

// IsDevMode reports whether the app is running against the separate dev data
// directory (the PELTON_DEV env var storage.DefaultPath checks), so the
// frontend can show a persistent indicator that this isn't a normal install -
// it's easy to forget a dev build is pointed at throwaway data instead of a
// real mailbox.
func (a *App) IsDevMode() bool {
	return os.Getenv("PELTON_DEV") != ""
}

// newApp creates the App with the build version. The heavy initialization
// happens in startup once wails has handed us a context we can emit runtime
// events on.
func newApp(version string) *App {
	return &App{
		log:        slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})),
		version:    version,
		storeReady: make(chan struct{}),
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
		close(a.storeReady)
		// even without a store the tray must come up: with the window closed
		// it is the only visible way to reopen or quit the app.
		a.startTray()
		return
	}
	a.store = store
	a.dataDir = dataDir
	a.queue = outbox.NewQueue(store)
	a.loadProxy()
	close(a.storeReady)

	// the Windows tray icon (no-op elsewhere). started after the store is up
	// so its menu labels can follow the language setting.
	a.startTray()

	// demo mode is purely cosmetic: the frontend renders fixed sample data, so we
	// skip everything that would touch the network or mutate the store (sync, idle,
	// the outbox worker, auto-update checks, download resume, migrations). the
	// store still opens so bound calls do not error, but nothing runs against it.
	if a.demoMode {
		return
	}

	// the old config-sync feature could redirect the data directory into a synced
	// folder ("in-place" mode). that feature is gone; if a device still has that
	// marker, migrate its data back to the normal app-support dir now, so the next
	// launch opens from the standard location again.
	if defaultPath, pathErr := storage.DefaultPath(); pathErr == nil {
		stateDir := filepath.Dir(defaultPath)
		if migrated, mErr := configsync.MigrateInPlaceBack(ctx, store, stateDir, filepath.Base(defaultPath)); mErr != nil {
			a.log.Error("migrate config-sync data back", "err", mErr)
		} else if migrated {
			a.log.Info("migrated config-sync in-place data back to the default folder")
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

// domReady is the wails OnDomReady hook. Native window calls (like the theme
// setter) need the webview up first, so the initial theme is applied here
// rather than in startup. It can run concurrently with startup (a large
// mailbox can still be opening when the webview signals dom-ready), so it
// waits for storeReady first to avoid racing on store.
func (a *App) domReady(ctx context.Context) {
	<-a.storeReady
	a.applyNativeTheme(a.stringSetting(storage.SettingTheme, defaultTheme))
}

// shutdown is the wails OnShutdown hook. It closes the store so the sqlite wal
// is checkpointed cleanly.
func (a *App) shutdown(ctx context.Context) {
	a.stopTray()
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
