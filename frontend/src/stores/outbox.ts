// outbox.ts tracks the durable send queue and the live sync state. it refetches
// when the backend emits outbox:changed and reflects sync:state so the ui can
// show sending, queued and failed messages plus a sync indicator.

import { writable } from 'svelte/store'
import type { OutboxRow } from '../lib/types'
import { listOutbox, clearSentOutbox } from '../lib/api'
import { errorMessage, toastSuccess } from './toast'

export const outbox = writable<OutboxRow[]>([])

// syncing reflects whether a background or manual sync is currently running.
export const syncing = writable<boolean>(false)

// lastSynced holds the epoch ms of the last completed sync, or null if none yet
// this session. the status bar renders it as a relative time.
export const lastSynced = writable<number | null>(null)

// loadOutbox refetches the queue. it swallows errors into an empty list since the
// outbox view is secondary; a transient failure should not break the app. when
// the refetch reveals freshly-sent messages it shows a brief confirmation and
// prunes them so the queue does not keep completed rows around.
export async function loadOutbox(): Promise<void> {
  try {
    const rows = await listOutbox()
    const sent = rows.filter((r) => r.state === 'sent')
    outbox.set(rows.filter((r) => r.state !== 'sent'))
    if (sent.length > 0) {
      toastSuccess(sent.length === 1 ? 'Message sent.' : `${sent.length} messages sent.`)
      // prune the sent rows; the resulting outbox:changed event triggers another
      // load that simply finds them gone.
      void clearSentOutbox()
    }
  } catch (err) {
    // keep the previous contents; log to the console for diagnosis.
    console.error('load outbox failed:', errorMessage(err))
  }
}
