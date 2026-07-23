// menubar.ts owns the in-app menu bar's customizable layout: the ordered menus
// and items the user arranges in settings. It loads the layout from the backend
// settings table (one JSON value), merges it against the current built-in
// default so newly shipped menus/items are never silently lost, applies edits
// live, and persists every change. The bar (MenuBar.svelte) and the editor
// (MenuBarCustomizer.svelte) both render from resolveBar() so an edit previews
// instantly. Nothing here uses localStorage: the settings table is the source
// of truth, which also means the layout rides the existing config export/import.

import { writable, get } from 'svelte/store'
import type { ComponentType } from 'svelte'
import { getSetting, setSetting, SettingKeys } from '../lib/api'
import { isMac } from '../lib/i18n'
import {
  defaultMenuLayout,
  catalogByAction,
  menuBarLayoutVersion,
  type MenuBarLayout,
  type MenuLayout,
  type MenuItemLayout,
  type MenuActionId,
  type IconNode,
} from '../lib/menuactions'

// NewItemsMode decides how a menu item shipped by a newer app version joins a
// saved layout: shown immediately, or added hidden so the bar looks unchanged
// until the user opts in.
export type NewItemsMode = 'visible' | 'hidden'

// the live layout and the new-items policy. Both start from the built-in
// default so the bar renders sanely before the first load resolves.
export const menuBarLayout = writable<MenuBarLayout>(defaultMenuLayout())
export const menuBarNewItems = writable<NewItemsMode>('visible')

// uid returns a short unique id for custom menus, items and separators.
function uid(prefix: string): string {
  return `${prefix}-${Math.random().toString(36).slice(2, 9)}`
}

// validItem keeps only entries the model understands, dropping built-in action
// items whose action left the catalog (a removed feature) while always keeping
// custom items and separators.
function validItem(item: MenuItemLayout): boolean {
  if (item.kind === 'separator') {
    return true
  }
  if (item.kind === 'custom') {
    return !!item.action && !!catalogByAction[item.action]
  }
  return item.kind === 'action' && !!item.action && !!catalogByAction[item.action]
}

// mergeWithDefault reconciles a saved layout with the current built-in default:
// unknown built-in menus/items are appended (respecting the new-items policy),
// removed built-ins are dropped, and every custom menu/item is preserved in the
// user's order. A missing or malformed saved layout falls back to the default.
function mergeWithDefault(saved: MenuBarLayout | null, mode: NewItemsMode): MenuBarLayout {
  const def = defaultMenuLayout()
  if (!saved || !Array.isArray(saved.menus)) {
    return def
  }
  const hideNew = mode === 'hidden'
  const savedMenus = saved.menus.filter((m) => m && Array.isArray(m.items))

  const merged: MenuLayout[] = savedMenus.map((menu) => {
    const defMenu = menu.builtin ? def.menus.find((d) => d.id === menu.id) : undefined
    // a saved built-in menu that no longer exists in the default is dropped.
    if (menu.builtin && !defMenu) {
      return null
    }
    let items = menu.items.filter(validItem)
    if (defMenu) {
      const present = new Set(items.map((i) => i.id))
      for (const di of defMenu.items) {
        // separators are positional, not identity-bearing; only merge in new
        // action items the user has never seen.
        if (di.kind === 'action' && !present.has(di.id)) {
          items = [...items, hideNew ? { ...di, hidden: true } : di]
        }
      }
    }
    return { ...menu, items }
  }).filter((m): m is MenuLayout => m !== null)

  // built-in menus present in the default but missing from the saved layout are
  // appended in default order.
  const haveIds = new Set(merged.map((m) => m.id))
  for (const dm of def.menus) {
    if (!haveIds.has(dm.id)) {
      merged.push(hideNew ? { ...dm, hidden: true } : dm)
    }
  }

  return { version: menuBarLayoutVersion, menus: merged }
}

