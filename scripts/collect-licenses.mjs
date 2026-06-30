#!/usr/bin/env node
// collect-licenses.mjs builds licenses/manifest.json: one entry per direct go
// module and npm package with its detected license type and full license text.
// the desktop backend embeds this file and serves it to the about section on
// demand, so the (large) text never ships in the frontend bundle.

import { execSync } from 'node:child_process'
import { readFileSync, writeFileSync, readdirSync, existsSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')

// detectType guesses the spdx-ish license name from the text, for the short
// "name: license" line. the full text is always kept so the guess is just a hint.
function detectType(text) {
  const t = text.toLowerCase()
  if (t.includes('apache license') && t.includes('version 2.0')) return 'Apache-2.0'
  if (t.includes('mit license') || (t.includes('permission is hereby granted, free of charge'))) return 'MIT'
  if (t.includes('mozilla public license') && t.includes('2.0')) return 'MPL-2.0'
  if (t.includes('gnu general public license') && t.includes('version 3')) return 'GPL-3.0'
  if (t.includes('gnu lesser general public')) return 'LGPL'
  if (t.includes('isc license')) return 'ISC'
  if (t.includes('redistribution and use in source and binary forms')) {
    if (t.includes('neither the name')) return 'BSD-3-Clause'
    return 'BSD-2-Clause'
  }
  if (t.includes('unlicense')) return 'Unlicense'
  return 'Other'
}

function findLicenseFile(dir) {
  if (!existsSync(dir)) return null
  const names = readdirSync(dir)
  const hit = names.find((n) => /^(LICENSE|LICENCE|COPYING|UNLICENSE)/i.test(n))
  return hit ? join(dir, hit) : null
}

const entries = []

// go modules: direct, non-main dependencies and their on-disk module cache dir.
console.log('==> go modules')
try {
  const out = execSync(
    "go list -m -f '{{if and (not .Indirect) (not .Main)}}{{.Path}}|{{.Dir}}{{end}}' all",
    { cwd: root, encoding: 'utf8' },
  )
  for (const line of out.split('\n')) {
    if (!line.trim()) continue
    const [path, dir] = line.split('|')
    const file = findLicenseFile(dir)
    if (!file) {
      console.log('  miss', path)
      continue
    }
    const text = readFileSync(file, 'utf8')
    entries.push({ group: 'go', name: path, license: detectType(text), text })
    console.log('  ok  ', path)
  }
} catch (err) {
  console.error('  go list failed:', err.message)
}

// npm packages: direct deps + devDeps from frontend/package.json.
console.log('==> npm packages')
const nm = join(root, 'frontend', 'node_modules')
if (existsSync(nm)) {
  const pkg = JSON.parse(readFileSync(join(root, 'frontend', 'package.json'), 'utf8'))
  const names = Object.keys({ ...pkg.dependencies, ...pkg.devDependencies })
  for (const name of names) {
    const dir = join(nm, name)
    const file = findLicenseFile(dir)
    let declared = ''
    try {
      declared = JSON.parse(readFileSync(join(dir, 'package.json'), 'utf8')).license || ''
    } catch {
      // no package.json license field; fall back to detection.
    }
    if (!file) {
      // some packages declare a license but ship no file; still list them.
      if (declared) {
        entries.push({ group: 'npm', name, license: declared, text: `License: ${declared} (no license file shipped).` })
        console.log('  ok* ', name)
      } else {
        console.log('  miss', name)
      }
      continue
    }
    const text = readFileSync(file, 'utf8')
    entries.push({ group: 'npm', name, license: declared || detectType(text), text })
    console.log('  ok  ', name)
  }
} else {
  console.log('  (frontend/node_modules missing, run: cd frontend && pnpm install)')
}

entries.sort((a, b) => (a.group + a.name).localeCompare(b.group + b.name))
writeFileSync(join(root, 'licenses', 'manifest.json'), JSON.stringify(entries, null, 2))
console.log(`done. ${entries.length} entries -> licenses/manifest.json`)
