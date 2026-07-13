package desktop

import (
	"errors"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

// ui setting keys. theme and editor mode already exist in storage; the rest are
// ui-only and defined here. raw strings keep the contract in one place next to
// the defaults.
const (
	settingAccent        = "accent"
	settingDensity       = "density"
	settingShowBadge     = "show_mailbox_badge"
	settingShowDateTime  = "show_datetime"
	settingShowPGP       = "show_pgp"
	settingShowAuth      = "show_auth"
	settingToastPosition = "toast_position"
	settingPaneLocked    = "pane_locked"
	settingSidebarWidth  = "sidebar_width"
	settingListWidth     = "list_width"
	settingSendDelay     = "send_delay_seconds"
	settingFlagHighlight = "flag_highlight"
	settingShortcutHints = "show_shortcut_hints"
	settingAccountEmail  = "show_account_email"
	settingRemoteAlways  = "remote_images_always"
	settingAvatarSource  = "avatar_source"
	settingAvatarStyle   = "avatar_style"
	settingMultiSelect   = "multi_select_enabled"
	settingSelectedCount = "show_selected_count"
	settingIndentGuides  = "sidebar_indent_guides"
	settingRowTemplate   = "row_template"
	settingRowAvatar     = "row_show_avatar"
	settingRowSnippet    = "row_show_snippet"
	settingPreviewLines  = "preview_lines"
	settingUIScale       = "ui_scale"
	settingMessageFont   = "message_font_size"
	settingFlaggedCount  = "show_flagged_count"
	// newer feature settings.
	settingFlagColorSync       = "flag_color_sync"
	settingShowOffline         = "show_offline_indicator"
	settingSwipeEnabled        = "swipe_enabled"
	settingSwipeLeft           = "swipe_left_action"
	settingSwipeRight          = "swipe_right_action"
	settingVimMode             = "compose_vim_mode"
	settingDownloadAtts        = "download_include_attachments"
	settingDownloadPending     = "download_pending_since"
	settingAppVimMode          = "app_vim_mode"
	settingLanguage            = "language"
	settingLowPower            = "low_power_mode"
	settingAutoSync            = "auto_sync_interval_seconds"
	settingDefaultEditor       = "default_editor_mode"
	settingComposeAutocomplete = "compose_autocomplete"
	settingComposeChips        = "compose_chips"
	settingEmptyStateImage     = "empty_state_image"
	// menu bar settings only matter on macOS: Windows/Linux always use the
	// in-app bar (no native menu is created there at all).
	settingMenuBarInApp         = "menu_bar_in_app"
	settingMenuBarNativeMinimal = "menu_bar_native_minimal"
	// icons in the in-app menu bar's dropdowns; off keeps the classic
	// text-only native look.
	settingMenuBarIcons = "menu_bar_icons"
	// clock preference for rendered times: auto (locale), 12, or 24.
	settingTimeFormat = "time_format"
	// disable ui transitions and animations (the os-level preference is
	// honored by the frontend css regardless).
	settingReduceMotion = "reduce_motion"
	// dark window bounds ("HH:MM") for the schedule theme mode.
	settingThemeDarkStart = "theme_dark_start"
	settingThemeDarkEnd   = "theme_dark_end"
	// reader fallback font for mail bodies (a key from the frontend's curated
	// list; mail that declares its own fonts keeps them).
	settingBodyFont = "body_font"
	// interface and monospace fonts (#58): keys from the frontend's curated
	// lists or "sys:<family>" for an installed font, overriding the --font-ui
	// and --font-mono tokens.
	settingUIFont   = "ui_font"
	settingMonoFont = "mono_font"
)

// settingUpdateCheckFreq, settingLastUpdateCheck and defaultUpdateCheckFrequency
// are defined in bind_update.go, next to the rest of the update-check logic.

// defaults for the ui preferences, applied server side so the frontend always
// receives a complete object on startup.
const (
	defaultTheme         = "system"
	defaultAccent        = "#465AF2"
	defaultDensity       = "medium"
	defaultToastPosition = "bottom-right"
	defaultSidebarWidth  = 264
	defaultListWidth     = 380
	// flag highlight styles: flag (icon only), left, both, off.
	defaultFlagHighlight = "flag"
	// avatar source preference (fallback chain): bimi_gravatar, gravatar_bimi, or
	// pfp (generated only, no network).
	defaultAvatarSource = "bimi_gravatar"
	// generated placeholder style: initials, mono, pixel, geometric.
	defaultAvatarStyle = "initials"
	// list row template: relaxed (avatar + 3 lines), comfortable (3 lines),
	// compact (2 lines), single (1 line).
	defaultRowTemplate = "relaxed"
	// how many snippet preview lines to show where the template allows it.
	defaultPreviewLines = 1
	// interface zoom factor as a string multiplier ("1" = 100%).
	defaultUIScale = "1"
	// base font size (px) for rendered email content.
	defaultMessageFont = 14
)

// UIPrefsDTO is the complete set of user-facing preferences this step exposes:
// theme, accent, density and the per-row technical-info toggles. Other settings
// come later.
type UIPrefsDTO struct {
	Theme            string `json:"theme"`
	Accent           string `json:"accent"`
	Density          string `json:"density"`
	ShowMailboxBadge bool   `json:"showMailboxBadge"`
	ShowDateTime     bool   `json:"showDateTime"`
	ShowPGP          bool   `json:"showPgp"`
	ShowAuth         bool   `json:"showAuth"`
	// layout and chrome preferences.
	ToastPosition string `json:"toastPosition"`
	PaneLocked    bool   `json:"paneLocked"`
	SidebarWidth  int    `json:"sidebarWidth"`
	ListWidth     int    `json:"listWidth"`
	// SendDelaySeconds holds an outgoing message for this many seconds so the user
	// can undo. 0 disables the delay (send immediately).
	SendDelaySeconds int `json:"sendDelaySeconds"`
	// FlagHighlight controls how flagged messages stand out in the list:
	// flag (icon only), left, right, both (colored edge bars), or off.
	FlagHighlight string `json:"flagHighlight"`
	// ShowShortcutHints shows inline keyboard shortcut chips in the ui. Off by
	// default to keep the interface clean.
	ShowShortcutHints bool `json:"showShortcutHints"`
	// ShowAccountEmail shows the account email instead of its display name in the
	// sidebar account header.
	ShowAccountEmail bool `json:"showAccountEmail"`
	// AlwaysLoadImages disables remote-image blocking globally. Off by default;
	// the ui guards turning it on with a tracking warning.
	AlwaysLoadImages bool `json:"alwaysLoadImages"`
	// AvatarSource selects the sender-photo fallback chain: bimi_gravatar,
	// gravatar_bimi, or pfp (generated only). AvatarStyle picks the generated
	// placeholder look: initials, mono, pixel, or geometric.
	AvatarSource string `json:"avatarSource"`
	AvatarStyle  string `json:"avatarStyle"`
	// MultiSelectEnabled allows selecting several messages at once for bulk
	// actions. ShowSelectedCount toggles the "N selected" count text in the
	// selection bar; both are independent so the count can be hidden while
	// multi-select stays on.
	MultiSelectEnabled bool `json:"multiSelectEnabled"`
	ShowSelectedCount  bool `json:"showSelectedCount"`
	// SidebarIndentGuides draws vertical guide lines for nested folders.
	SidebarIndentGuides bool `json:"sidebarIndentGuides"`
	// RowTemplate selects the message-list row layout: relaxed, comfortable,
	// compact, or single. RowShowAvatar/RowShowSnippet are per-field overrides and
	// PreviewLines clamps the snippet to that many lines.
	RowTemplate    string `json:"rowTemplate"`
	RowShowAvatar  bool   `json:"rowShowAvatar"`
	RowShowSnippet bool   `json:"rowShowSnippet"`
	PreviewLines   int    `json:"previewLines"`
	// UIScale zooms the whole interface (a string multiplier, "1" = 100%).
	// MessageFontSize sets the base font size in px for rendered email content.
	UIScale         string `json:"uiScale"`
	MessageFontSize int    `json:"messageFontSize"`
	// ShowFlaggedCount shows the count and bold styling on the sidebar Flagged
	// view. Off keeps the entry but renders it plain.
	ShowFlaggedCount bool `json:"showFlaggedCount"`
	// FlagColorSync pushes color labels to the server as imap keywords so they
	// show in other clients. Off keeps colors local only.
	FlagColorSync bool `json:"flagColorSync"`
	// ShowOfflineIndicator shows the little downloaded/offline badge on pinned
	// messages. On by default; can be hidden.
	ShowOfflineIndicator bool `json:"showOfflineIndicator"`
	// Swipe gestures on message rows (trackpad only). SwipeEnabled turns them on;
	// SwipeLeftAction/SwipeRightAction pick what each direction does
	// (delete, unread, read, flag, archive, snooze, none).
	SwipeEnabled     bool   `json:"swipeEnabled"`
	SwipeLeftAction  string `json:"swipeLeftAction"`
	SwipeRightAction string `json:"swipeRightAction"`
	// ComposeVimMode enables vim keybindings in the compose editor.
	ComposeVimMode bool `json:"composeVimMode"`
	// DownloadIncludeAttachments is the remembered default for the bulk range
	// download's per-run attachment choice.
	DownloadIncludeAttachments bool `json:"downloadIncludeAttachments"`
	// AppVimMode enables global vim-style navigation (h/j/k/l and friends) for
	// moving around the app window itself, outside of compose.
	AppVimMode bool `json:"appVimMode"`
	// Language is the ui locale code (en, de, fr, nl, es). Defaults to English;
	// the frontend only ever picks something else on an explicit user choice.
	Language string `json:"language"`
	// LowPowerMode pauses periodic auto-sync, blocks starting new bulk offline
	// downloads, and skips the post-sync address-book rescan.
	LowPowerMode bool `json:"lowPowerMode"`
	// AutoSyncIntervalSeconds is how often every account gets a full sync pass,
	// on top of the always-on imap idle push. 0 disables it.
	AutoSyncIntervalSeconds int `json:"autoSyncIntervalSeconds"`
	// DefaultEditorMode is the editor a new compose session starts in:
	// plaintext, markdown, or wysiwyg.
	DefaultEditorMode string `json:"defaultEditorMode"`
	// ComposeAutocomplete offers address-book suggestions while typing a
	// recipient. On by default.
	ComposeAutocomplete bool `json:"composeAutocomplete"`
	// ComposeChips renders recipients as removable chips. When off, the
	// recipient fields fall back to a plain comma-separated text input.
	ComposeChips bool `json:"composeChips"`
	// UpdateCheckFrequency controls the automatic GitHub-releases update
	// check: off (default), startup (every launch), weekly, or monthly. A
	// manual check ("Check now" in settings) always runs regardless.
	UpdateCheckFrequency string `json:"updateCheckFrequency"`
	// EmptyStateImage is a data-uri image shown in the reading pane when no
	// message is open. Empty means the bundled Pelton logo.
	EmptyStateImage string `json:"emptyStateImage"`
	// ThemeID selects an installed custom theme (see bind_themes.go). Empty
	// means the built-in default themes driven by the Theme setting.
	ThemeID string `json:"themeId"`
	// MenuBarInApp shows the in-app menu bar on macOS (it is always shown on
	// Windows/Linux regardless of this). MenuBarNativeMinimal then reduces the
	// native macOS menu to the app menu, dropping the duplicated submenus.
	MenuBarInApp         bool `json:"menuBarInApp"`
	MenuBarNativeMinimal bool `json:"menuBarNativeMinimal"`
	// MenuBarIcons shows icons next to the in-app menu bar's dropdown items.
	MenuBarIcons bool `json:"menuBarIcons"`
	// TimeFormat picks the clock for rendered times: auto (locale), 12, or 24.
	TimeFormat string `json:"timeFormat"`
	// ReduceMotion disables ui transitions and animations.
	ReduceMotion bool `json:"reduceMotion"`
	// ThemeDarkStart/ThemeDarkEnd bound the dark window ("HH:MM") for the
	// schedule theme mode.
	ThemeDarkStart string `json:"themeDarkStart"`
	ThemeDarkEnd   string `json:"themeDarkEnd"`
	// BodyFont is the reader fallback font for mail bodies.
	BodyFont string `json:"bodyFont"`
	// UIFont and MonoFont override the interface and monospace font tokens.
	UIFont   string `json:"uiFont"`
	MonoFont string `json:"monoFont"`
}

// GetUIPrefs returns all ui preferences with defaults filled in, so startup is a
// single round trip and the frontend never has to know the default values.
func (a *App) GetUIPrefs() (UIPrefsDTO, error) {
	if err := a.ready(); err != nil {
		return UIPrefsDTO{}, err
	}
	return UIPrefsDTO{
		Theme:               a.stringSetting(storage.SettingTheme, defaultTheme),
		Accent:              a.stringSetting(settingAccent, defaultAccent),
		Density:             a.stringSetting(settingDensity, defaultDensity),
		ShowMailboxBadge:    a.boolSetting(settingShowBadge, true),
		ShowDateTime:        a.boolSetting(settingShowDateTime, true),
		ShowPGP:             a.boolSetting(settingShowPGP, true),
		ShowAuth:            a.boolSetting(settingShowAuth, true),
		ToastPosition:       a.stringSetting(settingToastPosition, defaultToastPosition),
		PaneLocked:          a.boolSetting(settingPaneLocked, false),
		SidebarWidth:        a.intSetting(settingSidebarWidth, defaultSidebarWidth),
		ListWidth:           a.intSetting(settingListWidth, defaultListWidth),
		SendDelaySeconds:    a.intSetting(settingSendDelay, 0),
		FlagHighlight:       a.stringSetting(settingFlagHighlight, defaultFlagHighlight),
		ShowShortcutHints:   a.boolSetting(settingShortcutHints, false),
		ShowAccountEmail:    a.boolSetting(settingAccountEmail, false),
		AlwaysLoadImages:    a.boolSetting(settingRemoteAlways, false),
		AvatarSource:        a.stringSetting(settingAvatarSource, defaultAvatarSource),
		AvatarStyle:         a.stringSetting(settingAvatarStyle, defaultAvatarStyle),
		MultiSelectEnabled:  a.boolSetting(settingMultiSelect, true),
		ShowSelectedCount:   a.boolSetting(settingSelectedCount, true),
		SidebarIndentGuides: a.boolSetting(settingIndentGuides, false),
		RowTemplate:         a.stringSetting(settingRowTemplate, defaultRowTemplate),
		RowShowAvatar:       a.boolSetting(settingRowAvatar, true),
		RowShowSnippet:      a.boolSetting(settingRowSnippet, true),
		PreviewLines:        a.intSetting(settingPreviewLines, defaultPreviewLines),
		UIScale:             a.stringSetting(settingUIScale, defaultUIScale),
		MessageFontSize:     a.intSetting(settingMessageFont, defaultMessageFont),
		ShowFlaggedCount:    a.boolSetting(settingFlaggedCount, true),

		FlagColorSync:              a.boolSetting(settingFlagColorSync, false),
		ShowOfflineIndicator:       a.boolSetting(settingShowOffline, true),
		SwipeEnabled:               a.boolSetting(settingSwipeEnabled, true),
		SwipeLeftAction:            a.stringSetting(settingSwipeLeft, "delete"),
		SwipeRightAction:           a.stringSetting(settingSwipeRight, "unread"),
		ComposeVimMode:             a.boolSetting(settingVimMode, false),
		DownloadIncludeAttachments: a.boolSetting(settingDownloadAtts, true),
		AppVimMode:                 a.boolSetting(settingAppVimMode, false),
		Language:                   a.stringSetting(settingLanguage, "en"),
		LowPowerMode:               a.boolSetting(settingLowPower, false),
		AutoSyncIntervalSeconds:    a.intSetting(settingAutoSync, 900),
		DefaultEditorMode:          a.stringSetting(settingDefaultEditor, "plaintext"),
		ComposeAutocomplete:        a.boolSetting(settingComposeAutocomplete, true),
		ComposeChips:               a.boolSetting(settingComposeChips, true),
		UpdateCheckFrequency:       a.stringSetting(settingUpdateCheckFreq, defaultUpdateCheckFrequency),
		EmptyStateImage:            a.stringSetting(settingEmptyStateImage, ""),
		ThemeID:                    a.stringSetting(settingThemeID, ""),
		MenuBarInApp:               a.boolSetting(settingMenuBarInApp, false),
		MenuBarNativeMinimal:       a.boolSetting(settingMenuBarNativeMinimal, false),
		MenuBarIcons:               a.boolSetting(settingMenuBarIcons, false),
		TimeFormat:                 a.stringSetting(settingTimeFormat, "auto"),
		ReduceMotion:               a.boolSetting(settingReduceMotion, false),
		ThemeDarkStart:             a.stringSetting(settingThemeDarkStart, "19:00"),
		ThemeDarkEnd:               a.stringSetting(settingThemeDarkEnd, "07:00"),
		BodyFont:                   a.stringSetting(settingBodyFont, "default"),
		UIFont:                     a.stringSetting(settingUIFont, "default"),
		MonoFont:                   a.stringSetting(settingMonoFont, "default"),
	}, nil
}

// GetSetting returns a single setting's raw value. found is false when the key
// has never been written, so the frontend can fall back to its own default.
func (a *App) GetSetting(key string) (SettingResult, error) {
	if err := a.ready(); err != nil {
		return SettingResult{}, err
	}
	value, err := a.store.Get(a.ctx, key)
	if isSettingMissing(err) {
		return SettingResult{Found: false}, nil
	}
	if err != nil {
		return SettingResult{}, err
	}
	return SettingResult{Value: value, Found: true}, nil
}

// SetSetting writes a single setting. The frontend uses the known keys above
// (theme/accent/density/show_*) plus editor_mode.
func (a *App) SetSetting(key, value string) error {
	if err := a.ready(); err != nil {
		return err
	}
	if err := a.store.Set(a.ctx, key, value); err != nil {
		return err
	}
	if key == storage.SettingTheme {
		a.applyNativeTheme(value)
	}
	if key == settingLanguage || key == settingMenuBarInApp || key == settingMenuBarNativeMinimal {
		a.RebuildMenu()
	}
	return nil
}

// applyNativeTheme pushes the app's theme choice onto the native window
// chrome (title bar/menu strip on Windows; no-op on macOS/Linux).
func (a *App) applyNativeTheme(theme string) {
	switch theme {
	case "dark":
		wailsruntime.WindowSetDarkTheme(a.ctx)
	case "light":
		wailsruntime.WindowSetLightTheme(a.ctx)
	default:
		wailsruntime.WindowSetSystemDefaultTheme(a.ctx)
	}
}

// SettingResult is a setting lookup that distinguishes unset from empty.
type SettingResult struct {
	Value string `json:"value"`
	Found bool   `json:"found"`
}

// stringSetting reads a string setting, returning def when unset, on error, or
// if the store hasn't opened yet (domReady can fire before startup finishes
// opening a large store, see app.go) so startup never fails on a missing
// preference.
func (a *App) stringSetting(key, def string) string {
	if a.store == nil {
		return def
	}
	value, err := a.store.Get(a.ctx, key)
	if err != nil || value == "" {
		return def
	}
	return value
}

// boolSetting reads a bool setting, returning def when unset, unparsable, or
// if the store hasn't opened yet.
func (a *App) boolSetting(key string, def bool) bool {
	if a.store == nil {
		return def
	}
	value, err := a.store.GetBool(a.ctx, key)
	if err != nil {
		return def
	}
	return value
}

// intSetting reads an int setting, returning def when unset, unparsable, or if
// the store hasn't opened yet.
func (a *App) intSetting(key string, def int) int {
	if a.store == nil {
		return def
	}
	value, err := a.store.GetInt(a.ctx, key)
	if err != nil {
		return def
	}
	return value
}

// isSettingMissing reports whether err is the not-found sentinel, so callers can
// treat an unset key as "use the default" rather than an error.
func isSettingMissing(err error) bool {
	return errors.Is(err, storage.ErrSettingNotFound)
}

// lowPowerMode reports the current low-power setting, read fresh each call
// (a local sqlite read) so background loops always see live changes without
// needing their own change-notification plumbing.
func (a *App) lowPowerMode() bool {
	return a.boolSetting(settingLowPower, false)
}
