// settingsnav.ts lets anything ask for the settings screen on a given category.
//
// The settings panel is owned by App.svelte, which is where the open state and
// the code-split import live. Components deep in the tree (the profile chip in
// the status bar, for one) raise a request through this store instead of
// threading a callback down to them.

import { writable } from 'svelte/store'

/** The category to open settings on, or null when nothing was requested. */
export const settingsRequest = writable<string | null>(null)

/** Asks for the settings screen, on the given category. */
export function openSettingsAt(category: string): void {
  settingsRequest.set(category)
}

/** Clears the request once it has been acted on. */
export function clearSettingsRequest(): void {
  settingsRequest.set(null)
}
