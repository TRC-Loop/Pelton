// messages.ts owns the middle column: the list of summaries for the current
// selection (a folder or a unified view) or, when a query is active, the search
// results. it handles pagination and exposes optimistic helpers so flag toggles
// and deletes reflect immediately without a round trip.

import { writable, get } from 'svelte/store'
import type { MessageSummary, Selection } from '../lib/types'
import {
  listFolderMessages,
  listViewMessages,
  listSavedViewMessages,
  fetchOlderMessages,
  search,
} from '../lib/api'
import { type AsyncState, idle, loading, ready, failed } from '../lib/async'
import { errorMessage, push } from './toast'
import { prefs } from './prefs'

// how many rows we request per page.
export const PAGE_SIZE = 50

export interface ListData {
  items: MessageSummary[]
  total: number
  // searching marks a search result set, where pagination does not apply.
  searching: boolean
  // hasOlder means the server still holds mail older than anything cached, so
  // the end of this list is not the end of the mailbox. total only counts what
  // is cached locally.
  hasOlder: boolean
  // backfilling is true while older mail is being fetched from the server, so
  // the list can show that it is waiting on the network rather than on disk.
  backfilling: boolean
}

// backfillFailed marks a 'load older' round trip that errored. The list falls
// back to a retry button instead of retrying on its own, so a server that is
// down does not get hammered once per scroll.
export const backfillFailed = writable(false)

export const messageList = writable<AsyncState<ListData>>(idle())

// the selection the current list belongs to, so loadMore can ask for the next
// page of the same thing.
let currentSelection: Selection | null = null
let currentOffset = 0
// bumped on every loadList call so a response from a superseded selection
// never overwrites a later one's result.
let loadGeneration = 0

// appendPage joins a freshly read page onto the loaded rows, dropping ids that
// are already on screen. Paging is by offset over a date-ordered query, so mail
// arriving between two page reads (a sync, or a backfill inserting older mail)
// can shift rows across the page boundary and repeat one. The list renders rows
// keyed by id, which a duplicate would break outright.
function appendPage(loaded: MessageSummary[], page: MessageSummary[]): MessageSummary[] {
  const seen = new Set(loaded.map((m) => m.id))
  return [...loaded, ...page.filter((m) => !seen.has(m.id))]
}

// Page is one page of cached summaries plus whether the server still holds
// older mail beyond them.
interface Page {
  items: MessageSummary[]
  total: number
  hasOlder: boolean
}

// fetchPage reads one page for a selection at the given offset.
async function fetchPage(sel: Selection, offset: number): Promise<Page> {
  if (sel.kind === 'view') {
    const page = await listViewMessages(sel.view, PAGE_SIZE, offset)
    return { items: page.messages ?? [], total: page.total, hasOlder: page.hasOlder }
  }
  if (sel.kind === 'savedView') {
    const page = await listSavedViewMessages(sel.viewId, PAGE_SIZE, offset)
    return { items: page.messages ?? [], total: page.total, hasOlder: page.hasOlder }
  }
  const page = await listFolderMessages(sel.folderId, PAGE_SIZE, offset)
  return { items: page.messages ?? [], total: page.total, hasOlder: page.hasOlder }
}

// loadList loads the first page for a selection, replacing the list.
export async function loadList(sel: Selection): Promise<void> {
  currentSelection = sel
  currentOffset = 0
  backfillFailed.set(false)
  const generation = ++loadGeneration
  messageList.update((s) => loading(s))
  try {
    const { items, total, hasOlder } = await fetchPage(sel, 0)
    if (generation !== loadGeneration) {
      return
    }
    messageList.set(ready({ items, total, searching: false, hasOlder, backfilling: false }))
  } catch (err) {
    if (generation !== loadGeneration) {
      return
    }
    messageList.set(failed(errorMessage(err)))
  }
}

// loadMore appends the next page if there are more rows. it is a no-op while
// searching or when everything is already loaded.
//
// running out of cached rows is not the same as running out of mail: a first
// sync only caches a folder's newest messages (#175), so when the cache is
// exhausted and the server still has older mail, this hands off to the backfill
// unless the user turned automatic backfill off.
export async function loadMore(): Promise<void> {
  const state = get(messageList)
  if (!currentSelection || state.status !== 'ready' || !state.data || state.data.searching) {
    return
  }
  if (state.data.items.length >= state.data.total) {
    if (state.data.hasOlder && get(prefs).syncAutoBackfill) {
      await loadOlder()
    }
    return
  }
  const nextOffset = currentOffset + PAGE_SIZE
  const generation = loadGeneration
  messageList.update((s) => loading(s))
  try {
    const { items, total, hasOlder } = await fetchPage(currentSelection, nextOffset)
    if (generation !== loadGeneration) {
      return
    }
    currentOffset = nextOffset
    messageList.update((s) => {
      if (!s.data) {
        return s
      }
      return ready({
        ...s.data,
        items: appendPage(s.data.items, items),
        total,
        hasOlder,
        searching: false,
      })
    })
  } catch (err) {
    if (generation !== loadGeneration) {
      return
    }
    // leave currentOffset unchanged so a retry re-requests this same page
    // instead of skipping it; keep the existing list and surface the error
    // without discarding loaded rows.
    messageList.update((s) => (s.data ? ready(s.data) : failed(errorMessage(err))))
  }
}

