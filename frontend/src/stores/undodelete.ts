// undodelete.ts keeps a small stack of recently deleted messages so a global
// cmd+z can bring the last one back. a delete only marks the row pending on the
// backend (it is expunged on the next sync), so undo is possible until then; if
// the row is already gone the undo call fails and we surface it.

import { get, writable } from 'svelte/store'
import type { MessageSummary } from '../lib/types'
import { undoDelete } from '../lib/api'
import { restoreToList } from './messages'
import { toastInfo, toastError, errorMessage } from './toast'

// the stack of deleted summaries, most recent last.
const deleted = writable<MessageSummary[]>([])

// recordDeleted remembers a just-deleted message so it can be undone.
export function recordDeleted(summary: MessageSummary): void {
  deleted.update((s) => [...s, summary])
}

// hasUndoableDelete reports whether there is anything to undo, so the key handler
// can decide whether to swallow cmd+z.
export function hasUndoableDelete(): boolean {
  return get(deleted).length > 0
}

// triggerUndoDelete restores the most recently deleted message. it returns true
// when it had something to undo (and kicks off the async restore), false when the
// stack was empty.
export function triggerUndoDelete(): boolean {
  const stack = get(deleted)
  if (stack.length === 0) {
    return false
  }
  const last = stack[stack.length - 1]
  deleted.set(stack.slice(0, -1))
  void (async () => {
    try {
      await undoDelete(last.id)
      restoreToList(last)
      toastInfo('Deletion undone.')
    } catch (err) {
      toastError(errorMessage(err))
    }
  })()
  return true
}
