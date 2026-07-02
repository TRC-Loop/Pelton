package desktop

import (
	"errors"

	"github.com/TRC-Loop/Pelton/internal/configsync"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var errConfigSyncUnavailable = errors.New("config sync is not available")

// ConfigSyncStatusDTO mirrors configsync.Config for the frontend.
type ConfigSyncStatusDTO struct {
	Enabled      bool   `json:"enabled"`
	Mode         string `json:"mode"`
	Path         string `json:"path"`
	SyncSettings bool   `json:"syncSettings"`
	EmailScope   string `json:"emailScope"`
	LastSyncUnix int64  `json:"lastSyncUnix"`
	LastError    string `json:"lastError"`
}

func toConfigSyncDTO(cfg configsync.Config) ConfigSyncStatusDTO {
	return ConfigSyncStatusDTO{
		Enabled:      cfg.Enabled,
		Mode:         string(cfg.Mode),
		Path:         cfg.Path,
		SyncSettings: cfg.SyncSettings,
		EmailScope:   string(cfg.EmailScope),
		LastSyncUnix: cfg.LastSyncUnix,
		LastError:    cfg.LastError,
	}
}

// GetConfigSyncStatus returns the current settings-sync setup and last-run
// status, so the settings ui can render it without a separate poll.
func (a *App) GetConfigSyncStatus() (ConfigSyncStatusDTO, error) {
	if a.sync == nil {
		return ConfigSyncStatusDTO{}, errConfigSyncUnavailable
	}
	return toConfigSyncDTO(a.sync.Status()), nil
}

// ConfigureConfigSync sets up (or reconfigures) settings sync: the folder,
// the mode (copy or readonly), whether settings are included, and the email
// scope (off, metadata, or full). It runs an initial sync pass before
// returning, so setup errors (an unwritable folder, for instance) surface
// immediately in the setup modal.
func (a *App) ConfigureConfigSync(mode string, path string, syncSettings bool, emailScope string) (ConfigSyncStatusDTO, error) {
	if err := a.ready(); err != nil {
		return ConfigSyncStatusDTO{}, err
	}
	if a.sync == nil {
		return ConfigSyncStatusDTO{}, errConfigSyncUnavailable
	}
	cfg := configsync.Config{
		Mode:         configsync.Mode(mode),
		Path:         path,
		SyncSettings: syncSettings,
		EmailScope:   configsync.EmailScope(emailScope),
	}
	if err := a.sync.Configure(a.ctx, cfg); err != nil {
		return toConfigSyncDTO(a.sync.Status()), err
	}
	return toConfigSyncDTO(a.sync.Status()), nil
}

// DisableConfigSync turns settings sync off without touching the folder's
// contents, so another device still using it is unaffected.
func (a *App) DisableConfigSync() (ConfigSyncStatusDTO, error) {
	if err := a.ready(); err != nil {
		return ConfigSyncStatusDTO{}, err
	}
	if a.sync == nil {
		return ConfigSyncStatusDTO{}, errConfigSyncUnavailable
	}
	if err := a.sync.Disable(a.ctx); err != nil {
		return ConfigSyncStatusDTO{}, err
	}
	return toConfigSyncDTO(a.sync.Status()), nil
}

// TriggerConfigSync runs one manual sync pass immediately ("Sync now"),
// instead of waiting for the folder watcher to notice a change.
func (a *App) TriggerConfigSync() (ConfigSyncStatusDTO, error) {
	if err := a.ready(); err != nil {
		return ConfigSyncStatusDTO{}, err
	}
	if a.sync == nil {
		return ConfigSyncStatusDTO{}, errConfigSyncUnavailable
	}
	err := a.sync.TriggerSync(a.ctx)
	return toConfigSyncDTO(a.sync.Status()), err
}

// PickConfigSyncFolder opens a native directory picker for the sync folder,
// returning an empty string if the user cancels.
func (a *App) PickConfigSyncFolder() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Choose a sync folder"})
}
