// accounts.ts loads and holds the sidebar data: the accounts, each account's
// folder tree, and the unified cross-account views. it exposes one load function
// plus a lightweight refresh used when sync events report new mail so badges and
// counts stay current.

import { writable } from 'svelte/store'
import type { Account, Folder, UnifiedView } from '../lib/types'
import { listAccounts, listFolders, listUnifiedViews, listPinnedFolders } from '../lib/api'
import { type AsyncState, idle, loading, ready, failed } from '../lib/async'
import { errorMessage } from './toast'

// the whole sidebar payload loaded together so the tree renders in one pass.
export interface SidebarData {
  accounts: Account[]
  foldersByAccount: Record<number, Folder[]>
  views: UnifiedView[]
  // the Pinned group, across every account, in the order the user arranged it.
  // these folders also appear in foldersByAccount: pinning mirrors, it does not
  // move.
  pinned: Folder[]
}

export const sidebar = writable<AsyncState<SidebarData>>(idle())

// loadSidebar fetches accounts, their folders and the unified views. on failure
// it records the error so the sidebar can show an error state rather than a
// blank column.
export async function loadSidebar(): Promise<void> {
  sidebar.update((s) => loading(s))
  try {
    const accounts = await listAccounts()
    const foldersByAccount: Record<number, Folder[]> = {}
    await Promise.all(
      accounts.map(async (acc) => {
        foldersByAccount[acc.id] = await listFolders(acc.id)
      }),
    )
    const views = await listUnifiedViews()
    const pinned = await listPinnedFolders()
    sidebar.set(ready({ accounts, foldersByAccount, views, pinned }))
  } catch (err) {
    sidebar.set(failed(errorMessage(err)))
  }
}

// refreshSidebar reloads counts quietly. it reuses loadSidebar but is named for
// intent at the call sites that react to sync events.
export const refreshSidebar = loadSidebar
