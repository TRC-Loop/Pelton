// views.ts holds the user's saved Views (preset searches) with their eager-run
// counts. it loads once and refreshes whenever the backend emits views:changed
// (after sync, or a create/update/delete/reorder), so sidebar badges stay live.

import { writable, get } from 'svelte/store'
import type { View } from '../lib/types'
import { listViews } from '../lib/api'
import { selection, selectView } from './selection'
import { t } from '../lib/i18n'

export const views = writable<View[]>([])

// editingView holds the view currently open in the editor overlay, or null when
// it is closed. Hosting it in a store lets any surface (sidebar, search bar,
// menu) open the editor, with the overlay rendered once at the app root.
export const editingView = writable<View | null>(null)

// openViewEditor opens the editor to create a new view, optionally seeded (used
// by "save this search as a view").
export function openViewEditor(seed: Partial<View> = {}): void {
  editingView.set(blankView(seed))
}

// editViewInEditor opens the editor to edit an existing view.
export function editViewInEditor(v: View): void {
  editingView.set({ ...v })
}

// closeViewEditor dismisses the editor overlay.
export function closeViewEditor(): void {
  editingView.set(null)
}

// loadViews fetches the saved views with fresh counts. failures are swallowed to
// an empty list so a broken view never blanks the sidebar; the backend logs the
// underlying error.
export async function loadViews(): Promise<void> {
  try {
    const list = await listViews()
    views.set(list)
    leaveDeletedView(list)
  } catch {
    // the list is emptied, but the selection is left alone: a request that
    // failed says nothing about whether the view the user is reading still
    // exists, and moving them off it would be worse than a momentarily empty
    // sidebar.
    views.set([])
  }
}

// leaveDeletedView sends the list back to the unified inbox when the saved View
// it is showing is no longer among the saved views.
//
// Deleting the view you are reading used to leave the selection pointing at it,
// so the list kept querying a view that was gone while the sidebar row it came
// from had disappeared. It lives here rather than next to the delete button
// because every path that removes a view refreshes through loadViews, including
// one deleted from the editor or in another window.
function leaveDeletedView(list: View[]): void {
  const sel = get(selection)
  if (sel.kind !== 'savedView') {
    return
  }
  if (list.some((v) => v.id === sel.viewId)) {
    return
  }
  selectView('inbox', get(t)('sidebar.unifiedInbox'))
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
    queryFrom: [],
    queryTo: [],
    querySubject: '',
    useRegex: false,
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
