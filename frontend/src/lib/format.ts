// format.ts holds small pure display helpers: dates for the list and header,
// sender initials for the avatar, and human byte sizes for attachments. keeping
// them here keeps formatting out of component markup.

// formatListDate renders a date compactly for the message list: time for today,
// weekday for the last week, otherwise a short date. empty input yields "".
export function formatListDate(iso: string): string {
  if (!iso) {
    return ''
  }
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) {
    return ''
  }

  const now = new Date()
  const sameDay = date.toDateString() === now.toDateString()
  if (sameDay) {
    return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
  }

  const weekMs = 7 * 24 * 60 * 60 * 1000
  if (now.getTime() - date.getTime() < weekMs) {
    return date.toLocaleDateString(undefined, { weekday: 'short' })
  }

  const sameYear = date.getFullYear() === now.getFullYear()
  return date.toLocaleDateString(undefined, {
    day: 'numeric',
    month: 'short',
    year: sameYear ? undefined : 'numeric',
  })
}

// formatFullDate renders a full, unambiguous date for the message header.
export function formatFullDate(iso: string): string {
  if (!iso) {
    return ''
  }
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  return date.toLocaleString(undefined, {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// initials derives one or two uppercase letters for an avatar from a display
// name, falling back to the email local part, then to a neutral dot.
export function initials(name: string, email: string): string {
  const source = name.trim() || email.split('@')[0] || ''
  const parts = source.split(/[\s._-]+/).filter(Boolean)
  if (parts.length === 0) {
    return '•'
  }
  if (parts.length === 1) {
    return parts[0].slice(0, 2).toUpperCase()
  }
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
}

// avatarColor derives a stable, pleasant background color for an initials avatar
// from a seed (the sender email or name). the hue is hashed from the seed so the
// same sender always gets the same color; saturation and lightness are fixed so
// white text stays readable on top in both themes.
export function avatarColor(seed: string): string {
  let hash = 0
  for (let i = 0; i < seed.length; i++) {
    hash = (hash * 31 + seed.charCodeAt(i)) >>> 0
  }
  return `hsl(${hash % 360} 48% 45%)`
}

// formatBytes renders a byte count as a human size for the attachment list.
export function formatBytes(bytes: number): string {
  if (bytes <= 0) {
    return '0 B'
  }
  const units = ['B', 'KB', 'MB', 'GB']
  const exp = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  const value = bytes / Math.pow(1024, exp)
  const rounded = exp === 0 ? value : Math.round(value * 10) / 10
  return `${rounded} ${units[exp]}`
}

// formatRelative renders an epoch-ms timestamp as a short relative time for the
// status bar: "just now", "5m ago", "2h ago", otherwise a date.
export function formatRelative(epochMs: number): string {
  const diff = Date.now() - epochMs
  const sec = Math.floor(diff / 1000)
  if (sec < 45) {
    return 'just now'
  }
  const min = Math.floor(sec / 60)
  if (min < 60) {
    return `${min}m ago`
  }
  const hr = Math.floor(min / 60)
  if (hr < 24) {
    return `${hr}h ago`
  }
  return new Date(epochMs).toLocaleDateString(undefined, { day: 'numeric', month: 'short' })
}

// displayName returns the best label for a sender: name if present, else email.
export function displayName(name: string, email: string): string {
  return name.trim() || email
}
