// filetype.ts maps a content type (falling back to the filename extension) to a
// file-type glyph, so attachment cards in the reading pane and the composer read
// at a glance and stay consistent.
import {
  IconFileTypePdf,
  IconPhoto,
  IconFileTypeDoc,
  IconFileTypeXls,
  IconFileTypePpt,
  IconFileZip,
  IconFileText,
  IconMusic,
  IconMovie,
  IconFile,
} from '@tabler/icons-svelte'
import type { ComponentType } from 'svelte'

export function fileIcon(contentType: string, filename: string): ComponentType {
  const type = (contentType || '').toLowerCase()
  const ext = filename.includes('.') ? filename.split('.').pop()!.toLowerCase() : ''
  if (type.startsWith('image/') || ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'heic'].includes(ext)) return IconPhoto
  if (type.startsWith('audio/') || ['mp3', 'wav', 'm4a', 'flac', 'ogg'].includes(ext)) return IconMusic
  if (type.startsWith('video/') || ['mp4', 'mov', 'mkv', 'avi', 'webm'].includes(ext)) return IconMovie
  if (type === 'application/pdf' || ext === 'pdf') return IconFileTypePdf
  if (['doc', 'docx', 'odt', 'rtf'].includes(ext)) return IconFileTypeDoc
  if (['xls', 'xlsx', 'ods', 'csv'].includes(ext)) return IconFileTypeXls
  if (['ppt', 'pptx', 'odp'].includes(ext)) return IconFileTypePpt
  if (['zip', 'tar', 'gz', 'rar', '7z'].includes(ext)) return IconFileZip
  if (type.startsWith('text/') || ['txt', 'md', 'log'].includes(ext)) return IconFileText
  return IconFile
}
