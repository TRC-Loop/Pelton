// profiles.ts holds the profile list and the current one (#270).
//
// Switching is a backend operation: it stops the old profile's sync, re-scopes
// the store and fires profile:changed. Everything the ui holds is scoped to a
// profile, so the app reloads its state on that event rather than trying to
// patch it item by item.

import { get, writable } from 'svelte/store'
import { listProfiles, activeProfile, switchProfile as switchOnBackend } from '../lib/api'
import type { Profile } from '../lib/types'

/** Every profile on the install, main first. */
export const profiles = writable<Profile[]>([])

/** The profile the app is in, or null before the first load. */
export const currentProfile = writable<Profile | null>(null)

/** True while a switch is in flight, so the switcher can show it is working. */
export const switching = writable(false)

/** Loads the profile list and the current one. */
export async function loadProfiles(): Promise<void> {
  try {
    const [all, active] = await Promise.all([listProfiles(), activeProfile()])
    profiles.set(all)
    currentProfile.set(active)
  } catch {
    // an install that cannot read its profiles still has one, and every read
    // falls back to it. Leaving the stores alone keeps the ui on what it had.
  }
}

/**
 * Switches to another profile. Does nothing when it is already the current one,
 * so a stray click on the profile you are in is not a reload.
 */
export async function switchTo(id: number): Promise<void> {
  if (get(currentProfile)?.id === id || get(switching)) {
    return
  }
  switching.set(true)
  try {
    await switchOnBackend(id)
  } finally {
    switching.set(false)
  }
}

/** Switches to the next profile in the bar, wrapping at the end. */
export function switchRelative(step: number): Promise<void> {
  const all = get(profiles)
  if (all.length < 2) {
    return Promise.resolve()
  }
  const index = all.findIndex((p) => p.id === get(currentProfile)?.id)
  const next = all[(index + step + all.length) % all.length]
  return switchTo(next.id)
}
