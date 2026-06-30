// sidebarstate.ts remembers which account sections in the sidebar the user has
// collapsed, persisted in the backend settings table (not localStorage) so it
// survives restarts. the set holds the account ids that are collapsed; absence
// means expanded, which keeps newly added accounts open by default.

import { writable } from 'svelte/store'
import { getSetting, setSetting } from '../lib/api'

const KEY = 'sidebar_collapsed_accounts'

export const collapsedAccounts = writable<Set<number>>(new Set())

// initSidebarState loads the persisted collapsed set once at startup. a failed
// or missing lookup just leaves every account expanded.
export async function initSidebarState(): Promise<void> {
  try {
    const { value, found } = await getSetting(KEY)
    if (found && value) {
      const ids = value
        .split(',')
        .map((s) => Number(s))
        .filter((n) => Number.isFinite(n))
      collapsedAccounts.set(new Set(ids))
    }
  } catch {
    // ignore: default to all-expanded.
  }
}

// toggleAccountCollapsed flips one account's collapsed state and persists the
// whole set. persistence is fire-and-forget; a failed write only means the
// choice will not survive a restart.
export function toggleAccountCollapsed(id: number): void {
  collapsedAccounts.update((current) => {
    const next = new Set(current)
    if (next.has(id)) {
      next.delete(id)
    } else {
      next.add(id)
    }
    void setSetting(KEY, [...next].join(','))
    return next
  })
}