// initMenuBar loads the persisted layout and new-items policy, merges the layout
// against the current default, and publishes both. Call once at startup.
export async function initMenuBar(): Promise<void> {
  let mode: NewItemsMode = 'visible'
  try {
    const raw = await getSetting(SettingKeys.menuBarNewItems)
    if (raw.found && (raw.value === 'visible' || raw.value === 'hidden')) {
      mode = raw.value
    }
  } catch {
    // keep the default policy
  }
  menuBarNewItems.set(mode)

  let saved: MenuBarLayout | null = null
  try {
    const raw = await getSetting(SettingKeys.menuBarLayout)
    if (raw.found && raw.value) {
      saved = JSON.parse(raw.value) as MenuBarLayout
    }
  } catch {
    saved = null
  }
  menuBarLayout.set(mergeWithDefault(saved, mode))
}

// persist writes the current layout back to the settings table. Fire-and-forget:
// a failed write only means the arrangement will not survive a restart.
function persist(layout: MenuBarLayout): void {
  void setSetting(SettingKeys.menuBarLayout, JSON.stringify(layout))
}

// commit publishes and persists a new layout in one step.
function commit(layout: MenuBarLayout): void {
  menuBarLayout.set(layout)
  persist(layout)
}

// setMenus replaces the menu list (used by the editor's drag-and-drop after a
// reorder) and persists.
export function setMenus(menus: MenuLayout[]): void {
  commit({ version: menuBarLayoutVersion, menus })
}

// setMenuItems replaces one menu's items (drag-and-drop within a menu) and
// persists.
export function setMenuItems(menuId: string, items: MenuItemLayout[]): void {
  const layout = get(menuBarLayout)
  commit({
    ...layout,
    menus: layout.menus.map((m) => (m.id === menuId ? { ...m, items } : m)),
  })
}

// toggleMenuHidden shows or hides a whole top-level menu.
export function toggleMenuHidden(menuId: string): void {
  const layout = get(menuBarLayout)
  commit({
    ...layout,
    menus: layout.menus.map((m) => (m.id === menuId ? { ...m, hidden: !m.hidden } : m)),
  })
}

// toggleItemHidden shows or hides a single item within a menu.
export function toggleItemHidden(menuId: string, itemId: string): void {
  const layout = get(menuBarLayout)
  commit({
    ...layout,
    menus: layout.menus.map((m) =>
      m.id === menuId
        ? { ...m, items: m.items.map((i) => (i.id === itemId ? { ...i, hidden: !i.hidden } : i)) }
        : m,
    ),
  })
}

// CustomEntry is the data the editor collects for a new user-defined item.
export interface CustomEntry {
  label: string
  action: MenuActionId
  iconName?: string
  iconNodes?: IconNode[]
}

// addCustomItem appends a user-defined item to a menu.
export function addCustomItem(menuId: string, entry: CustomEntry): void {
  const layout = get(menuBarLayout)
  const item: MenuItemLayout = {
    kind: 'custom',
    id: uid('item'),
    action: entry.action,
    label: entry.label,
    iconName: entry.iconName || undefined,
    iconNodes: entry.iconNodes,
  }
  commit({
    ...layout,
    menus: layout.menus.map((m) => (m.id === menuId ? { ...m, items: [...m.items, item] } : m)),
  })
}

// addSeparator appends a divider to a menu.
export function addSeparator(menuId: string): void {
  const layout = get(menuBarLayout)
  const item: MenuItemLayout = { kind: 'separator', id: uid('sep') }
  commit({
    ...layout,
    menus: layout.menus.map((m) => (m.id === menuId ? { ...m, items: [...m.items, item] } : m)),
  })
}

// removeItem deletes a custom item or a separator; built-in action items can
// only be hidden, never removed, so a reset can always restore them.
export function removeItem(menuId: string, itemId: string): void {
  const layout = get(menuBarLayout)
  commit({
    ...layout,
    menus: layout.menus.map((m) =>
      m.id === menuId
        ? { ...m, items: m.items.filter((i) => i.id === itemId ? i.kind === 'action' : true) }
        : m,
    ),
  })
}

// addCustomMenu appends a new empty top-level menu with a user label.
export function addCustomMenu(label: string): string {
  const layout = get(menuBarLayout)
  const id = uid('menu')
  commit({
    ...layout,
    menus: [...layout.menus, { id, builtin: false, label, items: [] }],
  })
  return id
}

