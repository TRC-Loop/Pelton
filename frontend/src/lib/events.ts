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
