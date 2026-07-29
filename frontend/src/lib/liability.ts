// liability.ts records that the user has seen the warranty and liability
// notice. the acknowledgement is stored as a single setting holding the app
// version and the moment it was given, so it can be shown again after a
// change to the terms without asking twice for the same wording.

import { getSetting, setSetting, SettingKeys, appVersion } from './api'
import { get } from 'svelte/store'
import { locale } from './i18n'

/** a recorded acknowledgement of the warranty and liability notice. */
export interface LiabilityAck {
  /** app version the notice was accepted in. */
  version: string
  /** ISO timestamp of the acceptance. */
  at: string
}

/**
 * liabilityAccepted reports whether the notice has already been accepted. a
 * failed lookup counts as accepted so a broken settings read never locks the
 * user out of their mail.
 */
export async function liabilityAccepted(): Promise<boolean> {
  try {
    const r = await getSetting(SettingKeys.liabilityAccepted)
    return r.found && r.value !== ''
  } catch {
    return true
  }
}

/** acceptLiability records the acknowledgement with the current version. */
export async function acceptLiability(): Promise<void> {
  let version = ''
  try {
    version = await appVersion()
  } catch {
    version = ''
  }
  const ack: LiabilityAck = { version, at: new Date().toISOString() }
  await setSetting(SettingKeys.liabilityAccepted, JSON.stringify(ack))
}

/** termsUrl returns the terms page in the user's language. */
export function termsUrl(): string {
  return get(locale) === 'de' ? 'https://pelton.app/terms/de' : 'https://pelton.app/terms'
}