// renameCustomMenu changes a custom menu's label (built-in menus are not
// renamed).
export function renameCustomMenu(menuId: string, label: string): void {
  const layout = get(menuBarLayout)
  commit({
    ...layout,
    menus: layout.menus.map((m) => (m.id === menuId && !m.builtin ? { ...m, label } : m)),
  })
}

// removeMenu deletes a custom menu; built-in menus can only be hidden.
export function removeMenu(menuId: string): void {
  const layout = get(menuBarLayout)
  commit({
    ...layout,
    menus: layout.menus.filter((m) => (m.id === menuId ? m.builtin : true)),
  })
}

// setNewItemsMode persists the policy for how future built-in items join the
// layout.
export function setNewItemsMode(mode: NewItemsMode): void {
  menuBarNewItems.set(mode)
  void setSetting(SettingKeys.menuBarNewItems, mode)
}

// resetLayout restores the built-in default arrangement.
export function resetLayout(): void {
  commit(defaultMenuLayout())
}

// --- render resolution ---

// RenderItem is a menu item resolved for display: built-in items carry their
// catalog metadata; custom items carry their own label and icon. Labels stay as
// i18n keys (built-ins) or raw text (custom) so the bar can translate reactively.
export interface RenderItem {
  kind: 'action' | 'custom' | 'separator'
  id: string
  action?: MenuActionId
  labelKey?: string
  label?: string
  iconName?: string
  icon?: ComponentType
  iconNodes?: IconNode[]
  needsMessage?: boolean
  danger?: boolean
  hint?: string
}

// RenderMenu is a top-level menu resolved for display.
export interface RenderMenu {
  id: string
  labelKey?: string
  label?: string
  items: RenderItem[]
}

// resolveBar turns a stored layout into the render-ready menus the bar shows:
// hidden menus/items and platform-inapplicable actions are dropped, and each
// built-in action item is decorated from the catalog. Empty menus are omitted so
// a menu the user emptied (or that only holds macOS items off macOS) disappears.
export function resolveBar(layout: MenuBarLayout): RenderMenu[] {
  const menus: RenderMenu[] = []
  for (const menu of layout.menus) {
    if (menu.hidden) {
      continue
    }
    const items: RenderItem[] = []
    for (const item of menu.items) {
      if (item.hidden) {
        continue
      }
      if (item.kind === 'separator') {
        items.push({ kind: 'separator', id: item.id })
        continue
      }
      const def = item.action ? catalogByAction[item.action] : undefined
      if (!def || (def.macOnly && !isMac)) {
        continue
      }
      if (item.kind === 'custom') {
        items.push({
          kind: 'custom',
          id: item.id,
          action: item.action,
          label: item.label,
          iconName: item.iconName,
          iconNodes: item.iconNodes,
          needsMessage: def.needsMessage,
          danger: def.danger,
          hint: def.hint,
        })
      } else {
        items.push({
          kind: 'action',
          id: item.id,
          action: item.action,
          labelKey: def.labelKey,
          iconName: def.iconName,
          icon: def.icon,
          needsMessage: def.needsMessage,
          danger: def.danger,
          hint: def.hint,
        })
      }
    }
    // trim leading/trailing and doubled separators left by hidden neighbors.
    const trimmed = trimSeparators(items)
    if (trimmed.some((i) => i.kind !== 'separator')) {
      menus.push({ id: menu.id, labelKey: menu.labelKey, label: menu.label, items: trimmed })
    }
  }
  return menus
}

// trimSeparators removes leading, trailing and consecutive separators so hiding
// an item never leaves a dangling or doubled divider.
function trimSeparators(items: RenderItem[]): RenderItem[] {
  const out: RenderItem[] = []
  for (const item of items) {
    if (item.kind === 'separator') {
      if (out.length === 0 || out[out.length - 1].kind === 'separator') {
        continue
      }
    }
    out.push(item)
  }
  while (out.length > 0 && out[out.length - 1].kind === 'separator') {
    out.pop()
  }
  return out
}