// loadOlder fetches the next batch of older mail from the server and appends
// whatever it cached. Unlike loadMore this is a network round trip, so it is
// guarded against overlapping runs and reports failure through backfillFailed
// rather than replacing the list with an error.
export async function loadOlder(): Promise<void> {
  const state = get(messageList)
  if (!currentSelection || state.status !== 'ready' || !state.data) {
    return
  }
  if (state.data.backfilling || !state.data.hasOlder) {
    return
  }

  const sel = currentSelection
  const generation = loadGeneration
  backfillFailed.set(false)
  messageList.update((s) => (s.data ? ready({ ...s.data, backfilling: true }) : s))

  try {
    const result = await fetchOlderMessages(sel)
    if (generation !== loadGeneration) {
      return
    }
    if (result.fetched === 0) {
      messageList.update((s) =>
        s.data ? ready({ ...s.data, backfilling: false, hasOlder: result.hasOlder }) : s,
      )
      return
    }
    // the newly cached mail is older than everything on screen, so it lands at
    // the end of the list; re-read from the current offset rather than
    // reloading from the top, which would lose the user's scroll position.
    const nextOffset = currentOffset + PAGE_SIZE
    const { items, total } = await fetchPage(sel, nextOffset)
    if (generation !== loadGeneration) {
      return
    }
    currentOffset = nextOffset
    messageList.update((s) => {
      if (!s.data) {
        return s
      }
      return ready({
        ...s.data,
        items: appendPage(s.data.items, items),
        total,
        hasOlder: result.hasOlder,
        backfilling: false,
      })
    })
  } catch (err) {
    if (generation !== loadGeneration) {
      return
    }
    backfillFailed.set(true)
    messageList.update((s) => (s.data ? ready({ ...s.data, backfilling: false }) : s))
    push('error', errorMessage(err))
  }
}

// SearchFilter carries the structured constraints from the search chips: an
// optional date window (0 on a side leaves it open), field-scoped terms, and
// the attachment toggle. Free text is passed separately to runSearch.
export interface SearchFilter {
  afterUnix: number
  beforeUnix: number
  from: string
  to: string
  subject: string
  hasAttachment: boolean
}

export const emptyFilter: SearchFilter = {
  afterUnix: 0,
  beforeUnix: 0,
  from: '',
  to: '',
  subject: '',
  hasAttachment: false,
}

// filterActive reports whether any chip constraint is set (used to decide
// between the ranked search and the plain folder list).
export function filterActive(f: SearchFilter): boolean {
  return (
    f.afterUnix > 0 ||
    f.beforeUnix > 0 ||
    f.from !== '' ||
    f.to !== '' ||
    f.subject !== '' ||
    f.hasAttachment
  )
}

// runSearch replaces the list with ranked search results for a query and the
// structured chip constraints.
export async function runSearch(query: string, filter: SearchFilter = emptyFilter): Promise<void> {
  messageList.update((s) => loading(s))
  try {
    const items = await search({
      query,
      afterUnix: filter.afterUnix,
      beforeUnix: filter.beforeUnix,
      from: filter.from,
      to: filter.to,
      subject: filter.subject,
      hasAttachment: filter.hasAttachment,
      limit: 200,
    })
    // search runs over the local index, so backfilling older mail from the
    // server is not part of it: hasOlder stays false and the list shows no
    // "load older" affordance on a result set.
    messageList.set(
      ready({ items, total: items.length, searching: true, hasOlder: false, backfilling: false }),
    )
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
    const items = s.data.items.filter((m) => m.id !== id)
    // only decrement total if the id was actually loaded (e.g. deleting from
    // the detail view for a message outside the currently loaded page must
    // not desync total from the real remaining count).
    const removed = items.length !== s.data.items.length
    return ready({
      ...s.data,
      items,
      total: removed ? Math.max(0, s.data.total - 1) : s.data.total,
    })
  })
}

// restoreToList re-inserts a previously removed row (used by undo-delete),
// keeping the newest-first order by date.
export function restoreToList(summary: MessageSummary): void {
  messageList.update((s) => {
    if (s.status !== 'ready' || !s.data) {
      return s
    }
    if (s.data.items.some((m) => m.id === summary.id)) {
      return s
    }
    const items = [...s.data.items, summary].sort((a, b) => (a.date < b.date ? 1 : a.date > b.date ? -1 : 0))
    return ready({ ...s.data, items, total: s.data.total + 1 })
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
