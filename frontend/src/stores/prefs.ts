// prefs.ts is the single store for user preferences (theme, accent, density and
// the per-row technical-info toggles). it loads them from the backend on startup,
// applies them to the dom, and persists every change back through the settings
// bindings. nothing here uses localStorage: the backend settings table is the
// source of truth.

import { writable } from 'svelte/store'
import type { UIPrefs, ThemePref, DensityPref, EditorMode } from '../lib/types'
import { getUIPrefs, setSetting, SettingKeys, systemColorScheme, setWindowTheme, getThemeApply } from '../lib/api'
import { applyTheme, applyDensity, applyAccent, applyScale, applyCorners, watchSystemTheme, setSystemSchemeOverride, resolveTheme } from '../theme/theme'
import { applyUserTheme } from '../theme/usertheme'
import { setLocale, type Locale } from '../lib/i18n'

// defaults match the backend defaults so the ui renders sanely even before the
// first load resolves.
const defaults: UIPrefs = {
  theme: 'system',
  accent: '#465AF2',
  density: 'medium',
  showMailboxBadge: true,
  showDateTime: true,
  showPgp: true,
  showAuth: true,
  toastPosition: 'bottom-right',
  paneLocked: false,
  sidebarWidth: 264,
  listWidth: 380,
  sendDelaySeconds: 0,
  flagHighlight: 'flag',
  showShortcutHints: false,
  showAccountEmail: false,
  alwaysLoadImages: false,
  avatarSource: 'bimi_gravatar',
  avatarStyle: 'initials',
  multiSelectEnabled: true,
  showSelectedCount: true,
  sidebarIndentGuides: false,
  rowTemplate: 'relaxed',
  rowShowAvatar: true,
  rowShowSnippet: true,
  previewLines: 1,
  uiScale: '1',
  messageFontSize: 14,
  showFlaggedCount: true,
  flagColorSync: false,
  showOfflineIndicator: true,
  swipeEnabled: true,
  swipeLeftAction: 'delete',
  swipeRightAction: 'unread',
  composeVimMode: false,
  downloadIncludeAttachments: true,
  appVimMode: false,
  language: 'en',
  lowPowerMode: false,
  autoSyncIntervalSeconds: 900,
  defaultEditorMode: 'plaintext',
  composeAutocomplete: true,
  composeChips: true,
  updateCheckFrequency: 'off',
  emptyStateImage: '',
  cornerStyle: 'default',
  themeId: '',
}

export const prefs = writable<UIPrefs>(defaults)

// syncWindowChrome matches the native window chrome (the Windows caption bar) to
// the resolved theme so it does not stay light under a dark ui.
function syncWindowChrome(pref: ThemePref): void {
  setWindowTheme(resolveTheme(pref) === 'dark')
}

// applyAll pushes the current preferences onto the document.
function applyAll(p: UIPrefs): void {
  applyTheme(p.theme as ThemePref)
  syncWindowChrome(p.theme as ThemePref)
  applyDensity(p.density as DensityPref)
  applyAccent(p.accent)
  applyScale(p.uiScale)
  applyCorners(p.cornerStyle)
  setLocale(p.language as Locale)
}

// initPrefs loads preferences, applies them, and keeps the theme in sync with
// the os while in system mode. call once at startup.
export async function initPrefs(): Promise<void> {
  const loaded = await getUIPrefs()
  prefs.set(loaded)

  // on Linux the css prefers-color-scheme query never reports the desktop dark
  // preference, so resolve it natively first; elsewhere this returns "" and the
  // media query is used. done before applyAll so the first paint is correct.
  try {
    const scheme = await systemColorScheme()
    if (scheme === 'dark' || scheme === 'light') {
      setSystemSchemeOverride(scheme)
    }
  } catch {
    // fall back to the media query
  }

  applyAll(loaded)

  // a persisted custom theme applies after the base prefs so its base wins.
  // a missing or broken theme folder falls back to the default silently; the
  // stored selection stays, so reinstalling the theme brings it back.
  if (loaded.themeId) {
    try {
      applyUserTheme(await getThemeApply(loaded.themeId))
    } catch {
      applyUserTheme(null)
    }
  }

  watchSystemTheme(() => {
    let current: UIPrefs = defaults
    prefs.subscribe((p) => (current = p))()
    // a custom theme pins its own base; the os scheme only matters without one.
    if (current.theme === 'system' && !current.themeId) {
      applyTheme('system')
      syncWindowChrome('system')
    }
  })
}

