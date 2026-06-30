// shortcuts.ts (store) holds the live, user-customizable key bindings: a map from
// each app action to its current combo, seeded from the defaults and overlaid
// with the user's overrides (persisted in the backend settings table as json).
// App.svelte matches keydown events against this map, and the settings panel
// edits it.

import { writable, get } from 'svelte/store'
import { getSetting, setSetting } from '../lib/api'
import { shortcuts as defaults, type ShortcutAction } from '../lib/shortcuts'

const KEY = 'keyboard_shortcuts'

type Bindings = Record<ShortcutAction, string>

// defaultBindings builds the action -> default-combo map from the registry.
function defaultBindings(): Bindings {
  const map = {} as Bindings
  for (const s of defaults) {
    map[s.action] = s.combo
  }
  return map
}

export const bindings = writable<Bindings>(defaultBindings())

// recording is true while the settings panel is capturing a new combo, so the
// global keydown handler stands down and does not fire actions mid-capture.
export const recording = writable<boolean>(false)

// initShortcuts loads persisted overrides once at startup, overlaying them on the
// defaults so a newly added action always has a binding.
export async function initShortcuts(): Promise<void> {
  try {
    const { value, found } = await getSetting(KEY)
    if (found && value) {
      const parsed = JSON.parse(value) as Partial<Bindings>
      bindings.set({ ...defaultBindings(), ...parsed })
    }
  } catch {
    // ignore: keep the defaults.
  }
}

// persist writes only the entries that differ from the defaults, so resetting an
// action removes its override entirely.
function persist(map: Bindings): void {
  const base = defaultBindings()
  const overrides: Partial<Bindings> = {}
  for (const action of Object.keys(map) as ShortcutAction[]) {
    if (map[action] !== base[action]) {
      overrides[action] = map[action]
    }
  }
  void setSetting(KEY, JSON.stringify(overrides))
}

// setBinding rebinds one action to a new combo and persists the change.
export function setBinding(action: ShortcutAction, combo: string): void {
  bindings.update((map) => {
    const next = { ...map, [action]: combo }
    persist(next)
    return next
  })
}

// resetBinding restores one action to its default combo.
export function resetBinding(action: ShortcutAction): void {
  const base = defaultBindings()
  setBinding(action, base[action])
}

// resetAll restores every shortcut to its default.
export function resetAll(): void {
  const base = defaultBindings()
  bindings.set(base)
  persist(base)
}

// conflictsFor returns the action currently bound to combo other than the given
// one, or null. used by the editor to warn before assigning a duplicate.
export function conflictsFor(action: ShortcutAction, combo: string): ShortcutAction | null {
  const map = get(bindings)
  for (const other of Object.keys(map) as ShortcutAction[]) {
    if (other !== action && map[other] === combo) {
      return other
    }
  }
  return null
}
