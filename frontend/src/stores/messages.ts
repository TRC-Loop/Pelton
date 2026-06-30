// messages.ts owns the middle column: the list of summaries for the current
// selection (a folder or a unified view) or, when a query is active, the search
// results. it handles pagination and exposes optimistic helpers so flag toggles
// and deletes reflect immediately without a round trip.

import { writable, get } from 'svelte/store'
import type { MessageSummary, Selection } from '../lib/types'
import { listFolderMessages, listViewMessages, search } from '../lib/api'
import { type AsyncState, idle, loading, ready, failed } from '../lib/async'
import { errorMessage } from './toast'

// how many rows we request per page.
export const PAGE_SIZE = 50

export interface ListData {
  items: MessageSummary[]
  total: number
  // searching marks a search result set, where pagination does not apply.
  searching: boolean
}

export const messageList = writable<AsyncState<ListData>>(idle())

// the selection the current list belongs to, so loadMore can ask for the next
// page of the same thing.
let currentSelection: Selection | null = null
let currentOffset = 0

// fetchPage reads one page for a selection at the given offset.
async function fetchPage(sel: Selection, offset: number): Promise<{ items: MessageSummary[]; total: number }> {
  if (sel.kind === 'view') {
    const page = await listViewMessages(sel.view, PAGE_SIZE, offset)
    return { items: page.messages ?? [], total: page.total }
  }
  const page = await listFolderMessages(sel.folderId, PAGE_SIZE, offset)
  return { items: page.messages ?? [], total: page.total }
}

// loadList loads the first page for a selection, replacing the list.
export async function loadList(sel: Selection): Promise<void> {
  currentSelection = sel
  currentOffset = 0
  messageList.update((s) => loading(s))
  try {
    const { items, total } = await fetchPage(sel, 0)
    messageList.set(ready({ items, total, searching: false }))
  } catch (err) {
    messageList.set(failed(errorMessage(err)))
  }
}

// loadMore appends the next page if there are more rows. it is a no-op while
// searching or when everything is already loaded.
export async function loadMore(): Promise<void> {
  const state = get(messageList)
  if (!currentSelection || state.status !== 'ready' || !state.data || state.data.searching) {
    return
  }
  if (state.data.items.length >= state.data.total) {
    return
  }
  currentOffset += PAGE_SIZE
  try {
    const { items, total } = await fetchPage(currentSelection, currentOffset)
    messageList.update((s) => {
      if (s.status !== 'ready' || !s.data) {
        return s
      }
      return ready({ items: [...s.data.items, ...items], total, searching: false })
    })
  } catch (err) {
    // keep the existing list; surface the error without discarding loaded rows.
    messageList.update((s) => (s.data ? s : failed(errorMessage(err))))
  }
}

// SearchFilter is the optional date window applied to a search. 0 on a side
// leaves it open.
export interface SearchFilter {
  afterUnix: number
  beforeUnix: number
}

export const emptyFilter: SearchFilter = { afterUnix: 0, beforeUnix: 0 }

// runSearch replaces the list with ranked search results for a query and an
// optional date window.
export async function runSearch(query: string, filter: SearchFilter = emptyFilter): Promise<void> {
  messageList.update((s) => loading(s))
  try {
    const items = await search({
      query,
      afterUnix: filter.afterUnix,
      beforeUnix: filter.beforeUnix,
      limit: 200,
    })
    messageList.set(ready({ items, total: items.length, searching: true }))
  } catch (err) {
    messageList.set(failed(errorMessage(err)))
  }
}

// removeFromList drops a message from the loaded list after a delete.
export function removeFromList(id: number): void {
  messageList.update((s) => {
    if (s.status !== 'ready' || !s.data) {
      return s
    }
    return ready({
      ...s.data,
      items: s.data.items.filter((m) => m.id !== id),
      total: Math.max(0, s.data.total - 1),
    })
  })
}

// patchInList applies a partial update to one row, for optimistic flag changes.
export function patchInList(id: number, patch: Partial<MessageSummary>): void {
  messageList.update((s) => {
    if (s.status !== 'ready' || !s.data) {
      return s
    }
    return ready({
      ...s.data,
      items: s.data.items.map((m) => (m.id === id ? { ...m, ...patch } : m)),
    })
  })
}
