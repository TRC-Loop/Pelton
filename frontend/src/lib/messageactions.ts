// messageactions.ts holds the message operations that more than one surface
// runs: the message list's context menu, and the command palette. Each function
// patches the in-memory list first so the row reacts immediately, then reports
// failures as a toast rather than throwing, because every caller is a fire and
// forget menu click.
//
// Single-message actions that only ever come from a keyboard shortcut still
// live in App.svelte's dispatcher; this module is the shared subset.

import {
  setSeen,
  setFlagged,
  setFlagColor,
  deleteMessage,
  archiveMessage,
  downloadMessageOffline,
  removeOffline,
} from './api'
import type { ArchiveUndo } from './api'
import type { MessageSummary } from './types'
import { patchInList, removeFromList } from '../stores/messages'
import { openMessageId } from '../stores/selection'
import { clearSelection } from '../stores/listselect'
import { recordDeleted } from '../stores/undodelete'
import { recordArchived } from '../stores/undoarchive'
import { isVIPAddress, addVIP, removeVIP } from '../stores/vip'
import { errorMessage, toastError, toastSuccess } from '../stores/toast'
import { t } from './i18n'
import { get } from 'svelte/store'

/** Marks one message read or unread. */
export async function markSeen(item: MessageSummary, seen: boolean): Promise<void> {
  patchInList(item.id, { seen })
  try {
    await setSeen(item.id, seen)
  } catch (err) {
    toastError(errorMessage(err))
  }
}

/** Flags or unflags one message. */
export async function markFlagged(item: MessageSummary, flagged: boolean): Promise<void> {
  patchInList(item.id, { flagged })
  try {
    await setFlagged(item.id, flagged)
  } catch (err) {
    toastError(errorMessage(err))
  }
}

/** Sets a message's color label; 0 clears it. */
export async function markColor(item: MessageSummary, color: number): Promise<void> {
  patchInList(item.id, { flagColor: color })
  try {
    await setFlagColor(item.id, color)
  } catch (err) {
    toastError(errorMessage(err))
  }
}

/** Stars or unstars the message's sender. */
export async function toggleSenderVIP(item: MessageSummary): Promise<void> {
  try {
    if (isVIPAddress(item.fromAddress)) {
      await removeVIP(item.fromAddress)
    } else {
      await addVIP(item.fromAddress)
    }
  } catch (err) {
    toastError(errorMessage(err))
  }
}

/** Downloads a message for offline reading, or drops the offline copy. */
export async function setOffline(item: MessageSummary, offline: boolean): Promise<void> {
  patchInList(item.id, { offline })
  try {
    if (offline) {
      await downloadMessageOffline(item.id)
      toastSuccess(get(t)('messageList.toast.savedOffline'))
    } else {
      await removeOffline(item.id)
    }
  } catch (err) {
    toastError(errorMessage(err))
  }
}

/** Deletes one message, recording it for undo and closing it if it was open. */
export async function trashMessage(item: MessageSummary): Promise<void> {
  try {
    await deleteMessage(item.id)
    recordDeleted(item)
    removeFromList(item.id)
    if (get(openMessageId) === item.id) {
      openMessageId.set(null)
    }
  } catch (err) {
    toastError(errorMessage(err))
  }
}

/**
 * Reports a failed export-on-archive copy. The archive itself succeeded, so
 * this is a warning about the copy and never a failed action: staying quiet
 * would leave the user believing a local copy exists when it does not.
 */
export function reportArchiveExport(undo: ArchiveUndo): void {
  if (undo.exportError) {
    toastError(get(t)('mailboxes.export.failed').replace('{error}', undo.exportError))
  }
}

/** Archives one message, recording it for undo and closing it if it was open. */
export async function archive(item: MessageSummary): Promise<void> {
  try {
    const undo = await archiveMessage(item.id)
    reportArchiveExport(undo)
    if (undo.messageId) {
      recordArchived(item, undo.messageId, undo.originalFolderId)
    }
    removeFromList(item.id)
    if (get(openMessageId) === item.id) {
      openMessageId.set(null)
    }
  } catch (err) {
    toastError(errorMessage(err))
  }
}

// --- bulk variants ---
//
// These clear the multi-selection up front: the rows are about to change or
// disappear, and leaving them selected would let a second action run against a
// set the user can no longer see.

/** Marks every given message read or unread. */
export async function bulkMarkSeen(items: MessageSummary[], seen: boolean): Promise<void> {
  clearSelection()
  await Promise.all(items.map((item) => markSeen(item, seen)))
}

/** Flags or unflags every given message. */
export async function bulkMarkFlagged(items: MessageSummary[], flagged: boolean): Promise<void> {
  clearSelection()
  await Promise.all(items.map((item) => markFlagged(item, flagged)))
}

/** Sets the same color label on every given message. */
export async function bulkMarkColor(items: MessageSummary[], color: number): Promise<void> {
  clearSelection()
  await Promise.all(items.map((item) => markColor(item, color)))
}

/** Downloads every given message for offline reading, or drops the copies. */
export async function bulkSetOffline(items: MessageSummary[], offline: boolean): Promise<void> {
  clearSelection()
  await Promise.all(items.map((item) => setOffline(item, offline)))
}

/**
 * Deletes every given message. Sequential rather than parallel: each delete
 * pushes an undo entry and mutates the list, and the server is happier with a
 * queue than with fifty concurrent stores.
 */
export async function bulkTrash(items: MessageSummary[]): Promise<void> {
  clearSelection()
  for (const item of items) {
    await trashMessage(item)
  }
}

/** Archives every given message, sequentially, for the same reason as bulkTrash. */
export async function bulkArchive(items: MessageSummary[]): Promise<void> {
  clearSelection()
  for (const item of items) {
    await archive(item)
  }
}
