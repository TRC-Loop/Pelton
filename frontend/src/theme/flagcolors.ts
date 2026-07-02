// flagcolors.ts is the single source for the eight message flag colors. index 0
// is "no color"; 1..8 map to these entries and to imap $Label1..$Label8 when the
// user turns color syncing on. The hues are picked to stay legible as a thin
// left bar or a small dot in both light and dark themes.

export interface FlagColorDef {
  index: number
  name: string
  hex: string
}

export const flagColors: FlagColorDef[] = [
  { index: 1, name: 'Red', hex: '#E5484D' },
  { index: 2, name: 'Orange', hex: '#F76B15' },
  { index: 3, name: 'Yellow', hex: '#FFB224' },
  { index: 4, name: 'Green', hex: '#30A46C' },
  { index: 5, name: 'Teal', hex: '#12A594' },
  { index: 6, name: 'Blue', hex: '#3E63DD' },
  { index: 7, name: 'Purple', hex: '#8E4EC6' },
  { index: 8, name: 'Pink', hex: '#E93D82' },
]

// flagColorHex returns the hex for a color index, or empty for 0/unknown.
export function flagColorHex(index: number): string {
  const def = flagColors.find((c) => c.index === index)
  return def ? def.hex : ''
}

// flagColorName returns the human label for a color index.
export function flagColorName(index: number): string {
  const def = flagColors.find((c) => c.index === index)
  return def ? def.name : ''
}
