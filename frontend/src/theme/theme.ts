// theme.ts applies theme and density preferences to the document by setting data
// attributes that tokens.css keys off. it also resolves the "system" theme from
// the os preference and lets callers react to os changes while in system mode.
// it only touches the dom; persistence lives in the prefs store.

import type { ThemePref, DensityPref } from '../lib/types'
import { applyAccent } from './accent'

export { applyAccent }

// systemSchemeOverride is set from a native probe at startup for platforms where
// the css media query is unreliable (WebKitGTK on Linux never reports the
// desktop dark preference). null means "trust the media query".
let systemSchemeOverride: 'light' | 'dark' | null = null

// setSystemSchemeOverride records the os color scheme resolved natively, so
// resolveTheme('system') uses it instead of the (Linux-unreliable) media query.
export function setSystemSchemeOverride(scheme: 'light' | 'dark' | null): void {
  systemSchemeOverride = scheme
}

// prefersDark reports the current os color-scheme preference.
function prefersDark(): boolean {
  if (systemSchemeOverride) {
    return systemSchemeOverride === 'dark'
  }
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

// the schedule window for the "schedule" theme mode, as minutes since
// midnight. set from the prefs store; defaults match the backend defaults.
let scheduleStart = 19 * 60
let scheduleEnd = 7 * 60

// parseHM turns "HH:MM" into minutes since midnight, or null when malformed.
function parseHM(value: string): number | null {
  const m = /^(\d{1,2}):(\d{2})$/.exec(value)
  if (!m) {
    return null
  }
  const h = Number(m[1])
  const min = Number(m[2])
  if (h > 23 || min > 59) {
    return null
  }
  return h * 60 + min
}

// setThemeSchedule records the dark window for the "schedule" theme mode.
// malformed values keep the previous bound so a half-typed time never flips
// the theme.
export function setThemeSchedule(darkStart: string, darkEnd: string): void {
  scheduleStart = parseHM(darkStart) ?? scheduleStart
  scheduleEnd = parseHM(darkEnd) ?? scheduleEnd
}

// scheduleDark reports whether the current time falls inside the dark window.
// the window may cross midnight (the default 19:00-07:00 does); equal start
// and end mean dark around the clock.
function scheduleDark(): boolean {
  const now = new Date()
  const cur = now.getHours() * 60 + now.getMinutes()
  if (scheduleStart < scheduleEnd) {
    return cur >= scheduleStart && cur < scheduleEnd
  }
  return cur >= scheduleStart || cur < scheduleEnd
}

// resolveTheme turns a preference into the concrete light/dark value, consulting
// the os when the preference is "system" and the clock when it is "schedule".
export function resolveTheme(pref: ThemePref): 'light' | 'dark' {
  if (pref === 'system') {
    return prefersDark() ? 'dark' : 'light'
  }
  if (pref === 'schedule') {
    return scheduleDark() ? 'dark' : 'light'
  }
  return pref
}

// applyTheme writes the resolved theme onto the root element.
export function applyTheme(pref: ThemePref): void {
  document.documentElement.setAttribute('data-theme', resolveTheme(pref))
}

// applyDensity writes the density onto the root element.
export function applyDensity(pref: DensityPref): void {
  document.documentElement.setAttribute('data-density', pref)
}

// applyReduceMotion marks the root so css can disable transitions and
// animations. the os-level prefers-reduced-motion query is honored by the
// same css block regardless of this flag; this is the explicit in-app switch.
export function applyReduceMotion(on: boolean): void {
  if (on) {
    document.documentElement.setAttribute('data-reduce-motion', '')
  } else {
    document.documentElement.removeAttribute('data-reduce-motion')
  }
}

// applyScale zooms the whole interface by a string multiplier ("1" = 100%).
// zoom (rather than a root font-size) scales the px-based tokens and layout
// together, and is supported in both WKWebView and WebView2. an invalid or
// empty value resets to 100%.
export function applyScale(scale: string): void {
  const value = Number(scale)
  const factor = Number.isFinite(value) && value > 0 ? value : 1
  const root = document.documentElement
  // the typings don't include zoom, so assign through a loose style record.
  ;(root.style as unknown as Record<string, string>).zoom = String(factor)
  // expose the factor for currentUIScale below: fixed-positioned overlays need
  // it to convert unscaled pointer coordinates into the zoomed layout space.
  // the shell itself is sized with percentages, which follow zoom on their own.
  root.style.setProperty('--ui-scale', String(factor))
}

// currentUIScale reads back the factor applyScale last set, for any fixed-
// positioned overlay that must convert screen/pointer coordinates into the
// zoomed layout space (see ContextMenu.svelte's comment for why: css `zoom`
// leaves clientX/Y and getBoundingClientRect position in unscaled screen
// pixels while a `position: fixed` element is placed in zoomed layout space).
// a no-op divisor at the default 100%.
export function currentUIScale(): number {
  const raw = getComputedStyle(document.documentElement).getPropertyValue('--ui-scale')
  const n = parseFloat(raw)
  return n > 0 ? n : 1
}

// watchSystemTheme calls back whenever the os color scheme changes. the caller
// should re-apply the theme only while the preference is "system". returns an
// unsubscribe function.
export function watchSystemTheme(cb: () => void): () => void {
  const mq = window.matchMedia('(prefers-color-scheme: dark)')
  const handler = (): void => cb()
  mq.addEventListener('change', handler)
  return () => mq.removeEventListener('change', handler)
}
