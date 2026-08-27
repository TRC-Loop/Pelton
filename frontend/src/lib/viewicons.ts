// viewicons.ts is the curated icon and color palette a saved View can use. A
// small, statically-imported set (rather than the full ~2MB tabler dataset the
// menu-bar icon picker lazy-loads) keeps Views light: the sidebar renders a
// view's icon by name with a plain component lookup, and the editor offers the
// same set as a grid.

import type { ComponentType } from 'svelte'
import {
  IconBookmark,
  IconStar,
  IconFlag,
  IconTag,
  IconMail,
  IconMailOpened,
  IconInbox,
  IconSend,
  IconPaperclip,
  IconFileInvoice,
  IconReceipt,
  IconAlertTriangle,
  IconBell,
  IconUser,
  IconUsers,
  IconBuilding,
  IconCalendarEvent,
  IconClock,
  IconHeart,
  IconBriefcase,
  IconShoppingCart,
  IconWorld,
  IconCode,
  IconFolder,
  IconHome,
  IconDeviceMobile,
  IconBallFootball,
  IconMusic,
  IconPlane,
  IconStethoscope,
  IconCar,
  IconSchool,
  IconBook,
  IconCamera,
  IconGift,
  IconPaw,
  IconWallet,
  IconChartBar,
  IconLock,
  IconCloud,
  IconToolsKitchen2,
  IconTicket,
} from '@tabler/icons-svelte'

// viewIcons maps an icon name to its component. The keys are the stored values.
export const viewIcons: Record<string, ComponentType> = {
  bookmark: IconBookmark,
  star: IconStar,
  flag: IconFlag,
  tag: IconTag,
  mail: IconMail,
  'mail-opened': IconMailOpened,
  inbox: IconInbox,
  send: IconSend,
  paperclip: IconPaperclip,
  'file-invoice': IconFileInvoice,
  receipt: IconReceipt,
  alert: IconAlertTriangle,
  bell: IconBell,
  user: IconUser,
  users: IconUsers,
  building: IconBuilding,
  calendar: IconCalendarEvent,
  clock: IconClock,
  heart: IconHeart,
  briefcase: IconBriefcase,
  cart: IconShoppingCart,
  world: IconWorld,
  code: IconCode,
  folder: IconFolder,
  home: IconHome,
  phone: IconDeviceMobile,
  sport: IconBallFootball,
  music: IconMusic,
  plane: IconPlane,
  health: IconStethoscope,
  car: IconCar,
  school: IconSchool,
  book: IconBook,
  camera: IconCamera,
  gift: IconGift,
  pet: IconPaw,
  wallet: IconWallet,
  chart: IconChartBar,
  lock: IconLock,
  cloud: IconCloud,
  food: IconToolsKitchen2,
  ticket: IconTicket,
}

// defaultViewIcon backs a view with no icon set, and any stored name that is not
// in the curated set (e.g. after a downgrade).
export const defaultViewIcon = 'bookmark'

// viewIconNames is the editor's grid order.
export const viewIconNames = Object.keys(viewIcons)

// viewIconComponent resolves a stored icon name to its component, falling back to
// the default so an unknown name never renders blank.
export function viewIconComponent(name: string): ComponentType {
  return viewIcons[name] ?? viewIcons[defaultViewIcon]
}

// ViewColor is one accent swatch. The value is the stored string; css is the
// color the swatch and icon render in.
export interface ViewColor {
  value: string
  css: string
}

// viewColors is the curated accent palette. An empty stored color falls back to
// the tertiary text color, matching the unified views.
export const viewColors: ViewColor[] = [
  { value: '', css: 'var(--text-tertiary)' },
  { value: 'red', css: '#e5484d' },
  { value: 'orange', css: '#f76b15' },
  { value: 'amber', css: '#ffb224' },
  { value: 'green', css: '#30a46c' },
  { value: 'teal', css: '#12a594' },
  { value: 'blue', css: '#0091ff' },
  { value: 'indigo', css: '#3e63dd' },
  { value: 'purple', css: '#8e4ec6' },
  { value: 'pink', css: '#e93d82' },
]

// viewColorCss resolves a stored color value to its css color, defaulting to the
// tertiary text color.
export function viewColorCss(value: string): string {
  return viewColors.find((c) => c.value === value)?.css ?? 'var(--text-tertiary)'
}
