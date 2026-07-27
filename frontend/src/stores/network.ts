// network.ts tracks whether the app currently has internet access. it is pure
// runtime state (never persisted): the webview reports link changes through the
// standard online/offline window events, and navigator.onLine seeds the initial
// value. this is a best-effort signal (it reflects the OS network link, not a
// live reachability probe), so the sync path also flips it off when a connection
// attempt fails with a network error, and back on when a sync succeeds.

import { writable } from 'svelte/store'

export const online = writable<boolean>(typeof navigator === 'undefined' ? true : navigator.onLine)

// setOnline lets non-event code (a failed or successful sync) correct the signal.
export function setOnline(value: boolean): void {
  online.set(value)
}

if (typeof window !== 'undefined') {
  window.addEventListener('online', () => online.set(true))
  window.addEventListener('offline', () => online.set(false))
}
