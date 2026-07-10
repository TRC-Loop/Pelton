package desktop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Backup import/export lets a user move their configuration between installs
// through a plain JSON file, chosen per category. It is the local, file-based
// replacement for the removed folder-based config sync: Pelton writes and reads
// a file the user picks, and never talks to any server.

// backupCategorySettings and backupCategoryWhitelist are the export/import
// category ids the ui checkboxes map to.
const (
	backupCategorySettings  = "settings"
	backupCategoryWhitelist = "whitelist"
)

// backupFileTag identifies a Pelton backup file so import can reject unrelated
// json.
const backupFileTag = "pelton-backup"

// backupSkipSettings are settings that must not travel between installs: the
// search watermark and the pending-download marker are local, transient state,
// and the whitelist keys are exported under their own category instead.
var backupSkipSettings = map[string]bool{
	settingSearchWatermark: true,
	settingDownloadPending: true,
	settingRemoteSenders:   true,
	settingRemoteDomains:   true,
}

// whitelistBackup is the trusted-sender allowlist as exported.
type whitelistBackup struct {
	Senders []string `json:"senders"`
	Domains []string `json:"domains"`
}

// BackupFileDTO is the on-disk backup document (and what import inspects).
type BackupFileDTO struct {
	Tag        string            `json:"tag"`
	Version    int               `json:"version"`
	CreatedAt  string            `json:"createdAt"`
	AppVersion string            `json:"appVersion"`
	Settings   map[string]string `json:"settings,omitempty"`
	Whitelist  *whitelistBackup  `json:"whitelist,omitempty"`
}

// ExportData writes the selected categories to a user-chosen json file and
// returns its path, or an empty string if the dialog was cancelled.
func (a *App) ExportData(categories []string) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	want := toSet(categories)

	doc := BackupFileDTO{
		Tag:        backupFileTag,
		Version:    1,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		AppVersion: a.version,
	}
	if want[backupCategorySettings] {
		settings, err := a.exportSettings()
		if err != nil {
			return "", err
		}
		doc.Settings = settings
	}
	if want[backupCategoryWhitelist] {
		doc.Whitelist = &whitelistBackup{Senders: a.remoteSenders(), Domains: a.remoteDomains()}
	}

	dest, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: fmt.Sprintf("pelton-backup-%s.json", time.Now().Format("2006-01-02")),
		Title:           "Export Pelton data",
	})
	if err != nil {
		return "", err
	}
	if dest == "" {
		return "", nil
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Clean(dest), data, 0o644); err != nil {
		return "", err
	}
	return dest, nil
}

// exportSettings returns every persisted setting except the local/transient and
// whitelist keys (the whitelist has its own category).
func (a *App) exportSettings() (map[string]string, error) {
	all, err := a.store.AllSettings(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(all))
	for _, s := range all {
		if backupSkipSettings[s.Key] {
			continue
		}
		out[s.Key] = s.Value
	}
	return out, nil
}

// BackupInfoDTO describes a picked backup file so the import ui can show what it
// holds (and when it was made) before the user commits to importing.
type BackupInfoDTO struct {
	Path         string `json:"path"`
	CreatedAt    string `json:"createdAt"`
	AppVersion   string `json:"appVersion"`
	HasSettings  bool   `json:"hasSettings"`
	HasWhitelist bool   `json:"hasWhitelist"`
	SettingCount int    `json:"settingCount"`
}

// InspectBackupFile opens a file picker and parses the chosen backup, returning
// what it contains. An empty Path means the dialog was cancelled.
func (a *App) InspectBackupFile() (BackupInfoDTO, error) {
	if err := a.ready(); err != nil {
		return BackupInfoDTO{}, err
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Import Pelton data",
		Filters: []runtime.FileFilter{{DisplayName: "Pelton backup (*.json)", Pattern: "*.json"}},
	})
	if err != nil {
		return BackupInfoDTO{}, err
	}
	if path == "" {
		return BackupInfoDTO{}, nil
	}
	doc, err := readBackupFile(path)
	if err != nil {
		return BackupInfoDTO{}, err
	}
	return BackupInfoDTO{
		Path:         path,
		CreatedAt:    doc.CreatedAt,
		AppVersion:   doc.AppVersion,
		HasSettings:  len(doc.Settings) > 0,
		HasWhitelist: doc.Whitelist != nil,
		SettingCount: len(doc.Settings),
	}, nil
}

// ImportData applies the selected categories from the backup file at path.
func (a *App) ImportData(path string, categories []string) error {
	if err := a.ready(); err != nil {
		return err
	}
	doc, err := readBackupFile(path)
	if err != nil {
		return err
	}
	want := toSet(categories)
	if want[backupCategorySettings] {
		for key, value := range doc.Settings {
			if backupSkipSettings[key] {
				continue
			}
			if err := a.store.Set(a.ctx, key, value); err != nil {
				return err
			}
		}
	}
	if want[backupCategoryWhitelist] && doc.Whitelist != nil {
		if err := a.store.SetJSON(a.ctx, settingRemoteSenders, doc.Whitelist.Senders); err != nil {
			return err
		}
		if err := a.store.SetJSON(a.ctx, settingRemoteDomains, doc.Whitelist.Domains); err != nil {
			return err
		}
	}
	return nil
}

// readBackupFile reads and validates a Pelton backup json file.
func readBackupFile(path string) (BackupFileDTO, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return BackupFileDTO{}, err
	}
	var doc BackupFileDTO
	if err := json.Unmarshal(data, &doc); err != nil {
		return BackupFileDTO{}, fmt.Errorf("pelton: not a valid backup file: %w", err)
	}
	if doc.Tag != backupFileTag {
		return BackupFileDTO{}, fmt.Errorf("pelton: not a Pelton backup file")
	}
	return doc, nil
}

// toSet turns a category slice into a lookup set.
func toSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, v := range values {
		set[v] = true
	}
	return set
}
