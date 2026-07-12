// icons.ts holds the icon overrides of the active custom theme: a map from
// icon name (the tabler name without the Icon prefix, e.g. "pencil") to a
// sanitized inline svg string. Components render overrides through
// ThemedIcon.svelte, which falls back to the bundled tabler icon when the
// active theme does not override a name. The svgs were sanitized by the
// backend at import and again on every load; nothing unvalidated reaches
// this store.

import { writable } from 'svelte/store'

/** icon-name -> sanitized svg markup of the active theme; empty when the
 * built-in default theme is active or the theme overrides no icons. */
export const iconOverrides = writable<Record<string, string>>({})

/** setIconOverrides replaces the active override set (empty object clears). */
export function setIconOverrides(icons: Record<string, string> | null): void {
  iconOverrides.set(icons ?? {})
}
