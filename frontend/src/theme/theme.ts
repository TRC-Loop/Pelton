// theme.ts applies theme and density preferences to the document by setting data
// attributes that tokens.css keys off. it also resolves the "system" theme from
// the os preference and lets callers react to os changes while in system mode.
// it only touches the dom; persistence lives in the prefs store.

import type { ThemePref, DensityPref } from '../lib/types'
import { applyAccent } from './accent'

export { applyAccent }

// prefersDark reports the current os color-scheme preference.
function prefersDark(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

// resolveTheme turns a preference into the concrete light/dark value, consulting
// the os when the preference is "system".
export function resolveTheme(pref: ThemePref): 'light' | 'dark' {
  if (pref === 'system') {
    return prefersDark() ? 'dark' : 'light'
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
  // expose the factor so full-viewport flow containers (the shell) can divide
  // their vh/vw by it: zoom does not shrink vh, so a 100vh shell would otherwise
  // render factor*100vh tall and clip its bottom row (status bar) off-screen.
  root.style.setProperty('--ui-scale', String(factor))
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
