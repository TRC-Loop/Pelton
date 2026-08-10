// palette.ts holds the command palette's state (#134): whether it is open, what
// has been typed, which picker step is active, and how often each command has
// been run. Ranking lives here too, so the component only renders what this
// store hands it.

import { writable, get } from 'svelte/store'
import { getSetting, setSetting, SettingKeys, search } from '../lib/api'
import type { MessageSummary } from '../lib/types'
import { fuzzyMatch } from '../lib/fuzzy'
import type { PaletteCommand, CommandGroup, CommandStep } from '../lib/commands'

/** True while the palette is on screen. */
export const paletteOpen = writable(false)

/** What the user has typed, prefix included. */
export const paletteQuery = writable('')

/** The active picker step, or null at the top level. */
export const paletteStep = writable<CommandStep | null>(null)

/**
 * Whether rows advertise a quick-select key (Enter for the first, then mod+1
 * through mod+9). On by default; the keys still work when the hints are hidden
 * only if this is on, since turning it off is how you free those combos.
 */
export const paletteQuickSelect = writable(true)

/** Turns the quick-select keys and their hints on or off. */
export function setPaletteQuickSelect(value: boolean): void {
  paletteQuickSelect.set(value)
  void setSetting(SettingKeys.paletteQuickSelect, String(value))
}

// quickSelectMax is how far the digits reach: the first row is Enter, so mod+1
// through mod+9 cover rows two to ten.
export const quickSelectMax = 9

/** Prefixes that narrow the results to one group. */
const prefixes: Record<string, CommandGroup> = {
  '>': 'action',
  '@': 'navigate',
  '#': 'setting',
  '/': 'mail',
}

/** A query split into its optional group filter and the text to match. */
export interface ParsedQuery {
  group: CommandGroup | null
  text: string
}

/**
 * Splits a leading >, @, # or / off the query. Prefixes are a shortcut, not a
 * requirement: an unprefixed query searches every group except mail, which is
 * only ever reached through "/" so that typing never hits the backend by
 * accident.
 */
export function parseQuery(raw: string): ParsedQuery {
  const group = prefixes[raw[0] ?? '']
  return group ? { group, text: raw.slice(1).trimStart() } : { group: null, text: raw.trim() }
}

// --- mail search ---
//
// The "/" prefix searches mail through the same ranked backend index the search
// bar uses. Results are not re-scored here: the backend already ordered them,
// and fuzzy matching them again against the same query would only fight it.

/** The current mail hits for a "/" query. */
export const paletteMail = writable<MessageSummary[]>([])

/** True while a mail search is in flight, so the palette can say so. */
export const paletteMailBusy = writable(false)

// mailLimit keeps a broad query from returning more than the list can show.
const mailLimit = 25

// debounce so a typed word is one query, not one per keystroke.
const mailDebounce = 180

let mailTimer: ReturnType<typeof setTimeout> | null = null
// a monotonic id, so a slow early query cannot overwrite a faster later one.
let mailSeq = 0
let lastMailQuery: string | null = null

/** Runs a debounced mail search, or clears the results for an empty query. */
export function requestMailSearch(text: string): void {
  if (text === lastMailQuery) {
    return
  }
  lastMailQuery = text
  if (mailTimer) {
    clearTimeout(mailTimer)
  }
  if (text === '') {
    mailSeq++
    paletteMail.set([])
    paletteMailBusy.set(false)
    return
  }
  paletteMailBusy.set(true)
  const seq = ++mailSeq
  mailTimer = setTimeout(() => {
    search({
      query: text,
      afterUnix: 0,
      beforeUnix: 0,
      limit: mailLimit,
      // the palette shows a short preview of the best matches, so it never
      // pages: the list is where a full result set is worked through.
      offset: 0,
      from: '',
      to: '',
      subject: '',
      hasAttachment: false,
    })
      .then((res) => {
        if (seq === mailSeq) {
          paletteMail.set(res.messages)
        }
      })
      .catch(() => {
        if (seq === mailSeq) {
          paletteMail.set([])
        }
      })
      .finally(() => {
        if (seq === mailSeq) {
          paletteMailBusy.set(false)
        }
      })
  }, mailDebounce)
}

/** Drops any pending search and its results. */
function resetMailSearch(): void {
  if (mailTimer) {
    clearTimeout(mailTimer)
    mailTimer = null
  }
  mailSeq++
  lastMailQuery = null
  paletteMail.set([])
  paletteMailBusy.set(false)
}

// --- usage boosting ---

interface Usage {
  /** How many times the command has been run. */
  n: number
  /** When it last ran, epoch milliseconds. */
  at: number
}

let usage: Record<string, Usage> = {}

// weights. frequency is logarithmic so a command run fifty times does not bury
// everything else, and recency decays over a fortnight.
const frequencyWeight = 0.9
const recencyWeight = 1.2
const recencyWindow = 14 * 24 * 60 * 60 * 1000

