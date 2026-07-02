// pfp.ts generates a deterministic placeholder avatar ("pfp") for a sender as an
// inline SVG data uri, so it needs no network and is identical everywhere the
// same sender appears. four styles are offered:
//   - initials:  a colored disc with the sender's initials (the classic look)
//   - mono:      the same, but grayscale for a calmer, monochrome list
//   - pixel:     a github-style symmetric 5x5 identicon
//   - geometric: layered abstract polygons in a harmonious palette
// everything derives from a stable hash of the seed (the sender email or name).

import { initials } from './format'

export type PfpStyle = 'initials' | 'mono' | 'pixel' | 'geometric'

// fnv-1a style hash, kept unsigned so downstream maths stay positive.
function hash(seed: string): number {
  let h = 0
  for (let i = 0; i < seed.length; i++) {
    h = (h * 31 + seed.charCodeAt(i)) >>> 0
  }
  return h
}

// svgToDataUri url-encodes an svg string into an <img>-ready data uri.
function svgToDataUri(svg: string): string {
  return `data:image/svg+xml,${encodeURIComponent(svg)}`
}

// escapeXml neutralizes the characters that would otherwise break the svg's
// xml. label comes from a sender's display name, so a crafted or malformed
// From header (starting with "<" or "&", say) must not be able to produce
// invalid xml, which would make the data uri fail to decode and fall through
// to the browser's broken image icon instead of the generated placeholder.
function escapeXml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;')
}

// initialsSvg renders a disc with one or two letters. mono forces grayscale.
function initialsSvg(seed: string, label: string, mono: boolean): string {
  const h = hash(seed)
  const bg = mono ? `hsl(0 0% ${28 + (h % 14)}%)` : `hsl(${h % 360} 48% 45%)`
  const text = escapeXml(label || '•')
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
<circle cx="50" cy="50" r="50" fill="${bg}"/>
<text x="50" y="50" dy="0.35em" text-anchor="middle" fill="#fff" font-family="system-ui, sans-serif" font-size="42" font-weight="600">${text}</text>
</svg>`
}

// pixelSvg renders a symmetric 5x5 identicon. the left three columns are filled
// from hash bits and mirrored to the right, the github way.
function pixelSvg(seed: string): string {
  const h = hash(seed)
  const hue = h % 360
  const fg = `hsl(${hue} 55% 50%)`
  const bg = `hsl(${hue} 24% 94%)`
  const cell = 20
  let rects = ''
  for (let col = 0; col < 3; col++) {
    for (let row = 0; row < 5; row++) {
      // one bit per (col,row) cell; rotate the hash so each cell differs.
      const on = ((h >> (col * 5 + row)) & 1) === 1
      if (!on) {
        continue
      }
      const mirror = 4 - col
      rects += `<rect x="${col * cell}" y="${row * cell}" width="${cell}" height="${cell}" fill="${fg}"/>`
      if (mirror !== col) {
        rects += `<rect x="${mirror * cell}" y="${row * cell}" width="${cell}" height="${cell}" fill="${fg}"/>`
      }
    }
  }
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
<rect width="100" height="100" fill="${bg}"/>
${rects}
</svg>`
}

// geometricSvg renders a few overlapping translucent polygons in two analogous
// hues for a softer, "nice" abstract mark.
function geometricSvg(seed: string): string {
  const h = hash(seed)
  const hue = h % 360
  const bg = `hsl(${hue} 60% 42%)`
  const c2 = `hsl(${(hue + 40) % 360} 70% 60%)`
  const c3 = `hsl(${(hue + 320) % 360} 65% 55%)`
  // a couple of hash-driven coordinates so the composition varies per sender.
  const a = h % 50
  const b = (h >> 8) % 50
  const c = (h >> 16) % 50
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
<rect width="100" height="100" fill="${bg}"/>
<polygon points="0,${50 + a} ${50 + b},0 100,${30 + c} 100,100 0,100" fill="${c2}" opacity="0.85"/>
<polygon points="${20 + a},100 100,${20 + c} 100,100" fill="${c3}" opacity="0.9"/>
<circle cx="${30 + b}" cy="${30 + a}" r="14" fill="#fff" opacity="0.18"/>
</svg>`
}

// pfpDataUri returns the avatar for a style as an <img> data uri. label is only
// used by the initials styles; pass initials(name, email).
export function pfpDataUri(style: PfpStyle, seed: string, label: string): string {
  switch (style) {
    case 'mono':
      return svgToDataUri(initialsSvg(seed, label, true))
    case 'pixel':
      return svgToDataUri(pixelSvg(seed))
    case 'geometric':
      return svgToDataUri(geometricSvg(seed))
    default:
      return svgToDataUri(initialsSvg(seed, label, false))
  }
}

// pfpForSender is the convenience used across the ui: it derives the initials
// from the sender and returns the data uri for the chosen style.
export function pfpForSender(style: PfpStyle, name: string, email: string): string {
  const seed = (email || name).toLowerCase()
  return pfpDataUri(style, seed, initials(name, email))
}
