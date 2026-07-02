// move.ts drives the "Move to folder" dialog. a message's context menu calls
// openMove with the message; the dialog lists that account's folders and moves
// the message on the server (reusing the archive/undo machinery).

import { writable } from 'svelte/store'
import type { MessageSummary } from '../lib/types'

export const moveTarget = writable<MessageSummary | null>(null)

export function openMove(message: MessageSummary): void {
  moveTarget.set(message)
}

export function closeMove(): void {
  moveTarget.set(null)
}
