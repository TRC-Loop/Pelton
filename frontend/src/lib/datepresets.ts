// datepresets.ts holds small date helpers shared by features that build a list
// of friendly relative time presets (snooze, send-later): "this evening",
// "tomorrow morning", "next Monday", etc.

// atTime returns a copy of d with its time set to h:m:00.000.
export function atTime(d: Date, h: number, m: number): Date {
  const c = new Date(d)
  c.setHours(h, m, 0, 0)
  return c
}

// addDays returns a copy of d shifted by n days.
export function addDays(d: Date, n: number): Date {
  const c = new Date(d)
  c.setDate(c.getDate() + n)
  return c
}

// nextWeekday returns the next date whose day-of-week is target (0=Sun..6=Sat),
// strictly after from.
export function nextWeekday(from: Date, target: number): Date {
  const c = new Date(from)
  do {
    c.setDate(c.getDate() + 1)
  } while (c.getDay() !== target)
  return c
}
