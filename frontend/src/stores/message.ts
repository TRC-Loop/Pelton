// message.ts owns the detail pane: the full message for the open id. it loads on
// demand and exposes a way to swap in the remote-allowed body when the user opts
// to load remote images.

import { writable } from 'svelte/store'
import type { MessageDetail } from '../lib/types'
import { getMessage } from '../lib/api'
import { type AsyncState, idle, loading, ready, failed } from '../lib/async'
import { errorMessage } from './toast'

export const messageDetail = writable<AsyncState<MessageDetail>>(idle())

// loadMessage fetches the full message for the detail pane.
export async function loadMessage(id: number): Promise<void> {
  messageDetail.update((s) => loading(s))
  try {
    const detail = await getMessage(id)
    messageDetail.set(ready(detail))
  } catch (err) {
    messageDetail.set(failed(errorMessage(err)))
  }
}

// clearMessage empties the detail pane.
export function clearMessage(): void {
  messageDetail.set(idle())
}

// setBodyHtml swaps the rendered body, used after loading remote images.
export function setBodyHtml(html: string): void {
  messageDetail.update((s) => {
    if (s.status !== 'ready' || !s.data) {
      return s
    }
    return ready({ ...s.data, bodyHtmlSafe: html })
  })
}
