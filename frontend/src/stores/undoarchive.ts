// undoarchive.ts keeps a small stack of recently archived messages so a global
// cmd+z can move the last archive back to where it came from. archive moves the
// message on the server and drops the local row, so undo re-locates it in
// Archive by its rfc Message-ID and moves it back. undo is impossible for a
// message with no Message-ID (rare), which we surface rather than silently fail.
//
// One entry is one action, not one message, so archiving a selection of eleven
// is undone in one press.

import { get, writable } from 'svelte/store'
import type { MessageSummary } from '../lib/types'
import { unarchiveMessage } from '../lib/api'
import { t } from '../lib/i18n'
import { toastInfo, toastError, errorMessage } from './toast'

interface ArchivedMessage {
  summary: MessageSummary
  messageId: string
  originalFolderId: number
}

const archived = writable<ArchivedMessage[][]>([])

// recordArchived remembers a just-archived message so it can be moved back.
export function recordArchived(summary: MessageSummary, messageId: string, originalFolderId: number): void {
  recordArchivedBatch([{ summary, messageId, originalFolderId }])
}

// recordArchivedBatch remembers a whole bulk archive or move as one undo step.
export function recordArchivedBatch(entries: ArchivedMessage[]): void {
  if (entries.length === 0) {
    return
  }
  archived.update((s) => [...s, entries])
}

// triggerUndoArchive restores the most recent archive, however many messages it
// covered. It returns true when it had something to undo (and kicks off the
// async move), false otherwise.
export function triggerUndoArchive(): boolean {
  const stack = get(archived)
  if (stack.length === 0) {
    return false
  }
  const last = stack[stack.length - 1]
  archived.set(stack.slice(0, -1))
  void (async () => {
    let failure = ''
    for (const entry of last) {
      try {
        await unarchiveMessage(entry.messageId, entry.originalFolderId)
      } catch (err) {
        failure = errorMessage(err)
      }
    }
    if (failure !== '') {
      toastError(failure)
      return
    }
    const label = get(t)
    toastInfo(last.length === 1 ? label('undo.archiveDone') : label('undo.archiveDoneCount').replace('{n}', String(last.length)))
  })()
  return true
}

export type { ArchivedMessage }
