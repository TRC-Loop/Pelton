// shortcuts.ts defines the app-wide keyboard shortcuts and the helpers to parse,
// match and record key combos. combos use "mod" for the platform primary
// modifier (cmd on macos, ctrl elsewhere), plus optional "alt" and "shift", then
// a key, e.g. "mod+n", "mod+shift+a", "alt+s". the user can rebind any of these
// in settings; the live bindings live in stores/shortcuts.ts, while this module
// stays a pure, store-free library so it can be unit-reasoned and imported
// anywhere.

import { isMac } from './i18n'

// ShortcutAction is the set of app-wide actions a shortcut can trigger. The
// second group are message-level actions that act on the open message; they ship
// unbound (empty default combo) so the user can assign keys to the right-click
// menu items if they want.
export type ShortcutAction =
  | 'compose'
  | 'preferences'
  | 'sync'
  | 'search'
  | 'add-mailbox'
  | 'reply'
  | 'reply-all'
  | 'forward'
  | 'mark-read'
  | 'mark-unread'
  | 'flag'
  | 'snooze'
  | 'download-offline'
  | 'delete-message'
  | 'archive'

// Shortcut pairs an action with its default combo and the label key for display.
export interface Shortcut {
  action: ShortcutAction
  combo: string
  labelKey: string
}

// the default registry, also used to seed the editable bindings and render the
// shortcuts list in settings.
export const shortcuts: Shortcut[] = [
  { action: 'compose', combo: 'mod+n', labelKey: 'shortcut.compose' },
  { action: 'preferences', combo: 'mod+,', labelKey: 'shortcut.preferences' },
  { action: 'sync', combo: 'mod+r', labelKey: 'shortcut.sync' },
  { action: 'add-mailbox', combo: 'mod+m', labelKey: 'shortcut.addMailbox' },
  { action: 'search', combo: 'mod+f', labelKey: 'shortcut.search' },
  // message-level actions, unbound by default.
  { action: 'reply', combo: '', labelKey: 'shortcut.reply' },
  { action: 'reply-all', combo: '', labelKey: 'shortcut.replyAll' },
  { action: 'forward', combo: '', labelKey: 'shortcut.forward' },
  { action: 'mark-read', combo: '', labelKey: 'shortcut.markRead' },
  { action: 'mark-unread', combo: '', labelKey: 'shortcut.markUnread' },
  { action: 'flag', combo: '', labelKey: 'shortcut.flag' },
  { action: 'snooze', combo: '', labelKey: 'shortcut.snooze' },
  { action: 'download-offline', combo: '', labelKey: 'shortcut.downloadOffline' },
  { action: 'delete-message', combo: '', labelKey: 'shortcut.deleteMessage' },
  { action: 'archive', combo: '', labelKey: 'shortcut.archive' },
]

// ParsedCombo is a combo broken into its modifier flags and final key.
export interface ParsedCombo {
  mod: boolean
  alt: boolean
  shift: boolean
  key: string
}

// parseCombo splits a combo string into modifier flags and the final key.
export function parseCombo(combo: string): ParsedCombo {
  const parts = combo.toLowerCase().split('+')
  const key = parts[parts.length - 1] ?? ''
  return {
    mod: parts.includes('mod'),
    alt: parts.includes('alt'),
    shift: parts.includes('shift'),
    key,
  }
}

// isModifierKey reports whether a key event is a bare modifier press (so the
// recorder waits for a real key).
function isModifierKey(key: string): boolean {
  return key === 'Shift' || key === 'Control' || key === 'Alt' || key === 'Meta'
}

// normalizeKey maps a few keys to stable combo tokens.
function normalizeKey(key: string): string {
  if (key === ' ') {
    return 'space'
  }
  return key.toLowerCase()
}

// eventToCombo builds a combo string from a keydown event, or null when only a
// modifier is held. the primary modifier is recorded as "mod" per platform.
export function eventToCombo(event: KeyboardEvent): string | null {
  if (isModifierKey(event.key)) {
    return null
  }
  const parts: string[] = []
  const mod = isMac ? event.metaKey : event.ctrlKey
  if (mod) {
    parts.push('mod')
  }
  if (event.altKey) {
    parts.push('alt')
  }
  if (event.shiftKey) {
    parts.push('shift')
  }
  parts.push(normalizeKey(event.key))
  return parts.join('+')
}

// comboHasModifier reports whether a combo includes any modifier, so callers can
// avoid firing modifier-less shortcuts while the user is typing in a field.
export function comboHasModifier(combo: string): boolean {
  const p = parseCombo(combo)
  return p.mod || p.alt || p.shift
}

// comboMatches reports whether a keydown event exactly matches a combo. the match
// is strict on every modifier so, for example, cmd+n never also fires plain n,
// and ctrl on macos (a distinct modifier we do not bind) never matches.
export function comboMatches(event: KeyboardEvent, combo: string): boolean {
  const p = parseCombo(combo)
  const primary = isMac ? event.metaKey : event.ctrlKey
  if (primary !== p.mod) {
    return false
  }
  // reject the non-primary platform modifier so combos stay unambiguous.
  if (isMac ? event.ctrlKey : event.metaKey) {
    return false
  }
  if (event.altKey !== p.alt || event.shiftKey !== p.shift) {
    return false
  }
  return normalizeKey(event.key) === p.key
}

// matchShortcut returns the action whose bound combo matches the event, or null.
// bindings maps each action to its current combo (defaults overlaid with the
// user's overrides).
export function matchShortcut(
  event: KeyboardEvent,
  bindings: Record<string, string>,
): ShortcutAction | null {
  for (const action of Object.keys(bindings)) {
    const combo = bindings[action]
    // an empty combo is an unbound action; never match it.
    if (!combo) {
      continue
    }
    if (comboMatches(event, combo)) {
      return action as ShortcutAction
    }
  }
  return null
}
