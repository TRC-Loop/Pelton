// avatar.ts memoizes the remote photo candidate list for a sender so the same
// sender (or bimi domain) is only resolved once per source preference. the
// backend does the network work (bimi dns lookup, gravatar url); this dedupes
// the binding calls across the many avatars in the message list. keyed by source
// so changing the preference re-resolves.

import { senderPhotos } from './api'

const cache = new Map<string, Promise<string[]>>()

// photosFor returns the ordered remote photo candidates for a sender under the
// given source preference. errors and the "pfp" source resolve to an empty list
// so callers fall back to the generated placeholder.
export function photosFor(source: string, email: string): Promise<string[]> {
  if (source === 'pfp' || !email) {
    return Promise.resolve([])
  }
  const key = `${source}:${email.toLowerCase()}`
  let pending = cache.get(key)
  if (!pending) {
    // the binding returns json null for a sender without candidates (a nil Go
    // slice), which resolves successfully - normalize it, the catch alone
    // does not cover it.
    pending = senderPhotos(email)
      .then((found) => found ?? [])
      .catch(() => [])
    cache.set(key, pending)
  }
  return pending
}
