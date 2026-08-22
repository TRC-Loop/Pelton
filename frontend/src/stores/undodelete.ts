// undodelete.ts keeps a small stack of recently deleted messages so a global
// cmd+z can bring the last delete back. a delete only marks the row pending on
// the backend (it is expunged on the next sync), so undo is possible until
// then; if the row is already gone the undo call fails and we surface it.
//
// One entry is one action, not one message: deleting eleven selected messages
// is one thing the user did and one press has to put all eleven back.

import { get, writable } from 'svelte/store'
import type { MessageSummary } from '../lib/types'
import { undoDelete } from '../lib/api'
import { restoreToList } from './messages'
import { t } from '../lib/i18n'
import { toastInfo, toastError, errorMessage } from './toast'

// the stack of deleted batches, most recent last.
const deleted = writable<MessageSummary[][]>([])

// recordDeleted remembers a just-deleted message so it can be undone.
export function recordDeleted(summary: MessageSummary): void {
  recordDeletedBatch([summary])
}

// recordDeletedBatch remembers a whole bulk delete as one undo step.
export function recordDeletedBatch(summaries: MessageSummary[]): void {
  if (summaries.length === 0) {
    return
  }
  deleted.update((s) => [...s, summaries])
}

// hasUndoableDelete reports whether there is anything to undo, so the key handler
// can decide whether to swallow cmd+z.
export function hasUndoableDelete(): boolean {
  return get(deleted).length > 0
}

// triggerUndoDelete restores the most recent delete, however many messages it
// covered. It returns true when it had something to undo (and kicks off the
// async restore), false when the stack was empty.
export function triggerUndoDelete(): boolean {
  const stack = get(deleted)
  if (stack.length === 0) {
    return false
  }
  const last = stack[stack.length - 1]
  deleted.set(stack.slice(0, -1))
  void (async () => {
    // one failure does not stop the rest: the messages were deleted one by one
    // and a row that is already expunged cannot hold back the ones that are not.
    let failure = ''
    for (const summary of last) {
      try {
        await undoDelete(summary.id)
        restoreToList(summary)
      } catch (err) {
        failure = errorMessage(err)
      }
    }
    if (failure !== '') {
      toastError(failure)
      return
    }
    const label = get(t)
    toastInfo(last.length === 1 ? label('undo.deleteDone') : label('undo.deleteDoneCount').replace('{n}', String(last.length)))
  })()
  return true
}
