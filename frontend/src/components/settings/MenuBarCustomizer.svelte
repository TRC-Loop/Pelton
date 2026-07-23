<script lang="ts">
  // MenuBarCustomizer is the settings editor for the in-app menu bar layout:
  // drag to reorder top-level menus and their items (including across menus),
  // show/hide anything, add custom entries (label + action + icon) and new
  // top-level menus, choose how future built-in items join the layout, and reset
  // to the default. Every change flows through stores/menubar.ts, so the bar
  // previews it live and it persists (and rides config export/import).
  import { flip } from 'svelte/animate'
  import { dndzone, type DndEvent } from 'svelte-dnd-action'
  import {
    IconGripVertical,
    IconEye,
    IconEyeOff,
    IconTrash,
    IconPlus,
    IconChevronRight,
    IconChevronDown,
    IconRotate,
    IconSeparatorHorizontal,
    IconPhoto,
    IconCheck,
    IconX,
  } from '@tabler/icons-svelte'
  import ThemedIcon from '../common/ThemedIcon.svelte'
  import MenuGlyph from '../common/MenuGlyph.svelte'
  import IconPicker from './IconPicker.svelte'
  import { t, isMac } from '../../lib/i18n'
  import { prefs } from '../../stores/prefs'
  import {
    menuActionCatalog,
    catalogByAction,
    type MenuLayout,
    type MenuItemLayout,
    type MenuActionId,
    type IconNode,
  } from '../../lib/menuactions'
  import {
    menuBarLayout,
    menuBarNewItems,
    setMenus,
    toggleMenuHidden,
    toggleItemHidden,
    addCustomItem,
    addSeparator,
    removeItem,
    addCustomMenu,
    renameCustomMenu,
    removeMenu,
    setNewItemsMode,
    resetLayout,
    type NewItemsMode,
  } from '../../stores/menubar'

  // a local mirror of the store's menus that drag events mutate during a drag;
  // it resyncs whenever the store changes (including our own commits).
  let menuList: MenuLayout[] = $menuBarLayout.menus
  $: menuList = $menuBarLayout.menus

  $: flipMs = $prefs.reduceMotion ? 0 : 160

  // actions offered for a custom entry; macOS-only ones are dropped off macOS.
  $: actionOptions = menuActionCatalog.filter((d) => !(d.macOnly && !isMac))

  let expanded = new Set<string>()
  let confirmingReset = false

  // add-entry form state; only one menu shows its form at a time.
  let addingTo: string | null = null
  let formLabel = ''
  let formAction: MenuActionId = 'compose'
  let formIconName = ''
  let formIconNodes: IconNode[] | undefined
  let pickingIcon = false

  function toggleExpand(id: string): void {
    const next = new Set(expanded)
    if (next.has(id)) {
      next.delete(id)
    } else {
      next.add(id)
    }
    expanded = next
  }

  // --- drag and drop ---

  function menusConsider(e: CustomEvent<DndEvent<MenuLayout>>): void {
    menuList = e.detail.items
  }
  function menusFinalize(e: CustomEvent<DndEvent<MenuLayout>>): void {
    menuList = e.detail.items
    setMenus(menuList)
  }
  function itemsConsider(menuId: string, e: CustomEvent<DndEvent<MenuItemLayout>>): void {
    menuList = menuList.map((m) => (m.id === menuId ? { ...m, items: e.detail.items } : m))
  }
  function itemsFinalize(menuId: string, e: CustomEvent<DndEvent<MenuItemLayout>>): void {
    menuList = menuList.map((m) => (m.id === menuId ? { ...m, items: e.detail.items } : m))
    setMenus(menuList)
  }

  // --- labels ---

  function menuLabel(menu: MenuLayout): string {
    return menu.labelKey ? $t(menu.labelKey) : (menu.label ?? '')
  }
  function itemLabel(item: MenuItemLayout): string {
    if (item.kind === 'custom') {
      return item.label ?? ''
    }
    const def = item.action ? catalogByAction[item.action] : undefined
    return def ? $t(def.labelKey) : ''
  }
  function itemDef(item: MenuItemLayout) {
    return item.action ? catalogByAction[item.action] : undefined
  }
  function isMacOnly(item: MenuItemLayout): boolean {
    const def = itemDef(item)
    return !!def?.macOnly && !isMac
  }

  // --- add custom entry ---

  function openAddForm(menuId: string): void {
    addingTo = menuId
    formLabel = ''
    formAction = actionOptions[0]?.action ?? 'compose'
    formIconName = ''
    formIconNodes = undefined
    pickingIcon = false
  }
  function cancelAddForm(): void {
    addingTo = null
    pickingIcon = false
  }
  function onIconSelect(e: CustomEvent<{ iconName: string; iconNodes?: IconNode[] }>): void {
    formIconName = e.detail.iconName
    formIconNodes = e.detail.iconNodes
    pickingIcon = false
  }
  function onIconClear(): void {
    formIconName = ''
    formIconNodes = undefined
    pickingIcon = false
  }
  function submitAddForm(menuId: string): void {
    const label = formLabel.trim()
    if (label === '') {
      return
    }
    addCustomItem(menuId, {
      label,
      action: formAction,
      iconName: formIconName || undefined,
      iconNodes: formIconNodes,
    })
    cancelAddForm()
  }

  // --- add custom menu ---

  function addMenu(): void {
    const id = addCustomMenu($t('menuBar.newMenu'))
    expanded = new Set(expanded).add(id)
  }

  function doReset(): void {
    resetLayout()
    confirmingReset = false
    expanded = new Set()
  }
