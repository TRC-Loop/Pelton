// signatures.ts holds the user's reusable header/footer blocks (loaded from the
// backend, where they live in a real table) and the helpers compose and settings
// use. per-account default assignments are read/written through the api directly
// since they are small and account-scoped.

import { writable, get } from 'svelte/store'
import type { Signature } from '../lib/types'
import {
  listSignatures,
  saveSignature,
  deleteSignature,
  getAccountSignatures,
  setAccountSignatures,
} from '../lib/api'

export const signatures = writable<Signature[]>([])

// loadSignatures refreshes the list from the backend. call at startup and after
// any mutation.
export async function loadSignatures(): Promise<void> {
  signatures.set((await listSignatures()) ?? [])
}

// persistSignature creates or updates a block, then refreshes the list, and
// returns the stored block.
export async function persistSignature(s: Signature): Promise<Signature> {
  const saved = await saveSignature(s)
  await loadSignatures()
  return saved
}

// removeSignature deletes a block and refreshes the list.
export async function removeSignature(id: number): Promise<void> {
  await deleteSignature(id)
  await loadSignatures()
}

// signatureById looks up a loaded block by id (0 / missing -> undefined).
export function signatureById(id: number): Signature | undefined {
  if (!id) {
    return undefined
  }
  return get(signatures).find((s) => s.id === id)
}

// re-export the account-assignment api so callers import from one place.
export { getAccountSignatures, setAccountSignatures }
