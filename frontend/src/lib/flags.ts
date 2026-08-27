// flags.ts resolves a language to the flag shown next to it in the picker.
//
// A language is not a country, so the mapping is data rather than something
// derivable: each locale carries the flag people actually associate with it.
// Where the language is spoken in several countries the operating system's own
// region settles it, read locally from navigator.language, so an en-US install
// gets the US flag and an en-GB one the union flag without asking anybody or
// touching the network.
//
// The svgs come from flag-icons (MIT) and are bundled, never fetched.

/** Every flag in the set, keyed by lowercase ISO 3166-1 alpha-2 code. */
const files = import.meta.glob<string>('../../node_modules/flag-icons/flags/4x3/*.svg', {
  eager: true,
  query: '?url',
  import: 'default',
})

const byCountry: Record<string, string> = {}
for (const [path, url] of Object.entries(files)) {
  const code = path.slice(path.lastIndexOf('/') + 1, -4)
  byCountry[code] = url
}

// the flag a language falls back to when the operating system names no region,
// or names one the language is not spoken in.
const defaultCountry: Record<string, string> = {
  en: 'gb',
  de: 'de',
  fr: 'fr',
  nl: 'nl',
  es: 'es',
  pl: 'pl',
}

/**
 * The flag url for a language code, or undefined when there is nothing sensible
 * to show. Custom language files land here as undefined by design: an author's
 * own translation belongs to no country.
 */
export function flagFor(language: string): string | undefined {
  const code = language.toLowerCase().slice(0, 2)
  const country = osRegionFor(code) ?? defaultCountry[code]
  return country ? byCountry[country] : undefined
}

/**
 * The operating system's region, but only when it belongs to the language being
 * asked about. A de-DE install says nothing about which Spanish the user means,
 * so the Spanish tile keeps its default.
 */
function osRegionFor(language: string): string | undefined {
  const tag = typeof navigator === 'undefined' ? '' : navigator.language || ''
  const [base, region] = tag.toLowerCase().split('-')
  if (base !== language || !region || region.length !== 2) {
    return undefined
  }
  return byCountry[region] ? region : undefined
}
