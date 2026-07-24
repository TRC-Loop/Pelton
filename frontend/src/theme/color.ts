// color.ts is the color model behind the theme editor's picker: parsing the
// color notations a theme token may hold, converting between rgb and hsv for
// the saturation/hue controls, and formatting a value back out. Tokens can
// legally hold things this cannot parse (color-mix, named colors); parseColor
// returns null for those and the caller keeps the text field authoritative.

export interface RGBA {
  r: number
  g: number
  b: number
  a: number
}

export interface HSV {
  h: number
  s: number
  v: number
}

const hexPattern = /^#([0-9a-f]{3,8})$/i
const rgbPattern = /^rgba?\(([^)]+)\)$/i

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n))
}

// parseColor reads #rgb, #rgba, #rrggbb, #rrggbbaa, rgb() and rgba(), in both
// the comma and the space separated form. Returns null for anything else.
export function parseColor(value: string): RGBA | null {
  const input = value.trim()
  const hex = hexPattern.exec(input)
  if (hex) {
    const digits = hex[1]
    const short = digits.length === 3 || digits.length === 4
    if (!short && digits.length !== 6 && digits.length !== 8) {
      return null
    }
    const pair = (i: number): number => {
      const s = short ? digits[i].repeat(2) : digits.slice(i * 2, i * 2 + 2)
      return parseInt(s, 16)
    }
    const hasAlpha = digits.length === 4 || digits.length === 8
    return { r: pair(0), g: pair(1), b: pair(2), a: hasAlpha ? pair(3) / 255 : 1 }
  }
  const rgb = rgbPattern.exec(input)
  if (rgb) {
    const parts = rgb[1].split(/[,/\s]+/).filter((p) => p.length > 0)
    if (parts.length < 3) {
      return null
    }
    const channel = (raw: string): number => {
      const n = raw.endsWith('%') ? (parseFloat(raw) / 100) * 255 : parseFloat(raw)
      return Number.isFinite(n) ? clamp(Math.round(n), 0, 255) : NaN
    }
    const r = channel(parts[0])
    const g = channel(parts[1])
    const b = channel(parts[2])
    if (Number.isNaN(r) || Number.isNaN(g) || Number.isNaN(b)) {
      return null
    }
    let a = 1
    if (parts.length > 3) {
      const raw = parts[3]
      const n = raw.endsWith('%') ? parseFloat(raw) / 100 : parseFloat(raw)
      if (!Number.isFinite(n)) {
        return null
      }
      a = clamp(n, 0, 1)
    }
    return { r, g, b, a }
  }
  return null
}

// formatColor writes the shortest faithful notation: hex while the color is
// opaque, rgba() once it is not, matching how the built-in palettes are
// written (solid colors as hex, borders as rgba).
export function formatColor({ r, g, b, a }: RGBA): string {
  if (a >= 1) {
    return toHex({ r, g, b, a: 1 })
  }
  return `rgba(${r}, ${g}, ${b}, ${Math.round(a * 1000) / 1000})`
}

// toHex renders the opaque part as #rrggbb, which is what the css gradients
// in the picker need.
export function toHex({ r, g, b }: RGBA): string {
  const hex = (n: number): string => clamp(Math.round(n), 0, 255).toString(16).padStart(2, '0')
  return `#${hex(r)}${hex(g)}${hex(b)}`
}

export function rgbToHsv({ r, g, b }: RGBA): HSV {
  const rn = r / 255
  const gn = g / 255
  const bn = b / 255
  const max = Math.max(rn, gn, bn)
  const min = Math.min(rn, gn, bn)
  const d = max - min
  let h = 0
  if (d !== 0) {
    if (max === rn) {
      h = ((gn - bn) / d) % 6
    } else if (max === gn) {
      h = (bn - rn) / d + 2
    } else {
      h = (rn - gn) / d + 4
    }
    h *= 60
    if (h < 0) {
      h += 360
    }
  }
  return { h, s: max === 0 ? 0 : d / max, v: max }
}

export function hsvToRgb({ h, s, v }: HSV): RGBA {
  const c = v * s
  const x = c * (1 - Math.abs(((h / 60) % 2) - 1))
  const m = v - c
  const sector = Math.floor(((h % 360) + 360) % 360 / 60)
  const table: [number, number, number][] = [
    [c, x, 0],
    [x, c, 0],
    [0, c, x],
    [0, x, c],
    [x, 0, c],
    [c, 0, x],
  ]
  const [r, g, b] = table[sector]
  return {
    r: Math.round((r + m) * 255),
    g: Math.round((g + m) * 255),
    b: Math.round((b + m) * 255),
    a: 1,
  }
}
