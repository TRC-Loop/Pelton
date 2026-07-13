// fonts.ts is the curated list of reader fonts: what the mail body iframe
// falls back to when a message does not declare its own fonts. keys are what
// the body_font setting stores; stacks are plain css font-family values (the
// iframe cannot resolve the app's css variables, so the default resolves to
// the ui font at build time in MailBody). shared with compose so written and
// read mail can use the same face (#64).

export interface BodyFont {
  key: string
  // labelKey resolves through i18n for the two generic entries; the named
  // families are shown as their own name (fonts are proper nouns).
  labelKey?: string
  label?: string
  // stack is null for the default entry, which uses the app's ui font.
  stack: string | null
}

export const bodyFonts: BodyFont[] = [
  { key: 'default', labelKey: 'settingsPanel.bodyFont.default', stack: null },
  { key: 'serif', labelKey: 'settingsPanel.bodyFont.serif', stack: 'Georgia, "Times New Roman", Times, serif' },
  { key: 'sans', labelKey: 'settingsPanel.bodyFont.sans', stack: 'Helvetica, Arial, sans-serif' },
  { key: 'mono', labelKey: 'settingsPanel.bodyFont.mono', stack: 'ui-monospace, "SF Mono", Consolas, "Liberation Mono", monospace' },
  { key: 'arial', label: 'Arial', stack: 'Arial, Helvetica, sans-serif' },
  { key: 'verdana', label: 'Verdana', stack: 'Verdana, Geneva, sans-serif' },
  { key: 'georgia', label: 'Georgia', stack: 'Georgia, serif' },
  { key: 'times', label: 'Times New Roman', stack: '"Times New Roman", Times, serif' },
  { key: 'courier', label: 'Courier New', stack: '"Courier New", Courier, monospace' },
]

// bodyFontStack maps a stored key to its css stack, or null for the default
// (and for any unknown key, so a removed entry degrades to the ui font).
export function bodyFontStack(key: string): string | null {
  return bodyFonts.find((f) => f.key === key)?.stack ?? null
}
