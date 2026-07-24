// menubar.ts owns the in-app menu bar's customizable layout: the ordered menus,
// items and one-level submenus the user arranges in the bar's editor mode. It
// loads the layout from the backend settings table (one JSON value), merges it
// against the current built-in default so newly shipped menus/items are never
// silently lost, applies edits live, and persists every change. The bar
// (MenuBar.svelte) renders the resolved layout in normal mode and the raw layout
// in editor mode, so an edit previews instantly. Nothing here uses localStorage:
// the settings table is the source of truth, which also means the layout rides
// the existing config export/import.

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

// menuBarEditing toggles the bar's in-place editor mode (entered from settings).
export const menuBarEditing = writable(false)

export function setEditing(on: boolean): void {
  menuBarEditing.set(on)
}

// uid returns a short unique id for custom menus, items, submenus and separators.
function uid(prefix: string): string {
  return `${prefix}-${Math.random().toString(36).slice(2, 9)}`
}

// validItem keeps only entries the model understands, dropping built-in action
// items whose action left the catalog (a removed feature) while always keeping
// custom items, separators and submenus (whose children are validated too).
function validItem(item: MenuItemLayout): boolean {
  if (item.kind === 'separator') {
    return true
  }
  if (item.kind === 'submenu') {
    return true
  }
  return (item.kind === 'action' || item.kind === 'custom') && !!item.action && !!catalogByAction[item.action]
}

// sanitizeItems filters an items array and recurses one level into submenus.
function sanitizeItems(items: MenuItemLayout[]): MenuItemLayout[] {
  return items
    .filter(validItem)
    .map((item) =>
      item.kind === 'submenu'
        ? { ...item, items: (item.items ?? []).filter((c) => c.kind !== 'submenu' && validItem(c)) }
        : item,
    )
}

