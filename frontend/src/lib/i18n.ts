// i18n.ts is a small, dependency-free localization layer. it detects the locale
// from the browser, exposes t() for looking up strings, and provides the
// platform-correct modifier symbol so keyboard shortcut hints read natively
// (cmd on macos, ctrl elsewhere). it is intentionally minimal: enough to keep
// new user-facing strings and shortcut labels out of hardcoded english, and
// ready to grow into a fuller catalog without touching call sites.

export type Locale = 'en' | 'de'

// the string catalog. keys are stable identifiers; add locales by adding a map.
// only strings introduced by the menubar, shortcuts and the newer settings are
// catalogued here for now; older components can be migrated incrementally.
const catalog: Record<Locale, Record<string, string>> = {
  en: {
    'menu.preferences': 'Preferences',
    'menu.compose': 'Compose',
    'menu.sync': 'Sync Now',
    'menu.addMailbox': 'Add Mailbox',
    'shortcut.compose': 'Compose',
    'shortcut.preferences': 'Preferences',
    'shortcut.sync': 'Sync now',
    'shortcut.search': 'Search',
    'shortcut.addMailbox': 'Add mailbox',
    'shortcut.next': 'Next message',
    'shortcut.prev': 'Previous message',
    'shortcut.open': 'Open message',
    'settings.shortcuts': 'Keyboard shortcuts',
    'settings.toastPosition': 'Notification position',
    'settings.panes': 'Panes',
    'settings.lockPanes': 'Lock pane sizes',
    'addMailbox.cta': 'Add mailbox',
  },
  de: {
    'menu.preferences': 'Einstellungen',
    'menu.compose': 'Verfassen',
    'menu.sync': 'Jetzt synchronisieren',
    'menu.addMailbox': 'Postfach hinzufügen',
    'shortcut.compose': 'Verfassen',
    'shortcut.preferences': 'Einstellungen',
    'shortcut.sync': 'Jetzt synchronisieren',
    'shortcut.search': 'Suchen',
    'shortcut.addMailbox': 'Postfach hinzufügen',
    'shortcut.next': 'Nächste Nachricht',
    'shortcut.prev': 'Vorherige Nachricht',
    'shortcut.open': 'Nachricht öffnen',
    'settings.shortcuts': 'Tastenkürzel',
    'settings.toastPosition': 'Position der Benachrichtigungen',
    'settings.panes': 'Bereiche',
    'settings.lockPanes': 'Bereichsgrößen sperren',
    'addMailbox.cta': 'Postfach hinzufügen',
  },
}

// detectLocale reads the browser language and falls back to english for any
// locale we do not have a catalog for.
function detectLocale(): Locale {
  const lang = (navigator.language || 'en').slice(0, 2).toLowerCase()
  return lang === 'de' ? 'de' : 'en'
}

export const locale: Locale = detectLocale()

// t returns the localized string for a key, falling back to english, then to
// the key itself so a missing translation is visible but never crashes.
export function t(key: string): string {
  return catalog[locale][key] ?? catalog.en[key] ?? key
}

// isMac drives the modifier symbol and is used by shortcut matching.
export const isMac = /mac/i.test(navigator.userAgent)

// modSymbol is the display glyph for the primary modifier on this platform.
export const modSymbol = isMac ? '⌘' : 'Ctrl'

// shortcutLabel renders a combo like "mod+n" into a localized, platform-correct
// hint such as "⌘N" or "Ctrl+N".
export function shortcutLabel(combo: string): string {
  return combo
    .split('+')
    .map((part) => {
      if (part === 'mod') return modSymbol
      if (part === 'shift') return isMac ? '⇧' : 'Shift'
      if (part === 'alt') return isMac ? '⌥' : 'Alt'
      if (part === 'space') return 'Space'
      return part.length === 1 ? part.toUpperCase() : part
    })
    .join(isMac ? '' : '+')
}