</script>

<div class="customizer">
  <p class="hint">{$t('menuBar.customize.hint')}</p>

  <section
    class="menus"
    use:dndzone={{ items: menuList, type: 'menus', flipDurationMs: flipMs, dropTargetStyle: {} }}
    on:consider={menusConsider}
    on:finalize={menusFinalize}
  >
    {#each menuList as menu (menu.id)}
      <div class="menu" class:dimmed={menu.hidden} animate:flip={{ duration: flipMs }}>
        <div class="menu-head">
          <span class="grip" aria-hidden="true"><IconGripVertical size={16} stroke={1.7} /></span>
          <button type="button" class="expander" on:click={() => toggleExpand(menu.id)} aria-expanded={expanded.has(menu.id)}>
            {#if expanded.has(menu.id)}
              <IconChevronDown size={16} stroke={1.7} />
            {:else}
              <IconChevronRight size={16} stroke={1.7} />
            {/if}
          </button>

          {#if menu.builtin}
            <span class="menu-title">{menuLabel(menu)}</span>
          {:else}
            <input
              class="menu-title-input"
              type="text"
              value={menu.label ?? ''}
              aria-label={$t('menuBar.menuName')}
              on:input={(e) => renameCustomMenu(menu.id, e.currentTarget.value)}
            />
          {/if}

          <span class="count">{menu.items.length}</span>

          <button
            type="button"
            class="icon-btn"
            title={menu.hidden ? $t('menuBar.show') : $t('menuBar.hide')}
            on:click={() => toggleMenuHidden(menu.id)}
          >
            {#if menu.hidden}<IconEyeOff size={16} stroke={1.7} />{:else}<IconEye size={16} stroke={1.7} />{/if}
          </button>
          {#if !menu.builtin}
            <button type="button" class="icon-btn danger" title={$t('menuBar.removeMenu')} on:click={() => removeMenu(menu.id)}>
              <IconTrash size={16} stroke={1.7} />
            </button>
          {/if}
        </div>

        {#if expanded.has(menu.id)}
          <div
            class="items"
            use:dndzone={{ items: menu.items, type: 'menu-items', flipDurationMs: flipMs, dropTargetStyle: {} }}
            on:consider={(e) => itemsConsider(menu.id, e)}
            on:finalize={(e) => itemsFinalize(menu.id, e)}
          >
            {#each menu.items as item (item.id)}
              {@const def = itemDef(item)}
              <div class="item" class:dimmed={item.hidden} animate:flip={{ duration: flipMs }}>
                <span class="grip" aria-hidden="true"><IconGripVertical size={15} stroke={1.7} /></span>

                {#if item.kind === 'separator'}
                  <span class="sep-glyph"><IconSeparatorHorizontal size={15} stroke={1.7} /></span>
                  <span class="item-label muted">{$t('menuBar.separator')}</span>
                {:else}
                  <span class="item-glyph">
                    {#if item.kind === 'custom'}
                      <MenuGlyph iconName={item.iconName} iconNodes={item.iconNodes} size={15} />
                    {:else if def}
                      <ThemedIcon name={def.iconName} icon={def.icon} size={15} stroke={1.7} />
                    {/if}
                  </span>
                  <span class="item-label">{itemLabel(item)}</span>
                  {#if isMacOnly(item)}<span class="badge">{$t('menuBar.macOnly')}</span>{/if}
                {/if}

                <span class="spacer"></span>

                {#if item.kind !== 'separator'}
                  <button
                    type="button"
                    class="icon-btn"
                    title={item.hidden ? $t('menuBar.show') : $t('menuBar.hide')}
                    on:click={() => toggleItemHidden(menu.id, item.id)}
                  >
                    {#if item.hidden}<IconEyeOff size={15} stroke={1.7} />{:else}<IconEye size={15} stroke={1.7} />{/if}
                  </button>
                {/if}
                {#if item.kind !== 'action'}
                  <button type="button" class="icon-btn danger" title={$t('menuBar.remove')} on:click={() => removeItem(menu.id, item.id)}>
                    <IconTrash size={15} stroke={1.7} />
                  </button>
                {/if}
              </div>
            {/each}
          </div>

          {#if addingTo === menu.id}
            <div class="add-form">
              <div class="add-row">
                <input
                  class="text-input"
                  type="text"
                  bind:value={formLabel}
                  placeholder={$t('menuBar.entryLabel')}
                  aria-label={$t('menuBar.entryLabel')}
                />
                <select class="select" bind:value={formAction} aria-label={$t('menuBar.entryAction')}>
                  {#each actionOptions as opt (opt.action)}
                    <option value={opt.action}>{$t(opt.labelKey)}</option>
                  {/each}
                </select>
                <button type="button" class="icon-choose" on:click={() => (pickingIcon = !pickingIcon)}>
                  {#if formIconName || formIconNodes}
                    <MenuGlyph iconName={formIconName} iconNodes={formIconNodes} size={16} />
                  {:else}
                    <IconPhoto size={16} stroke={1.7} />
                  {/if}
                  <span>{$t('menuBar.icon.choose')}</span>
                </button>
              </div>

              {#if pickingIcon}
                <IconPicker on:select={onIconSelect} on:clear={onIconClear} />
              {/if}

              <div class="add-actions">
                <button type="button" class="btn ghost" on:click={cancelAddForm}>
                  <IconX size={15} stroke={1.7} /><span>{$t('menuBar.cancel')}</span>
                </button>
                <button type="button" class="btn primary" disabled={formLabel.trim() === ''} on:click={() => submitAddForm(menu.id)}>
                  <IconCheck size={15} stroke={1.7} /><span>{$t('menuBar.add')}</span>
                </button>
              </div>
            </div>
          {:else}
            <div class="menu-actions">
              <button type="button" class="mini" on:click={() => openAddForm(menu.id)}>
                <IconPlus size={14} stroke={1.7} /><span>{$t('menuBar.addEntry')}</span>
              </button>
              <button type="button" class="mini" on:click={() => addSeparator(menu.id)}>
                <IconSeparatorHorizontal size={14} stroke={1.7} /><span>{$t('menuBar.addSeparator')}</span>
              </button>
            </div>
          {/if}
        {/if}
      </div>
    {/each}
  </section>

  <div class="toolbar">
    <button type="button" class="btn" on:click={addMenu}>
      <IconPlus size={15} stroke={1.7} /><span>{$t('menuBar.addMenu')}</span>
    </button>
    {#if confirmingReset}
      <div class="reset-confirm">
        <span>{$t('menuBar.resetConfirm')}</span>
        <button type="button" class="btn ghost" on:click={() => (confirmingReset = false)}>{$t('menuBar.cancel')}</button>
        <button type="button" class="btn danger" on:click={doReset}>{$t('menuBar.reset')}</button>
      </div>
    {:else}
      <button type="button" class="btn ghost" on:click={() => (confirmingReset = true)}>
        <IconRotate size={15} stroke={1.7} /><span>{$t('menuBar.reset')}</span>
      </button>
    {/if}
  </div>

  <div class="new-items">
    <span class="row-label">{$t('menuBar.newItems.label')}</span>
    <div class="segment" role="group" aria-label={$t('menuBar.newItems.label')}>
      {#each [['visible', $t('menuBar.newItems.visible')], ['hidden', $t('menuBar.newItems.hidden')]] as [mode, text] (mode)}
        <button
          type="button"
          class="seg-btn"
          class:active={$menuBarNewItems === mode}
          on:click={() => setNewItemsMode(mode as NewItemsMode)}
        >
          {text}
        </button>
      {/each}
    </div>
  </div>
  <p class="hint">{$t('menuBar.newItems.hint')}</p>
</div>

<style>
  .customizer {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-top: var(--space-2);
  }

  .menus {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .menu {
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-card);
    background: var(--surface-sunken);
  }

  .menu.dimmed > .menu-head .menu-title,
  .item.dimmed .item-label {
    opacity: 0.45;
  }

  .menu-head {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2);
  }

  .menu-title {
    font-size: var(--fz-label);
    color: var(--text-primary);
    font-weight: 500;
  }

  .menu-title-input {
    flex: 1;
    min-width: 0;
    padding: var(--space-1) var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
  }

  .count {
    margin-left: auto;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .expander {
    display: inline-flex;
    padding: 0;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
  }

  .grip {
    display: inline-flex;
    color: var(--text-tertiary);
    cursor: grab;
  }

  .items {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: 0 var(--space-2) var(--space-2) var(--space-5);
  }

  .item {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
  }

  .item-glyph,
  .sep-glyph {
    display: inline-flex;
    color: var(--text-tertiary);
    flex: none;
  }

  .item-label {
    font-size: var(--fz-label);
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .item-label.muted {
    color: var(--text-tertiary);
    font-style: italic;
  }

  .badge {
    padding: 0 var(--space-1);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-control);
  }

  .spacer {
    flex: 1;
  }

  .icon-btn {
    display: inline-flex;
    padding: var(--space-1);
    border: none;
    background: transparent;
    color: var(--text-secondary);
    border-radius: var(--radius-control);
    cursor: pointer;
  }

  .icon-btn:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .icon-btn.danger:hover {
    color: var(--danger);
  }

  .menu-actions {
    display: flex;
    gap: var(--space-2);
    padding: 0 var(--space-2) var(--space-2) var(--space-5);
  }

  .mini,
  .icon-choose {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1) var(--space-2);
    border: var(--hairline) solid var(--border-default);
    background: var(--surface-raised);
    color: var(--text-secondary);
    font-size: var(--fz-meta);
    border-radius: var(--radius-control);
    cursor: pointer;
  }

  .mini:hover,
  .icon-choose:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .add-form {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-2) var(--space-3) var(--space-5);
  }

  .add-row {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    flex-wrap: wrap;
  }

  .text-input,
  .select {
    padding: var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
  }

  .text-input {
    flex: 1;
    min-width: 140px;
  }

  .add-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
  }

  .toolbar,
  .new-items {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
  }

  .new-items {
    justify-content: space-between;
  }

  .row-label {
    font-size: var(--fz-label);
    color: var(--text-primary);
  }

  .btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
    border-radius: var(--radius-control);
    cursor: pointer;
  }

  .btn:hover:not(:disabled) {
    background: var(--surface-hover);
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .btn.ghost {
    background: transparent;
    color: var(--text-secondary);
  }

  .btn.primary {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--accent-fg);
  }

  .btn.danger {
    border-color: var(--danger);
    color: var(--danger);
  }

  .reset-confirm {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .segment {
    display: inline-flex;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    overflow: hidden;
  }

  .seg-btn {
    padding: var(--space-1) var(--space-3);
    border: none;
    background: var(--surface-raised);
    color: var(--text-secondary);
    font-size: var(--fz-meta);
    cursor: pointer;
  }

  .seg-btn.active {
    background: var(--accent);
    color: var(--accent-fg);
  }

  .hint {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    margin: 0;
  }
</style>
