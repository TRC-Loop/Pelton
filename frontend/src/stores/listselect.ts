// listselect.ts holds the message-list multi-selection: the set of selected
// message ids and the anchor row for shift-range selection. this is ephemeral ui
// state, cleared whenever the list changes (folder switch, search, delete), so an
// in-memory store is correct here. the "open" message (detail pane) is tracked
// separately in selection.ts; this is only the bulk-action set.

import { writable, get } from 'svelte/store'

// the currently selected message ids. empty means no multi-selection is active.
export const selectedIds = writable<Set<number>>(new Set())

// the anchor row id for shift-click range selection, or null when there is none.
let anchorId: number | null = null

// clearSelection drops every selection and the anchor. called on folder/search
// changes so a stale selection never lingers across lists.
export function clearSelection(): void {
  anchorId = null
  selectedIds.set(new Set())
}

// toggleSelect flips one id (cmd/ctrl-click) and makes it the new anchor.
export function toggleSelect(id: number): void {
  const next = new Set(get(selectedIds))
  if (next.has(id)) {
    next.delete(id)
  } else {
    next.add(id)
  }
  anchorId = id
  selectedIds.set(next)
}

// selectOnly replaces the selection with a single id and sets it as the anchor.
export function selectOnly(id: number): void {
  anchorId = id
  selectedIds.set(new Set([id]))
}

// selectRange selects every id from the anchor through targetId inclusive, using
// the given ordered id list. when there is no anchor yet it falls back to a
// single selection.
export function selectRange(orderedIds: number[], targetId: number): void {
  if (anchorId === null) {
    selectOnly(targetId)
    return
  }
  const a = orderedIds.indexOf(anchorId)
  const b = orderedIds.indexOf(targetId)
  if (a < 0 || b < 0) {
    selectOnly(targetId)
    return
  }
  const [lo, hi] = a <= b ? [a, b] : [b, a]
  selectedIds.set(new Set(orderedIds.slice(lo, hi + 1)))
}

// deselect removes one id without touching the anchor, used after a bulk action
// consumes part of the selection.
export function deselect(id: number): void {
  const next = new Set(get(selectedIds))
  if (next.delete(id)) {
    selectedIds.set(next)
  }
}
