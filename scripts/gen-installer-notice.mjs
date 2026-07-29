#!/usr/bin/env node
// gen-installer-notice.mjs renders DISCLAIMER.md into the plain text file the
// windows installer shows on its liability page
// (build/windows/installer/disclaimer.txt). NSIS reads that file verbatim, so it
// needs hard-wrapped lines, no markdown, CRLF line endings and a UTF-8 BOM for
// the umlauts to survive.
//
// the generated file is committed; `make disclaimer` (and `make build-win`)
// regenerate it after every edit to DISCLAIMER.md.

import { readFileSync, writeFileSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const WIDTH = 76

// wrap breaks a paragraph into lines of at most WIDTH characters. words longer
// than the limit (urls) get their own line rather than being split.
function wrap(text) {
  const lines = []
  let line = ''
  for (const word of text.split(/\s+/)) {
    if (line && line.length + 1 + word.length > WIDTH) {
      lines.push(line)
      line = word
    } else {
      line = line ? `${line} ${word}` : word
    }
  }
  if (line) lines.push(line)
  return lines
}

function render(markdown) {
  const out = []
  const paragraphs = markdown.split(/\n{2,}/)

  for (const raw of paragraphs) {
    const block = raw.trim()
    if (!block) continue
    if (block === '---') {
      out.push('', '-'.repeat(WIDTH), '')
      continue
    }

    const heading = block.match(/^(#{1,6})\s+(.*)$/)
    if (heading) {
      const title = heading[2].trim()
      out.push(heading[1].length === 1 ? title.toUpperCase() : title, '')
      continue
    }

    const text = block
      .replace(/\n/g, ' ')
      .replace(/\*\*(.+?)\*\*/g, '$1')
      .replace(/\[(.+?)\]\((.+?)\)/g, '$1 ($2)')
      .replace(/<(https?:[^>]+)>/g, '$1')
      .replace(/\s+/g, ' ')
      .trim()
    out.push(...wrap(text), '')
  }

  return out.join('\n').replace(/\n{3,}/g, '\n\n').trim()
}

const markdown = readFileSync(join(root, 'DISCLAIMER.md'), 'utf8')
const target = join(root, 'build', 'windows', 'installer', 'disclaimer.txt')
writeFileSync(target, `﻿${render(markdown).replace(/\n/g, '\r\n')}\r\n`, 'utf8')
console.log(`wrote ${target}`)
