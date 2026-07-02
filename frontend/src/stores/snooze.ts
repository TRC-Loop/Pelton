// snooze.ts drives the snooze dialog. a row's context menu (or a shortcut) calls
// openSnooze with the message; the dialog reads snoozeTarget, collects a time and
// the "hide now" choice, and calls the backend. keeping the target in a store
// lets the dialog live once at the app root instead of per row.

import { writable } from 'svelte/store'

export interface SnoozeTarget {
  id: number
  subject: string
}

export const snoozeTarget = writable<SnoozeTarget | null>(null)

export function openSnooze(id: number, subject: string): void {
  snoozeTarget.set({ id, subject })
}

export function closeSnooze(): void {
  snoozeTarget.set(null)
}
