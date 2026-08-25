// syncfailures.ts holds how each account's last sync went (#322). A sync run
// reported success as long as one account got through, so a mailbox that had
// been failing for weeks was invisible. The marks in the sidebar and the status
// bar both read this.
//
// The backend pushes the whole list when a run ends, so this is never polled.

import { get, writable } from 'svelte/store'
import { accountSyncStates, syncAccountNow } from '../lib/api'
import { onAccountSyncState } from '../lib/events'
import type { AccountSyncState } from '../lib/types'

// every account that has a recorded outcome, failing or not. Accounts that have
// never synced are absent: nothing to say yet is not the same as broken.
export const syncStates = writable<AccountSyncState[]>([])

// the accounts whose last sync failed, which is what the marks key off.
export const failedSyncs = writable<AccountSyncState[]>([])

// the account whose failure detail is open, or null when the dialog is closed.
// Both the sidebar mark and the status bar open the same one.
export const failureDetail = writable<AccountSyncState | null>(null)

function apply(states: AccountSyncState[]): void {
  syncStates.set(states)
  failedSyncs.set(states.filter((s) => s.failedAt !== ''))
  // keep an open dialog on the same account, so a retry that failed again
  // updates the reason in place instead of showing what it said last time.
  const open = get(failureDetail)
  if (open) {
    failureDetail.set(states.find((s) => s.accountId === open.accountId) ?? null)
  }
}

/** Loads the current state. Called once at startup; events carry it after that. */
export async function loadSyncStates(): Promise<void> {
  try {
    apply(await accountSyncStates())
  } catch {
    // a sync problem the ui cannot read about is not worth a second error on
    // top. The next run pushes the state anyway.
  }
}

/** Subscribes to the backend's end-of-run state. Returns the unsubscribe. */
export function watchSyncStates(): () => void {
  return onAccountSyncState((e) => apply(e.states ?? []))
}

/** Opens the failure detail for one account. */
export function showSyncFailure(state: AccountSyncState): void {
  failureDetail.set(state)
}

/** Closes the failure detail. */
export function closeSyncFailure(): void {
  failureDetail.set(null)
}

/**
 * Syncs one account again. Resolves true when it worked. The state is pushed by
 * the backend either way, so the dialog updates itself.
 */
export async function retrySync(accountId: number): Promise<boolean> {
  try {
    await syncAccountNow(accountId)
    return true
  } catch {
    return false
  }
}
