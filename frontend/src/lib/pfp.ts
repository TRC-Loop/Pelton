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

// rng returns a deterministic pseudo-random stream (mulberry32) from a seed, so
// the generative geometric avatar can pull many varied values while staying
// identical for the same sender everywhere.
function rng(seed: number): () => number {
  let s = seed >>> 0
  return () => {
    s = (s + 0x6d2b79f5) >>> 0
    let t = s
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
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

// geometricSvg renders a generative Bauhaus-style mark on a 2x2 grid: each cell
// gets one of several shape primitives (disc, half-disc, quarter-disc, triangle,
// ring, dot) at a hash-driven rotation and color, with occasional empty cells for
// negative space. driving the layout, shapes and rotations from a seeded prng
// (not just the hue) makes every sender's mark structurally distinct.
function geometricSvg(seed: string): string {
  const rand = rng(hash(seed))
  const irand = (n: number): number => Math.floor(rand() * n)
  const baseHue = irand(360)
  const palette = [
    `hsl(${baseHue} 65% 55%)`,
    `hsl(${(baseHue + 35) % 360} 68% 45%)`,
    `hsl(${(baseHue + 190 + irand(60)) % 360} 72% 62%)`,
    `hsl(${(baseHue + 15) % 360} 34% 93%)`,
    `hsl(${(baseHue + 20) % 360} 42% 17%)`,
  ]
  const bg = palette[irand(2)]
  const others = palette.filter((c) => c !== bg)
  const s = 50
  const rot = (): number => 90 * irand(4)
  const col = (): string => others[irand(others.length)]

  // each entry draws one cell-sized primitive in local (0,0)-(s,s) coordinates.
  const shapes: Array<() => string> = [
    () => `<circle cx="${s / 2}" cy="${s / 2}" r="${s / 2}" fill="${col()}"/>`,
    () =>
      `<path d="M0 ${s / 2} A ${s / 2} ${s / 2} 0 0 1 ${s} ${s / 2} Z" fill="${col()}" transform="rotate(${rot()} ${s / 2} ${s / 2})"/>`,
    () =>
      `<path d="M0 0 L${s} 0 A ${s} ${s} 0 0 1 0 ${s} Z" fill="${col()}" transform="rotate(${rot()} ${s / 2} ${s / 2})"/>`,
    () => `<polygon points="0,0 ${s},0 0,${s}" fill="${col()}" transform="rotate(${rot()} ${s / 2} ${s / 2})"/>`,
    () => `<circle cx="${s / 2}" cy="${s / 2}" r="${s / 2}" fill="${col()}"/><circle cx="${s / 2}" cy="${s / 2}" r="${s / 4}" fill="${bg}"/>`,
    () => `<circle cx="${s / 2}" cy="${s / 2}" r="${s / 3.2}" fill="${col()}"/>`,
  ]

  let body = `<rect width="100" height="100" fill="${bg}"/>`
  for (let gx = 0; gx < 2; gx++) {
    for (let gy = 0; gy < 2; gy++) {
      if (rand() < 0.12) {
        continue
      }
      body += `<g transform="translate(${gx * s} ${gy * s})">${shapes[irand(shapes.length)]()}</g>`
    }
  }
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">${body}</svg>`
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