// mergeWithDefault reconciles a saved layout with the current built-in default:
// unknown built-in menus/items are appended (respecting the new-items policy),
// removed built-ins are dropped, and every custom menu/item/submenu is preserved
// in the user's order. A missing or malformed saved layout falls back to default.
function mergeWithDefault(saved: MenuBarLayout | null, mode: NewItemsMode): MenuBarLayout {
  const def = defaultMenuLayout()
  if (!saved || !Array.isArray(saved.menus)) {
    return def
  }
  const hideNew = mode === 'hidden'
  const savedMenus = saved.menus.filter((m) => m && Array.isArray(m.items))

  const merged: MenuLayout[] = savedMenus
    .map((menu) => {
      const defMenu = menu.builtin ? def.menus.find((d) => d.id === menu.id) : undefined
      if (menu.builtin && !defMenu) {
        return null
      }
      let items = sanitizeItems(menu.items)
      if (defMenu) {
        const present = new Set(items.map((i) => i.id))
        for (const di of defMenu.items) {
          if (di.kind === 'action' && !present.has(di.id)) {
            items = [...items, hideNew ? { ...di, hidden: true } : di]
          }
        }
      }
      return { ...menu, items }
    })
    .filter((m): m is MenuLayout => m !== null)

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

function commit(layout: MenuBarLayout): void {
  menuBarLayout.set(layout)
  persist(layout)
}

// --- container addressing ---
//
// Mutations target a container: a top-level menu (submenuId null) or a submenu
// within it (submenuId set). mapContainer applies fn to that container's items.

function mapContainer(
  menus: MenuLayout[],
  menuId: string,
  submenuId: string | null,
  fn: (items: MenuItemLayout[]) => MenuItemLayout[],
): MenuLayout[] {
  return menus.map((m) => {
    if (m.id !== menuId) {
      return m
    }
    if (submenuId === null) {
      return { ...m, items: fn(m.items) }
    }
    return {
      ...m,
      items: m.items.map((it) =>
        it.id === submenuId && it.kind === 'submenu' ? { ...it, items: fn(it.items ?? []) } : it,
      ),
    }
  })
}

function mutate(
  menuId: string,
  submenuId: string | null,
  fn: (items: MenuItemLayout[]) => MenuItemLayout[],
): void {
  const layout = get(menuBarLayout)
  commit({ ...layout, menus: mapContainer(layout.menus, menuId, submenuId, fn) })
}

// setMenus replaces the whole menu list (the editor's drag-and-drop commits the
// full reordered tree here, covering cross-menu and cross-submenu moves).
export function setMenus(menus: MenuLayout[]): void {
  commit({ version: menuBarLayoutVersion, menus })
}

// --- menu-level mutations ---

export function toggleMenuHidden(menuId: string): void {
  const layout = get(menuBarLayout)
  commit({
    ...layout,
    menus: layout.menus.map((m) => (m.id === menuId ? { ...m, hidden: !m.hidden } : m)),
  })
}

export function addCustomMenu(label: string): string {
  const layout = get(menuBarLayout)
  const id = uid('menu')
  commit({ ...layout, menus: [...layout.menus, { id, builtin: false, label, items: [] }] })
  return id
}

export function renameCustomMenu(menuId: string, label: string): void {
  const layout = get(menuBarLayout)
  commit({
    ...layout,
    menus: layout.menus.map((m) => (m.id === menuId && !m.builtin ? { ...m, label } : m)),
  })
}

export function removeMenu(menuId: string): void {
  const layout = get(menuBarLayout)
  commit({ ...layout, menus: layout.menus.filter((m) => (m.id === menuId ? m.builtin : true)) })
}

// --- item-level mutations (container-addressed) ---

export function toggleItemHidden(menuId: string, submenuId: string | null, itemId: string): void {
  mutate(menuId, submenuId, (items) =>
    items.map((i) => (i.id === itemId ? { ...i, hidden: !i.hidden } : i)),
  )
}

// CustomEntry is the data collected for a new or edited user-defined item.
export interface CustomEntry {
  label: string
  action: MenuActionId
  iconName?: string
  iconNodes?: IconNode[]
}

export function addCustomItem(menuId: string, submenuId: string | null, entry: CustomEntry): string {
  const id = uid('item')
  mutate(menuId, submenuId, (items) => [
    ...items,
    {
      kind: 'custom',
      id,
      action: entry.action,
      label: entry.label,
      iconName: entry.iconName || undefined,
      iconNodes: entry.iconNodes,
    },
  ])
  return id
}

export function addSeparator(menuId: string, submenuId: string | null): void {
  mutate(menuId, submenuId, (items) => [...items, { kind: 'separator', id: uid('sep') }])
}

// addSubmenu appends an empty submenu to a top-level menu (submenus never nest).
export function addSubmenu(menuId: string, label: string): string {
  const id = uid('sub')
  mutate(menuId, null, (items) => [...items, { kind: 'submenu', id, label, items: [] }])
  return id
}

// removeItem deletes a custom item, separator or submenu; built-in action items
// can only be hidden, never removed, so a reset can always restore them.
export function removeItem(menuId: string, submenuId: string | null, itemId: string): void {
  mutate(menuId, submenuId, (items) =>
    items.filter((i) => (i.id === itemId ? i.kind === 'action' : true)),
  )
}

// ItemPatch is the editable subset of an item the popover changes.
export interface ItemPatch {
  label?: string
  action?: MenuActionId
  iconName?: string | undefined
  iconNodes?: IconNode[] | undefined
}

// updateItem applies a popover edit to a custom item or submenu.
export function updateItem(
  menuId: string,
  submenuId: string | null,
  itemId: string,
  patch: ItemPatch,
): void {
  mutate(menuId, submenuId, (items) =>
    items.map((i) => (i.id === itemId ? { ...i, ...patch } : i)),
  )
}

export function setNewItemsMode(mode: NewItemsMode): void {
  menuBarNewItems.set(mode)
  void setSetting(SettingKeys.menuBarNewItems, mode)
}

export function resetLayout(): void {
  commit(defaultMenuLayout())
}

// --- render resolution ---

// RenderItem is a menu entry resolved for display: built-in items carry their
// catalog metadata; custom items and submenus carry their own label and icon;
// submenus carry their resolved children. Labels stay as i18n keys (built-ins)
// or raw text (custom) so the bar can translate reactively.
export interface RenderItem {
  kind: 'action' | 'custom' | 'separator' | 'submenu'
  id: string
  action?: MenuActionId
  labelKey?: string
  label?: string
  iconName?: string
  icon?: ComponentType
  iconNodes?: IconNode[]
  items?: RenderItem[]
  needsMessage?: boolean
  danger?: boolean
  hint?: string
}

export interface RenderMenu {
  id: string
  labelKey?: string
  label?: string
  items: RenderItem[]
}

// resolveItems turns stored items into render-ready ones: hidden and
// platform-inapplicable entries drop out, built-ins are decorated from the
// catalog, and submenus resolve their children (empty submenus are dropped).
function resolveItems(items: MenuItemLayout[], insideSubmenu: boolean): RenderItem[] {
  const out: RenderItem[] = []
  for (const item of items) {
    if (item.hidden) {
      continue
    }
    if (item.kind === 'separator') {
      out.push({ kind: 'separator', id: item.id })
      continue
    }
    if (item.kind === 'submenu') {
      if (insideSubmenu) {
        continue
      }
      const children = trimSeparators(resolveItems(item.items ?? [], true))
      if (children.length === 0) {
        continue
      }
      out.push({
        kind: 'submenu',
        id: item.id,
        label: item.label,
        iconName: item.iconName,
        iconNodes: item.iconNodes,
        items: children,
      })
      continue
    }
    const def = item.action ? catalogByAction[item.action] : undefined
    if (!def || (def.macOnly && !isMac)) {
      continue
    }
    if (item.kind === 'custom') {
      out.push({
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
      out.push({
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
  return out
}

// resolveBar turns a stored layout into the render-ready menus the bar shows in
// normal mode. Hidden menus and menus left empty (or holding only macOS items
// off macOS) are omitted.
export function resolveBar(layout: MenuBarLayout): RenderMenu[] {
  const menus: RenderMenu[] = []
  for (const menu of layout.menus) {
    if (menu.hidden) {
      continue
    }
    const items = trimSeparators(resolveItems(menu.items, false))
    if (items.some((i) => i.kind !== 'separator')) {
      menus.push({ id: menu.id, labelKey: menu.labelKey, label: menu.label, items })
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
