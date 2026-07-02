// preview.ts drives the in-app attachment previewer. an attachment card opens it
// with the message id and the attachment; the modal (mounted once at the app
// root) fetches the bytes and renders pdf, images, or text/code/markdown.

import { writable } from 'svelte/store'
import type { Attachment } from '../lib/types'

export interface PreviewTarget {
  messageId: number
  attachment: Attachment
}

export const previewTarget = writable<PreviewTarget | null>(null)

export function openPreview(messageId: number, attachment: Attachment): void {
  previewTarget.set({ messageId, attachment })
}

export function closePreview(): void {
  previewTarget.set(null)
}
