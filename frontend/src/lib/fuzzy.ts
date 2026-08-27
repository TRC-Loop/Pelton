// fuzzy.ts ranks candidate strings against a typed query for the command
// palette. It is a pure, store-free module so the ranking can be reasoned about
// on its own; the palette layers usage boosting on top of the score returned
// here.
//
// The algorithm is a small dynamic program in the style of fzy: every character
// of the query must appear in the candidate in order, and a match scores higher
// when its characters land on word boundaries and run consecutively. That is
// what makes "ea" prefer "Empty trash" over "Search mail".

/** A successful match: its score and which candidate characters were hit. */
export interface FuzzyMatch {
  /** Higher is better. Only comparable between candidates for the same query. */
  score: number
  /** Indices into the candidate that the query matched, ascending. */
  positions: number[]
}

// scoring weights. consecutive runs dominate, then word starts, then the
// string start; gaps cost little on their own so a late but clean match still
// beats an early ragged one.
const scoreMatch = 1
const bonusConsecutive = 1.4
const bonusWordStart = 1.1
const bonusStringStart = 1.6
const bonusCamel = 0.8
const penaltyGap = 0.06
const penaltyLeading = 0.01

// maxCandidate caps the dynamic program. Command labels are short; anything
// longer is matched on its first slice rather than allocating a large table.
const maxCandidate = 120

/** True when text[i] starts a word, so a match there reads as an initialism. */
function isWordStart(text: string, i: number): boolean {
  if (i === 0) {
    return true
  }
  const prev = text[i - 1]
  return prev === ' ' || prev === '-' || prev === '_' || prev === '/' || prev === '.' || prev === ':'
}

/** True when text[i] is a lowercase-to-uppercase step, as in "markUnread". */
function isCamelStep(text: string, i: number): boolean {
  if (i === 0) {
    return false
  }
  const prev = text[i - 1]
  const cur = text[i]
  return prev === prev.toLowerCase() && prev !== prev.toUpperCase() && cur === cur.toUpperCase() && cur !== cur.toLowerCase()
}

/** The bonus for landing a match on position i of text. */
function positionBonus(text: string, i: number): number {
  if (i === 0) {
    return bonusStringStart
  }
  if (isWordStart(text, i)) {
    return bonusWordStart
  }
  if (isCamelStep(text, i)) {
    return bonusCamel
  }
  return 0
}

/**
 * Scores query against text, returning null when text does not contain every
 * query character in order. An empty query matches everything with score 0, so
 * callers can use it to mean "no filtering".
 */
export function fuzzyMatch(query: string, text: string): FuzzyMatch | null {
  if (query === '') {
    return { score: 0, positions: [] }
  }
  const candidate = text.length > maxCandidate ? text.slice(0, maxCandidate) : text
  const q = query.toLowerCase()
  const t = candidate.toLowerCase()
  const m = q.length
  const n = t.length
  if (m > n) {
    return null
  }

  // best[i][j] is the score of matching q[0..i] within t[0..j] with q[i] landing
  // exactly on t[j]; running[i][j] is the best over any placement up to j.
  // from[i][j] records whether q[i] extended a run, so positions can be walked
  // back without a second pass over the table.
  const best = new Float64Array(m * n).fill(-Infinity)
  const running = new Float64Array(m * n).fill(-Infinity)

  for (let i = 0; i < m; i++) {
    let rowBest = -Infinity
    for (let j = 0; j < n; j++) {
      const cell = i * n + j
      if (q[i] === t[j]) {
        let score: number
        if (i === 0) {
          score = scoreMatch + positionBonus(candidate, j) - j * penaltyLeading
        } else if (j === 0) {
          score = -Infinity
        } else {
          const prevRow = (i - 1) * n + (j - 1)
          const consecutive = best[prevRow] + bonusConsecutive
          const gapped = running[prevRow] + positionBonus(candidate, j) - penaltyGap
          score = Math.max(consecutive, gapped)
          if (score !== -Infinity) {
            score += scoreMatch
          }
        }
        best[cell] = score
      }
      rowBest = Math.max(rowBest, best[cell])
      running[cell] = rowBest
    }
    if (rowBest === -Infinity) {
      return null
    }
  }

  // walk the last query character back to the start, preferring the placement
  // that produced the winning score.
  const positions: number[] = []
  let i = m - 1
  let j = n - 1
  let target = running[i * n + (n - 1)]
  while (i >= 0) {
    let hit = -1
    for (let k = j; k >= 0; k--) {
      if (best[i * n + k] === target) {
        hit = k
        break
      }
    }
    if (hit < 0) {
      // the table disagrees with the walk, which only happens on rounding;
      // fall back to a greedy placement so the caller still gets highlights.
      return { score: target, positions: greedyPositions(q, t) }
    }
    positions.unshift(hit)
    if (i === 0) {
      break
    }
    j = hit - 1
    target = running[(i - 1) * n + j]
    i--
  }

  return { score: running[(m - 1) * n + (n - 1)], positions }
}

/** Leftmost in-order placement of q in t, used only as a highlight fallback. */
function greedyPositions(q: string, t: string): number[] {
  const out: number[] = []
  let j = 0
  for (const ch of q) {
    const at = t.indexOf(ch, j)
    if (at < 0) {
      return out
    }
    out.push(at)
    j = at + 1
  }
  return out
}

/**
 * Splits text into alternating plain and matched runs for rendering, given the
 * positions from a match. Runs are never empty.
 */
export function highlightRuns(text: string, positions: number[]): { text: string; hit: boolean }[] {
  if (positions.length === 0) {
    return text === '' ? [] : [{ text, hit: false }]
  }
  const hits = new Set(positions)
  const runs: { text: string; hit: boolean }[] = []
  let start = 0
  let hit = hits.has(0)
  for (let i = 1; i <= text.length; i++) {
    const next = i < text.length && hits.has(i)
    if (i === text.length || next !== hit) {
      runs.push({ text: text.slice(start, i), hit })
      start = i
      hit = next
    }
  }
  return runs
}
