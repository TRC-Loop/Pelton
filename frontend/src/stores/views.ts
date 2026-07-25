// views.ts holds the user's saved Views (preset searches) with their eager-run
// counts. it loads once and refreshes whenever the backend emits views:changed
// (after sync, or a create/update/delete/reorder), so sidebar badges stay live.

import { writable, get } from 'svelte/store'
import type { View } from '../lib/types'
import { listViews } from '../lib/api'

export const views = writable<View[]>([])

// loadViews fetches the saved views with fresh counts. failures are swallowed to
// an empty list so a broken view never blanks the sidebar; the backend logs the
// underlying error.
export async function loadViews(): Promise<void> {
  try {
    views.set(await listViews())
  } catch {
    views.set([])
  }
}

// upsertView merges a saved view into the store without a full reload, so the
// editor reflects a create/update immediately.
export function upsertView(v: View): void {
  views.update((list) => {
    const i = list.findIndex((x) => x.id === v.id)
    if (i === -1) {
      return [...list, v].sort((a, b) => a.position - b.position)
    }
    const next = [...list]
    next[i] = v
    return next
  })
}

// viewById returns the currently loaded view with the given id, or undefined.
export function viewById(id: number): View | undefined {
  return get(views).find((v) => v.id === id)
}

// blankView returns an empty view for the editor's create flow. Optional seed
// fields let "save this search as a view" prefill the query.
export function blankView(seed: Partial<View> = {}): View {
  return {
    id: 0,
    name: '',
    icon: 'bookmark',
    color: '',
    queryText: '',
    queryFrom: '',
    queryTo: '',
    querySubject: '',
    withinDays: 0,
    unreadOnly: false,
    flaggedOnly: false,
    hasAttachment: false,
    accountId: 0,
    position: 0,
    unreadCount: 0,
    totalCount: 0,
    ...seed,
  }
}