// the setters below update the store, apply the change immediately, and persist
// it. they are fire-and-forget for persistence; a failed write only means the
// choice will not survive a restart, not that the ui should block.

export function setTheme(theme: ThemePref): void {
  prefs.update((p) => ({ ...p, theme }))
  applyTheme(theme)
  syncWindowChrome(theme)
  void setSetting(SettingKeys.theme, theme)
}

// setThemeId activates an installed custom theme ('' returns to the built-in
// default). It loads and applies the theme before persisting, so a broken
// theme folder surfaces as a rejection here instead of a half-applied ui.
export async function setThemeId(themeId: string): Promise<void> {
  if (themeId) {
    applyUserTheme(await getThemeApply(themeId))
    syncWindowChrome(themeIdBase())
  } else {
    applyUserTheme(null)
    let current: UIPrefs = defaults
    prefs.subscribe((p) => (current = p))()
    applyTheme(current.theme as ThemePref)
    syncWindowChrome(current.theme as ThemePref)
  }
  prefs.update((p) => ({ ...p, themeId }))
  void setSetting(SettingKeys.themeId, themeId)
}

// themeIdBase reads the base the injected theme pinned on the root element,
// so the native window chrome can follow it.
function themeIdBase(): 'light' | 'dark' {
  return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'light'
}

// setCornerStyle picks the corner radius look and applies it immediately.
export function setCornerStyle(value: string): void {
  prefs.update((p) => ({ ...p, cornerStyle: value }))
  applyCorners(value)
  void setSetting(SettingKeys.cornerStyle, value)
}

export function setDensity(density: DensityPref): void {
  prefs.update((p) => ({ ...p, density }))
  applyDensity(density)
  void setSetting(SettingKeys.density, density)
}

// setUIScale zooms the whole interface and persists the multiplier.
export function setUIScale(scale: string): void {
  prefs.update((p) => ({ ...p, uiScale: scale }))
  applyScale(scale)
  void setSetting(SettingKeys.uiScale, scale)
}

// setMessageFontSize sets the base font size (px) for rendered email content.
export function setMessageFontSize(size: number): void {
  prefs.update((p) => ({ ...p, messageFontSize: size }))
  void setSetting(SettingKeys.messageFontSize, String(size))
}

// setShowFlaggedCount toggles the count and bold styling on the sidebar Flagged
// view (the entry itself always stays).
export function setShowFlaggedCount(value: boolean): void {
  prefs.update((p) => ({ ...p, showFlaggedCount: value }))
  void setSetting(SettingKeys.showFlaggedCount, String(value))
}

// setFlagColorSync toggles pushing color labels to the server as imap keywords.
export function setFlagColorSync(value: boolean): void {
  prefs.update((p) => ({ ...p, flagColorSync: value }))
  void setSetting(SettingKeys.flagColorSync, String(value))
}

// setShowOfflineIndicator toggles the little downloaded badge on pinned messages.
export function setShowOfflineIndicator(value: boolean): void {
  prefs.update((p) => ({ ...p, showOfflineIndicator: value }))
  void setSetting(SettingKeys.showOfflineIndicator, String(value))
}

// setSwipeEnabled toggles trackpad swipe gestures on message rows.
export function setSwipeEnabled(value: boolean): void {
  prefs.update((p) => ({ ...p, swipeEnabled: value }))
  void setSetting(SettingKeys.swipeEnabled, String(value))
}

// setSwipeLeftAction / setSwipeRightAction pick what each swipe direction does.
export function setSwipeLeftAction(action: string): void {
  prefs.update((p) => ({ ...p, swipeLeftAction: action }))
  void setSetting(SettingKeys.swipeLeftAction, action)
}

export function setSwipeRightAction(action: string): void {
  prefs.update((p) => ({ ...p, swipeRightAction: action }))
  void setSetting(SettingKeys.swipeRightAction, action)
}

// setComposeVimMode toggles vim keybindings in the compose editor.
export function setComposeVimMode(value: boolean): void {
  prefs.update((p) => ({ ...p, composeVimMode: value }))
  void setSetting(SettingKeys.composeVimMode, String(value))
}

