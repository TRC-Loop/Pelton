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

// what the in-app previewer can render. Everything else opens externally.
export type PreviewKind = 'pdf' | 'image' | 'text' | 'none'

const textExts = [
  'txt', 'md', 'markdown', 'log', 'csv', 'tsv', 'json', 'xml', 'yaml', 'yml',
  'js', 'ts', 'jsx', 'tsx', 'go', 'py', 'rs', 'c', 'h', 'cpp', 'java', 'rb',
  'sh', 'bash', 'zsh', 'toml', 'ini', 'conf', 'css', 'html', 'svelte', 'sql',
]

// previewKind classifies an attachment for the previewer.
export function previewKind(contentType: string, filename: string): PreviewKind {
  const type = (contentType || '').toLowerCase()
  const ext = filename.includes('.') ? filename.split('.').pop()!.toLowerCase() : ''
  if (type === 'application/pdf' || ext === 'pdf') return 'pdf'
  if (type.startsWith('image/') || ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp'].includes(ext)) return 'image'
  if (type.startsWith('text/') || textExts.includes(ext)) return 'text'
  return 'none'
}

// isMarkdown reports whether an attachment should render as markdown.
export function isMarkdown(filename: string): boolean {
  const ext = filename.includes('.') ? filename.split('.').pop()!.toLowerCase() : ''
  return ext === 'md' || ext === 'markdown'
}
