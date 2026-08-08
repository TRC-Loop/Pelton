// folderdialog.ts drives the create/rename/delete mailbox dialog (#132). The
// sidebar rows are deeply nested, so they raise a request through this store
// rather than each rendering its own copy of the dialog.

import { writable } from 'svelte/store'
import type { Folder } from '../lib/types'

/** What the dialog is being opened for. */
export type FolderDialogMode = 'create' | 'rename' | 'delete'

/** An open request for the folder dialog. */
export interface FolderDialogRequest {
  mode: FolderDialogMode
  accountId: number
  /** The folder being renamed or deleted; the parent when creating a subfolder. */
  folder: Folder | null
}

export const folderDialog = writable<FolderDialogRequest | null>(null)

/**
 * Opens the dialog to create a mailbox. Pass the parent to create a subfolder
 * under it, or null to create at the root of the account.
 */
export function openCreateFolder(accountId: number, parent: Folder | null = null): void {
  folderDialog.set({ mode: 'create', accountId, folder: parent })
}

/** Opens the dialog to rename a mailbox. */
export function openRenameFolder(folder: Folder): void {
  folderDialog.set({ mode: 'rename', accountId: folder.accountId, folder })
}

/** Opens the delete confirmation for a mailbox. */
export function openDeleteFolder(folder: Folder): void {
  folderDialog.set({ mode: 'delete', accountId: folder.accountId, folder })
}

/** Closes the dialog without acting. */
export function closeFolderDialog(): void {
  folderDialog.set(null)
}
