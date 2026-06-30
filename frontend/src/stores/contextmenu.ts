// contextmenu.ts drives the single custom right-click menu. components build a
// list of items and call openContextMenu with the pointer position; the
// ContextMenu component renders it and closes on outside-click, escape or scroll.
// this replaces the webview's default menu (which exposed inspect and reload).

import { writable } from 'svelte/store'
import type { ComponentType } from 'svelte'

// a menu entry, or the 'separator' marker for a divider.
export interface MenuItem {
  label: string
  action: () => void
  icon?: ComponentType
  danger?: boolean
}

export type MenuEntry = MenuItem | 'separator'

export interface MenuState {
  x: number
  y: number
  entries: MenuEntry[]
}

export const contextMenu = writable<MenuState | null>(null)

// openContextMenu shows the menu at the given viewport coordinates.
export function openContextMenu(x: number, y: number, entries: MenuEntry[]): void {
  contextMenu.set({ x, y, entries })
}

export function closeContextMenu(): void {
  contextMenu.set(null)
}