/** Loads persisted usage counts and preferences. Starts fresh on bad data. */
export async function initPalette(): Promise<void> {
  try {
    const stored = await getSetting(SettingKeys.paletteUsage)
    if (stored.found && stored.value) {
      const parsed: unknown = JSON.parse(stored.value)
      if (parsed && typeof parsed === 'object') {
        usage = parsed as Record<string, Usage>
      }
    }
  } catch {
    usage = {}
  }
  try {
    const stored = await getSetting(SettingKeys.paletteQuickSelect)
    if (stored.found) {
      paletteQuickSelect.set(stored.value !== 'false')
    }
  } catch {
    // the default stands
  }
}

// maxTracked caps what is persisted, so the setting cannot grow without bound
// as folders and themes come and go.
const maxTracked = 200

/** Records that a command ran, for future ranking. */
export function recordUse(id: string): void {
  const prev = usage[id]
  usage[id] = { n: (prev?.n ?? 0) + 1, at: Date.now() }
  const ids = Object.keys(usage)
  if (ids.length > maxTracked) {
    const stalest = ids.sort((a, b) => (usage[a].at ?? 0) - (usage[b].at ?? 0)).slice(0, ids.length - maxTracked)
    for (const id of stalest) {
      delete usage[id]
    }
  }
  void setSetting(SettingKeys.paletteUsage, JSON.stringify(usage))
}

/** The ranking bonus a command has earned, 0 when it has never been run. */
function usageBoost(id: string): number {
  const u = usage[id]
  if (!u) {
    return 0
  }
  const age = Date.now() - u.at
  const recency = age < recencyWindow ? 1 - age / recencyWindow : 0
  return Math.log1p(u.n) * frequencyWeight + recency * recencyWeight
}

// --- ranking ---

/** A command that survived filtering, with the highlight positions to render. */
export interface RankedCommand {
  command: PaletteCommand
  score: number
  positions: number[]
}

// maxResults keeps the rendered list bounded; with every folder, setting and
// theme in the index an unfiltered query can run to several hundred entries.
const maxResults = 60

/**
 * Filters and orders commands for a query. An empty query returns nothing: the
 * palette opens as a bare input and only fills in once something is typed.
 */
export function rankCommands(commands: PaletteCommand[], raw: string): RankedCommand[] {
  const { group, text } = parseQuery(raw)
  if (text === '') {
    return []
  }
  const ranked: RankedCommand[] = []
  for (const command of commands) {
    if (group && command.group !== group) {
      continue
    }
    const onLabel = fuzzyMatch(text, command.label)
    // keywords match but never highlight: they are synonyms the user did not
    // see, so marking up the label against them would light up the wrong letters.
    const onKeywords = onLabel ? null : command.keywords ? fuzzyMatch(text, command.keywords) : null
    const hit = onLabel ?? onKeywords
    if (!hit) {
      continue
    }
    ranked.push({
      command,
      score: (onLabel ? hit.score : hit.score * 0.6) + usageBoost(command.id),
      positions: onLabel ? hit.positions : [],
    })
  }
  ranked.sort((a, b) => b.score - a.score || a.command.label.localeCompare(b.command.label))
  return ranked.slice(0, maxResults)
}

/** Ranks the entries of an open picker step, which have no group filtering. */
export function rankStep(step: CommandStep, raw: string): RankedCommand[] {
  const text = raw.trim()
  if (text === '') {
    return step.items.map((command) => ({ command, score: 0, positions: [] }))
  }
  const ranked: RankedCommand[] = []
  for (const command of step.items) {
    const hit = fuzzyMatch(text, command.label)
    if (hit) {
      ranked.push({ command, score: hit.score + usageBoost(command.id), positions: hit.positions })
    }
  }
  ranked.sort((a, b) => b.score - a.score || a.command.label.localeCompare(b.command.label))
  return ranked.slice(0, maxResults)
}

// --- open / close ---

/** Opens the palette empty. */
export function openPalette(): void {
  paletteQuery.set('')
  paletteStep.set(null)
  resetMailSearch()
  paletteOpen.set(true)
}

/** Opens the palette already inside a picker step. */
export function openPaletteStep(step: CommandStep): void {
  paletteQuery.set('')
  paletteStep.set(step)
  resetMailSearch()
  paletteOpen.set(true)
}

/** Closes the palette and forgets the query. */
export function closePalette(): void {
  paletteOpen.set(false)
  paletteQuery.set('')
  paletteStep.set(null)
  resetMailSearch()
}

/** Toggles the palette, used by the shortcut so the same key closes it again. */
export function togglePalette(): void {
  if (get(paletteOpen)) {
    closePalette()
  } else {
    openPalette()
  }
}
