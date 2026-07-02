// undoarchive.ts keeps a small stack of recently archived messages so a global
// cmd+z can move the last one back to where it came from. archive moves the
// message on the server and drops the local row, so undo re-locates it in Archive
// by its rfc Message-ID and moves it back. undo is impossible for a message with
// no Message-ID (rare), which we surface rather than silently fail.

import { get, writable } from 'svelte/store'
import type { MessageSummary } from '../lib/types'
import { unarchiveMessage } from '../lib/api'
import { toastInfo, toastError, errorMessage } from './toast'

interface ArchivedEntry {
  summary: MessageSummary
  messageId: string
  originalFolderId: number
}

const archived = writable<ArchivedEntry[]>([])

// recordArchived remembers a just-archived message so it can be moved back.
export function recordArchived(summary: MessageSummary, messageId: string, originalFolderId: number): void {
  archived.update((s) => [...s, { summary, messageId, originalFolderId }])
}

// triggerUndoArchive restores the most recently archived message. it returns true
// when it had something to undo (and kicks off the async move), false otherwise.
export function triggerUndoArchive(): boolean {
  const stack = get(archived)
  if (stack.length === 0) {
    return false
  }
  const last = stack[stack.length - 1]
  archived.set(stack.slice(0, -1))
  void (async () => {
    try {
      await unarchiveMessage(last.messageId, last.originalFolderId)
      toastInfo('Moved back from Archive.')
    } catch (err) {
      toastError(errorMessage(err))
    }
  })()
  return true
}
