// selection.ts holds the ephemeral navigation state: which view or folder the
// message list is showing, which message is open, and the current search query.
// this is pure ui state so in-memory stores are correct here.

import { writable } from 'svelte/store'
import type { Selection, Folder, ViewKey } from '../lib/types'

// the default startup selection is the unified inbox, per the design.
export const defaultSelection: Selection = { kind: 'view', view: 'inbox', label: 'Unified Inbox' }

export const selection = writable<Selection>(defaultSelection)

// the currently open message id, or null when the detail pane is empty.
export const openMessageId = writable<number | null>(null)

// the active search query. an empty string means the normal list is shown.
export const searchQuery = writable<string>('')

// selectView switches the list to a unified view and clears the open message.
export function selectView(view: ViewKey, label: string): void {
  selection.set({ kind: 'view', view, label })
  openMessageId.set(null)
  searchQuery.set('')
}

// selectFolder switches the list to a single account folder.
export function selectFolder(folder: Folder): void {
  selection.set({
    kind: 'folder',
    folderId: folder.id,
    accountId: folder.accountId,
    label: folder.name,
  })
  openMessageId.set(null)
  searchQuery.set('')
}

// openMessage sets the open message for the detail pane.
export function openMessage(id: number): void {
  openMessageId.set(id)
}
