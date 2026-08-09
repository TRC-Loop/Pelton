// virustotal.ts holds the VirusTotal integration's settings and the verdicts
// scanned for the message currently open. the settings are loaded once at
// startup and refreshed whenever the External settings section changes them, so
// the reading pane can decide whether to offer a scan without asking the
// backend on every render.
//
// the verdict maps are per message and are cleared when a different message
// opens. the backend keeps the real cache; this only avoids re-crossing the
// bridge while one message is on screen.

import { get, writable } from 'svelte/store'
import { getVirusTotalConfig } from '../lib/api'
import type { VirusTotalConfig, Verdict } from '../lib/types'

// off in every respect until the backend says otherwise, so nothing can scan
// during the moment before the first load resolves.
const defaults: VirusTotalConfig = {
  enabled: false,
  hasApiKey: false,
  autoScanLinks: false,
  autoScanAttachments: false,
}

export const virusTotal = writable<VirusTotalConfig>(defaults)

// linkVerdicts is keyed by url, attachmentVerdicts by attachment id. both hold
// only the open message's results.
export const linkVerdicts = writable<Map<string, Verdict>>(new Map())
export const attachmentVerdicts = writable<Map<number, Verdict>>(new Map())

// scanning is true while a scan is in flight, so the ui can show progress and
// refuse to start a second one.
export const scanning = writable(false)

// the message the current verdicts belong to, so a stale result arriving after
// the user moved on is discarded rather than shown against the wrong message.
let verdictsFor = -1

/** loadVirusTotalConfig refreshes the settings from the backend. */
export async function loadVirusTotalConfig(): Promise<void> {
  try {
    virusTotal.set(await getVirusTotalConfig())
  } catch {
    // settings that cannot be read stay at the off defaults, which is the safe
    // direction to fail in: nothing gets scanned.
    virusTotal.set(defaults)
  }
}

/** scanEnabled reports whether scanning is possible: switched on and keyed. */
export function scanEnabled(cfg: VirusTotalConfig): boolean {
  return cfg.enabled && cfg.hasApiKey
}

/**
 * resetVerdicts clears the verdicts when a different message opens, and records
 * which message subsequent results belong to.
 */
export function resetVerdicts(messageId: number): void {
  if (verdictsFor === messageId) {
    return
  }
  verdictsFor = messageId
  linkVerdicts.set(new Map())
  attachmentVerdicts.set(new Map())
}

/** currentVerdictMessage is the message the stored verdicts belong to. */
export function currentVerdictMessage(): number {
  return verdictsFor
}

/** putLinkVerdict records one link result, ignoring it if the message changed. */
export function putLinkVerdict(messageId: number, url: string, verdict: Verdict): void {
  if (messageId !== verdictsFor) {
    return
  }
  const next = new Map(get(linkVerdicts))
  next.set(url, verdict)
  linkVerdicts.set(next)
}

/** putAttachmentVerdict records one attachment result for the open message. */
export function putAttachmentVerdict(messageId: number, attachmentId: number, verdict: Verdict): void {
  if (messageId !== verdictsFor) {
    return
  }
  const next = new Map(get(attachmentVerdicts))
  next.set(attachmentId, verdict)
  attachmentVerdicts.set(next)
}
