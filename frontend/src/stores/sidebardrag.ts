// sidebardrag.ts tracks which account a folder is currently being dragged
// within, so the other account sections can dim while it happens. A folder can
// only be reordered among its own siblings, and dimming makes the one valid drop
// area obvious instead of leaving the user to discover it by failing.

import { get, writable } from 'svelte/store'

// the account id of the folder drag in progress, or null when none is.
export const folderDragAccount = writable<number | null>(null)

// startFolderDrag marks a drag as running in one account. It only writes on a
// real change, so re-entry during a drag costs nothing.
export function startFolderDrag(accountId: number): void {
  if (get(folderDragAccount) !== accountId) {
    folderDragAccount.set(accountId)
  }
}

// endFolderDrag clears the drag state once the drop settles or is cancelled.
export function endFolderDrag(): void {
  if (get(folderDragAccount) !== null) {
    folderDragAccount.set(null)
  }
}