// setDownloadIncludeAttachments remembers the range-download attachment choice.
export function setDownloadIncludeAttachments(value: boolean): void {
  prefs.update((p) => ({ ...p, downloadIncludeAttachments: value }))
  void setSetting(SettingKeys.downloadIncludeAttachments, String(value))
}

// setAppVimMode toggles global vim-style navigation of the app window.
export function setAppVimMode(value: boolean): void {
  prefs.update((p) => ({ ...p, appVimMode: value }))
  void setSetting(SettingKeys.appVimMode, String(value))
}

// setLanguage persists the chosen ui locale and applies it immediately.
export function setLanguage(language: Locale): void {
  prefs.update((p) => ({ ...p, language }))
  setLocale(language)
  void setSetting(SettingKeys.language, language)
}

// setLowPowerMode toggles pausing periodic auto-sync, bulk downloads and
// address-book rescans.
export function setLowPowerMode(value: boolean): void {
  prefs.update((p) => ({ ...p, lowPowerMode: value }))
  void setSetting(SettingKeys.lowPowerMode, String(value))
}

// setDefaultEditorMode sets the editor a new compose session starts in.
export function setDefaultEditorMode(mode: EditorMode): void {
  prefs.update((p) => ({ ...p, defaultEditorMode: mode }))
  void setSetting(SettingKeys.defaultEditorMode, mode)
}

// setComposeAutocomplete toggles address-book suggestions while typing a
// recipient.
export function setComposeAutocomplete(value: boolean): void {
  prefs.update((p) => ({ ...p, composeAutocomplete: value }))
  void setSetting(SettingKeys.composeAutocomplete, String(value))
}

// setComposeChips toggles rendering recipients as removable chips versus a
// plain comma-separated text field.
export function setComposeChips(value: boolean): void {
  prefs.update((p) => ({ ...p, composeChips: value }))
  void setSetting(SettingKeys.composeChips, String(value))
}

// setUpdateCheckFrequency persists how often the app checks GitHub releases
// for a newer version: 'off', 'startup', 'weekly', or 'monthly'.
export function setUpdateCheckFrequency(value: string): void {
  prefs.update((p) => ({ ...p, updateCheckFrequency: value }))
  void setSetting(SettingKeys.updateCheckFrequency, value)
}

// setEmptyStateImage persists the reading-pane empty-state image as a data uri;
// an empty string restores the bundled Pelton logo.
export function setEmptyStateImage(value: string): void {
  prefs.update((p) => ({ ...p, emptyStateImage: value }))
  void setSetting(SettingKeys.emptyStateImage, value)
}

// setAutoSyncInterval persists how often a full sync pass runs, in seconds (0
// disables it).
export function setAutoSyncInterval(seconds: number): void {
  prefs.update((p) => ({ ...p, autoSyncIntervalSeconds: seconds }))
  void setSetting(SettingKeys.autoSyncIntervalSeconds, String(seconds))
}

export function setAccent(accent: string): void {
  prefs.update((p) => ({ ...p, accent }))
  applyAccent(accent)
  void setSetting(SettingKeys.accent, accent)
}

// toggle keys map a boolean preference to its setting key so setToggle stays
// generic over the four technical-info switches.
type ToggleKey = 'showMailboxBadge' | 'showDateTime' | 'showPgp' | 'showAuth'

const toggleSettingKey: Record<ToggleKey, string> = {
  showMailboxBadge: SettingKeys.showMailboxBadge,
  showDateTime: SettingKeys.showDateTime,
  showPgp: SettingKeys.showPgp,
  showAuth: SettingKeys.showAuth,
}

export function setToggle(key: ToggleKey, value: boolean): void {
  prefs.update((p) => ({ ...p, [key]: value }))
  void setSetting(toggleSettingKey[key], String(value))
}

export function setToastPosition(position: string): void {
  prefs.update((p) => ({ ...p, toastPosition: position }))
  void setSetting(SettingKeys.toastPosition, position)
}

// setSendDelay persists the undo-send window in seconds (0 disables it).
export function setSendDelay(seconds: number): void {
  prefs.update((p) => ({ ...p, sendDelaySeconds: seconds }))
  void setSetting(SettingKeys.sendDelay, String(seconds))
}

