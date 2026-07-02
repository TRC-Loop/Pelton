// i18n.ts is a small, dependency-free localization layer. it exposes a
// reactive locale store (backed by the settings table, like every other
// preference), a t() translation store for looking up strings, and the
// platform-correct modifier symbol so keyboard shortcut hints read natively
// (cmd on macos, ctrl elsewhere).
//
// coverage note: this catalogs the app's chrome (menus, shortcuts, common
// settings labels and actions, onboarding) rather than literally every string
// in the app. that is a deliberate scope choice: exhaustively translating
// every screen in one pass would be both huge and risky to review, whereas
// the catalog here is additive and any component can adopt t() incrementally.
//
// each language lives in its own file under lib/locales/. english is bundled
// directly since it is the always-available fallback; the other four are
// dynamically imported only once the user actually selects them, so nothing
// unused ships in the initial bundle.

import { writable, derived } from 'svelte/store'
import en from './locales/en'

export type Locale = 'en' | 'de' | 'fr' | 'nl' | 'es'

export const locales: Locale[] = ['en', 'de', 'fr', 'nl', 'es']

// each language is shown in its own spelling, not translated into the
// currently active one, so it stays recognizable no matter what is selected.
export const localeNames: Record<Locale, string> = {
  en: 'English',
  de: 'Deutsch',
  fr: 'Français',
  nl: 'Nederlands',
  es: 'Español',
}

const loaders: Record<Exclude<Locale, 'en'>, () => Promise<{ default: Record<string, string> }>> = {
  de: () => import('./locales/de'),
  fr: () => import('./locales/fr'),
  nl: () => import('./locales/nl'),
  es: () => import('./locales/es'),
}

// catalogs holds every locale's strings that have been loaded so far. english
// is present from the start; the rest are filled in by ensureLoaded.
const catalogs = writable<Partial<Record<Locale, Record<string, string>>>>({ en })

async function ensureLoaded(l: Locale): Promise<void> {
  if (l === 'en') return
  let has = false
  catalogs.update((c) => {
    has = !!c[l]
    return c
  })
  if (has) return
  const mod = await loaders[l]()
  catalogs.update((c) => ({ ...c, [l]: mod.default }))
}

// detectOSLocale reads the browser/OS language, used only to mark a
// "Recommended" option in the picker. it is never used to silently pick the
// active language: first run always defaults to English, and after that the
// user's own choice (persisted via settings) always wins.
export function detectOSLocale(): Locale {
  const lang = (navigator.language || 'en').slice(0, 2).toLowerCase()
  return (locales as string[]).includes(lang) ? (lang as Locale) : 'en'
}

// the active locale. initPrefs (stores/prefs.ts) sets this from the persisted
// "language" setting on startup; the default here is only what renders before
// that first load resolves.
export const locale = writable<Locale>('en')

export function setLocale(l: Locale): void {
  locale.set(l)
  void ensureLoaded(l)
}

// t is reactive: components use $t('key') so the whole tree re-renders the
// instant the language changes (or its catalog finishes loading), with no
// reload required. while a locale's catalog is still loading, keys fall back
// to english until it arrives.
export const t = derived([locale, catalogs], ([$locale, $catalogs]) => (key: string): string => {
  return $catalogs[$locale]?.[key] ?? $catalogs.en?.[key] ?? key
})

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
