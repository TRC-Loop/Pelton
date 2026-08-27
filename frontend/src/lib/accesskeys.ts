// accesskeys.ts derives the menu bar's access keys, the underlined letter that
// opens a menu with alt+letter on Windows and Linux. The letters come from the
// resolved titles rather than a table: the bar is user-customizable, so menus
// can be renamed, added and reordered, and a fixed table could neither cover a
// custom menu nor stay in step across the locale files.

// AccessSplit is a title cut around its access letter, ready to render with the
// middle part underlined. A title with no free letter left has an empty letter
// and the whole title in before.
export interface AccessSplit {
  before: string
  letter: string
  after: string
}

// isLetterOrDigit reports whether a character can serve as an access key.
// Punctuation and spaces cannot: there is no alt+space convention to hang them
// on, and the underline would be invisible.
function isLetterOrDigit(ch: string): boolean {
  return /[\p{L}\p{N}]/u.test(ch)
}

// assignAccessKeys picks one lowercase letter per title: the first that no
// earlier title has claimed. Order decides, so with Mailbox before Mail,
// Mailbox takes "m" and Mail takes "a". A title whose every letter is taken
// gets an empty string and stays reachable by arrow keys only.
export function assignAccessKeys(titles: string[]): string[] {
  const claimed = new Set<string>()
  return titles.map((title) => {
    for (const ch of title) {
      if (!isLetterOrDigit(ch)) {
        continue
      }
      const key = ch.toLowerCase()
      if (claimed.has(key)) {
        continue
      }
      claimed.add(key)
      return key
    }
    return ''
  })
}

// splitAccessKey cuts a title around the first occurrence of its access letter,
// so the bar can underline that one character.
export function splitAccessKey(title: string, key: string): AccessSplit {
  if (!key) {
    return { before: title, letter: '', after: '' }
  }
  const at = Array.from(title).findIndex((ch) => ch.toLowerCase() === key)
  if (at === -1) {
    return { before: title, letter: '', after: '' }
  }
  const chars = Array.from(title)
  return {
    before: chars.slice(0, at).join(''),
    letter: chars[at],
    after: chars.slice(at + 1).join(''),
  }
}
