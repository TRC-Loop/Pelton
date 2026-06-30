// prefs.ts is the single store for user preferences (theme, accent, density and
// the per-row technical-info toggles). it loads them from the backend on startup,
// applies them to the dom, and persists every change back through the settings
// bindings. nothing here uses localStorage: the backend settings table is the
// source of truth.

import { writable } from 'svelte/store'
import type { UIPrefs, ThemePref, DensityPref } from '../lib/types'
import { getUIPrefs, setSetting, SettingKeys } from '../lib/api'
import { applyTheme, applyDensity, applyAccent, applyScale, watchSystemTheme } from '../theme/theme'

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
}

export const prefs = writable<UIPrefs>(defaults)

// applyAll pushes the current preferences onto the document.
function applyAll(p: UIPrefs): void {
  applyTheme(p.theme as ThemePref)
  applyDensity(p.density as DensityPref)
  applyAccent(p.accent)
  applyScale(p.uiScale)
}

// initPrefs loads preferences, applies them, and keeps the theme in sync with
// the os while in system mode. call once at startup.
export async function initPrefs(): Promise<void> {
  const loaded = await getUIPrefs()
  prefs.set(loaded)
  applyAll(loaded)

  watchSystemTheme(() => {
    let current: UIPrefs = defaults
    prefs.subscribe((p) => (current = p))()
    if (current.theme === 'system') {
      applyTheme('system')
    }
  })
}

// the setters below update the store, apply the change immediately, and persist
// it. they are fire-and-forget for persistence; a failed write only means the
// choice will not survive a restart, not that the ui should block.

export function setTheme(theme: ThemePref): void {
  prefs.update((p) => ({ ...p, theme }))
  applyTheme(theme)
  void setSetting(SettingKeys.theme, theme)
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
