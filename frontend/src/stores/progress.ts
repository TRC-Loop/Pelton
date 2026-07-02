// progress.ts mirrors the backend download and attachment progress events into
// stores the status bar reads. it subscribes once at startup; the events carry
// running=false on the final tick so the ui can clear the bar.

import { writable } from 'svelte/store'
import {
  onDownloadProgress,
  onAttachmentProgress,
  type DownloadProgressEvent,
  type AttachmentProgressEvent,
} from '../lib/events'

export const downloadProgress = writable<DownloadProgressEvent | null>(null)
export const attachmentProgress = writable<AttachmentProgressEvent | null>(null)

// initProgress wires the event subscriptions. call once at startup.
export function initProgress(): void {
  onDownloadProgress((e) => {
    downloadProgress.set(e.running ? e : null)
    // keep a completed/failed message visible briefly, then clear.
    if (!e.running) {
      const finished = e
      downloadProgress.set(finished)
      setTimeout(() => downloadProgress.update((cur) => (cur === finished ? null : cur)), 2500)
    }
  })
  onAttachmentProgress((e) => {
    if (!e.running) {
      const finished = e
      attachmentProgress.set(finished)
      setTimeout(() => attachmentProgress.update((cur) => (cur === finished ? null : cur)), 1200)
    } else {
      attachmentProgress.set(e)
    }
  })
}
