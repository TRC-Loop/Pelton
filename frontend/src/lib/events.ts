// events.ts is the typed contract for the wails runtime events the backend emits
// (see events.go). it wraps EventsOn so subscribers get a typed payload and a
// single unsubscribe function, and keeps the event-name strings in one place.

import { EventsOn } from '../../wailsjs/runtime/runtime'

// event names, matching the go constants exactly.
export const EventNames = {
  mailNew: 'mail:new',
  syncProgress: 'sync:progress',
  syncState: 'sync:state',
  outboxChanged: 'outbox:changed',
  menu: 'menu:action',
  downloadProgress: 'download:progress',
  attachmentProgress: 'attachment:progress',
  updateAvailable: 'update:available',
} as const

// payloads, mirroring the go event structs.
export interface MailNewEvent {
  accountId: number
  folderId: number
  count: number
}

export interface SyncProgressEvent {
  accountId: number
  folder: string
  done: number
  total: number
}

export interface SyncStateEvent {
  running: boolean
  error: string
}

export interface DownloadProgressEvent {
  running: boolean
  done: number
  total: number
  percent: number
  etaSeconds: number
  label: string
  error: string
}

export interface AttachmentProgressEvent {
  running: boolean
  filename: string
  bytesDone: number
  bytesTotal: number
  filesDone: number
  filesTotal: number
  error: string
}


// UpdateAvailableEvent mirrors go's UpdateCheckResult, fired after an
// automatic (frequency-driven) update check completes.
export interface UpdateAvailableEvent {
  checked: boolean
  available: boolean
  currentVersion: string
  latestVersion: string
  releaseUrl: string
  error: string
}

// Unsubscribe removes an event listener.
export type Unsubscribe = () => void

// onMailNew fires when sync or idle pulled new messages.
export function onMailNew(cb: (e: MailNewEvent) => void): Unsubscribe {
  return EventsOn(EventNames.mailNew, (e: MailNewEvent) => cb(e))
}

// onSyncProgress fires per folder as a sync runs.
export function onSyncProgress(cb: (e: SyncProgressEvent) => void): Unsubscribe {
  return EventsOn(EventNames.syncProgress, (e: SyncProgressEvent) => cb(e))
}

// onSyncState fires when background sync starts or stops.
export function onSyncState(cb: (e: SyncStateEvent) => void): Unsubscribe {
  return EventsOn(EventNames.syncState, (e: SyncStateEvent) => cb(e))
}

// onOutboxChanged fires when the outbox contents or a message state change. it
// carries no payload; subscribers refetch the outbox.
export function onOutboxChanged(cb: () => void): Unsubscribe {
  return EventsOn(EventNames.outboxChanged, () => cb())
}

// onMenu fires when a native menubar item is chosen. the payload is a short
// action string (preferences, compose, sync, add-mailbox, about).
export function onMenu(cb: (action: string) => void): Unsubscribe {
  return EventsOn(EventNames.menu, (action: string) => cb(action))
}

// onDownloadProgress fires during a bulk offline range download.
export function onDownloadProgress(cb: (e: DownloadProgressEvent) => void): Unsubscribe {
  return EventsOn(EventNames.downloadProgress, (e: DownloadProgressEvent) => cb(e))
}

// onAttachmentProgress fires while saving one or more attachments.
export function onAttachmentProgress(cb: (e: AttachmentProgressEvent) => void): Unsubscribe {
  return EventsOn(EventNames.attachmentProgress, (e: AttachmentProgressEvent) => cb(e))
}

// onUpdateAvailable fires after an automatic update check completes (never
// for a manual "check now", which gets its result directly instead).
export function onUpdateAvailable(cb: (e: UpdateAvailableEvent) => void): Unsubscribe {
  return EventsOn(EventNames.updateAvailable, (e: UpdateAvailableEvent) => cb(e))
}
