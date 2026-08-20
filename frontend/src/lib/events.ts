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
  mailtoCompose: 'mailto:compose',
  viewsChanged: 'views:changed',
  importProgress: 'import:progress',
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

// ImportProgressEvent reports a running mail import. running is false on the
// final event, which carries the totals and any error. Progress is measured in
// source bytes: an mbox does not say how many messages it holds until it has
// been read through.
export interface ImportProgressEvent {
  running: boolean
  folder: string
  imported: number
  skipped: number
  failed: number
  bytesDone: number
  bytesTotal: number
  folders: string[]
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

// MailtoDraft mirrors the go MailtoDraft: a compose prefill parsed from a
// mailto: link. Address fields are comma-joined for the raw recipient inputs.
export interface MailtoDraft {
  to: string
  cc: string
  bcc: string
  subject: string
  body: string
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

// onMailtoCompose fires when a mailto: link is opened while the app is already
// running. A mailto that launched the app is delivered via consumePendingMailto.
export function onMailtoCompose(cb: (e: MailtoDraft) => void): Unsubscribe {
  return EventsOn(EventNames.mailtoCompose, (e: MailtoDraft) => cb(e))
}

// onImportProgress fires while mail is being imported from files, and once more
// with running false when the job ends.
export function onImportProgress(cb: (e: ImportProgressEvent) => void): Unsubscribe {
  return EventsOn(EventNames.importProgress, (e: ImportProgressEvent) => cb(e))
}

// onViewsChanged fires when saved views or their eager-run counts change. it
// carries no payload; subscribers reload the views store.
export function onViewsChanged(cb: () => void): Unsubscribe {
  return EventsOn(EventNames.viewsChanged, () => cb())
}