// setFlagHighlight persists how flagged rows are highlighted.
export function setFlagHighlight(style: string): void {
  prefs.update((p) => ({ ...p, flagHighlight: style }))
  void setSetting(SettingKeys.flagHighlight, style)
}

// setShortcutHints toggles the inline keyboard shortcut chips.
export function setShortcutHints(value: boolean): void {
  prefs.update((p) => ({ ...p, showShortcutHints: value }))
  void setSetting(SettingKeys.shortcutHints, String(value))
}

// setShowAccountEmail toggles showing the account email instead of its name.
export function setShowAccountEmail(value: boolean): void {
  prefs.update((p) => ({ ...p, showAccountEmail: value }))
  void setSetting(SettingKeys.accountEmail, String(value))
}

// setAlwaysLoadImages toggles the global remote-image override. the settings ui
// guards enabling it with a tracking warning.
export function setAlwaysLoadImages(value: boolean): void {
  prefs.update((p) => ({ ...p, alwaysLoadImages: value }))
  void setSetting(SettingKeys.alwaysLoadImages, String(value))
}

// setAvatarSource selects the sender-photo fallback chain (bimi_gravatar,
// gravatar_bimi, pfp).
export function setAvatarSource(source: string): void {
  prefs.update((p) => ({ ...p, avatarSource: source }))
  void setSetting(SettingKeys.avatarSource, source)
}

// setAvatarStyle selects the generated placeholder look (initials, mono, pixel,
// geometric).
export function setAvatarStyle(style: string): void {
  prefs.update((p) => ({ ...p, avatarStyle: style }))
  void setSetting(SettingKeys.avatarStyle, style)
}

// setMultiSelectEnabled toggles whether several rows can be selected at once.
export function setMultiSelectEnabled(value: boolean): void {
  prefs.update((p) => ({ ...p, multiSelectEnabled: value }))
  void setSetting(SettingKeys.multiSelectEnabled, String(value))
}

// setShowSelectedCount toggles the "N selected" count text in the selection bar.
export function setShowSelectedCount(value: boolean): void {
  prefs.update((p) => ({ ...p, showSelectedCount: value }))
  void setSetting(SettingKeys.showSelectedCount, String(value))
}

// setSidebarIndentGuides toggles the nested-folder guide lines.
export function setSidebarIndentGuides(value: boolean): void {
  prefs.update((p) => ({ ...p, sidebarIndentGuides: value }))
  void setSetting(SettingKeys.sidebarIndentGuides, String(value))
}

// setRowTemplate selects the message-list row layout.
export function setRowTemplate(template: string): void {
  prefs.update((p) => ({ ...p, rowTemplate: template }))
  void setSetting(SettingKeys.rowTemplate, template)
}

// setRowShowAvatar / setRowShowSnippet are per-field overrides on the template.
export function setRowShowAvatar(value: boolean): void {
  prefs.update((p) => ({ ...p, rowShowAvatar: value }))
  void setSetting(SettingKeys.rowShowAvatar, String(value))
}

export function setRowShowSnippet(value: boolean): void {
  prefs.update((p) => ({ ...p, rowShowSnippet: value }))
  void setSetting(SettingKeys.rowShowSnippet, String(value))
}

// setPreviewLines clamps the snippet to 1..3 lines.
export function setPreviewLines(lines: number): void {
  const clamped = Math.max(1, Math.min(3, Math.round(lines)))
  prefs.update((p) => ({ ...p, previewLines: clamped }))
  void setSetting(SettingKeys.previewLines, String(clamped))
}

export function setPaneLocked(locked: boolean): void {
  prefs.update((p) => ({ ...p, paneLocked: locked }))
  void setSetting(SettingKeys.paneLocked, String(locked))
}

// setPaneWidths persists the two resizable column widths. it is called as the
// user finishes dragging a divider, not on every pixel, to avoid hammering the
// settings table.
export function setPaneWidths(sidebarWidth: number, listWidth: number): void {
  prefs.update((p) => ({ ...p, sidebarWidth, listWidth }))
  void setSetting(SettingKeys.sidebarWidth, String(Math.round(sidebarWidth)))
  void setSetting(SettingKeys.listWidth, String(Math.round(listWidth)))
}
