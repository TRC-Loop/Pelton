<script lang="ts">
  // the in-app menu bar: the top row of the ui on Windows/Linux (where the
  // native menu bar cannot follow the app theme and, on Linux, crashed GTK on
  // rebuild), opt-in on macOS. it emits the same action strings the native menu
  // did, so App.svelte handles both through one dispatcher. its structure is the
  // user-customizable layout from stores/menubar.ts, resolved to render-ready
  // menus with one level of submenus; labels come from the frontend i18n
  // (built-ins) or the user's text (custom entries) and live-update on a language
  // change; item state follows the open-message selection directly. In editor
  // mode the bar hands off to MenuBarEditor for in-place customization.
  import { createEventDispatcher } from 'svelte'
  import ThemedIcon from './ThemedIcon.svelte'
  import MenuGlyph from './MenuGlyph.svelte'
  import MenuBarEditor from './MenuBarEditor.svelte'
  import { IconChevronRight } from '@tabler/icons-svelte'
  import { t, shortcutLabel } from '../../lib/i18n'
  import { prefs } from '../../stores/prefs'
  import { bindings } from '../../stores/shortcuts'
  import { openMessageId } from '../../stores/selection'
  import { menuBarLayout, menuBarEditing, resolveBar, type RenderItem } from '../../stores/menubar'

  const dispatch = createEventDispatcher<{ action: string }>()

  // the render-ready menus, reactive to layout edits. i18n labels resolve in the
  // template so a language switch updates them without rebuilding.
  $: menus = resolveBar($menuBarLayout)

  let openKey: string | null = null
  let openSub: string | null = null
  let barEl: HTMLElement

  function itemLabel(item: RenderItem, tFn: (key: string) => string): string {
    if (item.kind === 'custom' || item.kind === 'submenu') {
      return item.label ?? ''
    }
    return item.labelKey ? tFn(item.labelKey) : ''
  }

  // hintFor resolves an item's shortcut hint: a fixed combo if the item declares
  // one, otherwise the live (user-rebindable) binding for its action.
  function hintFor(item: RenderItem, map: Record<string, string>): string {
    const combo = item.hint ?? (item.action ? map[item.action] : undefined) ?? ''
    return combo ? shortcutLabel(combo) : ''
  }

  function isDisabled(item: RenderItem, openId: number | null): boolean {
    return !!item.needsMessage && openId === null
  }

  function toggle(key: string): void {
    openKey = openKey === key ? null : key
    openSub = null
  }

  // while a menu is open, hovering another top-level title switches to it, the
  // way native menu bars roll.
  function hoverTitle(key: string): void {
    if (openKey !== null) {
      openKey = key
      openSub = null
    }
  }

  function close(): void {
    openKey = null
    openSub = null
  }

  function run(item: RenderItem): void {
    if (isDisabled(item, $openMessageId) || !item.action) {
      return
    }
    close()
    dispatch('action', item.action)
  }

  // dropdownItems lists the focusable buttons of the open dropdown for the
  // arrow-key navigation below (submenu triggers included, flyout items excluded
  // so the top list stays predictable).
  function dropdownItems(): HTMLButtonElement[] {
    return Array.from(barEl?.querySelectorAll<HTMLButtonElement>('.dropdown > .item, .dropdown > .sub-wrap > .item') ?? [])
      .filter((el) => !el.hasAttribute('disabled'))
  }

  function focusItem(delta: number): void {
    const items = dropdownItems()
    if (items.length === 0) {
      return
    }
    const idx = items.findIndex((el) => el === document.activeElement)
    const next = idx === -1 ? (delta > 0 ? 0 : items.length - 1) : (idx + delta + items.length) % items.length
    items[next].focus()
  }

  function switchMenu(delta: number): void {
    if (openKey === null) {
      return
    }
    const keys = menus.map((m) => m.id)
    const next = (keys.indexOf(openKey) + delta + keys.length) % keys.length
    openKey = keys[next]
    openSub = null
  }

  function onBarKeydown(event: KeyboardEvent): void {
    if (openKey === null) {
      return
    }
    switch (event.key) {
      case 'Escape':
        close()
        break
      case 'ArrowDown':
        focusItem(1)
        break
      case 'ArrowUp':
        focusItem(-1)
        break
      case 'ArrowLeft':
        switchMenu(-1)
        break
      case 'ArrowRight':
        switchMenu(1)
        break
      default:
        return
    }
    event.preventDefault()
    event.stopPropagation()
  }
</script>

{#if $menuBarEditing}
  <MenuBarEditor />
{:else}
  {#if openKey !== null}
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
    <div class="scrim" on:click={close} on:contextmenu|preventDefault={close}></div>
  {/if}

  <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
  <nav class="menubar" class:raised={openKey !== null} aria-label="Pelton" bind:this={barEl} on:keydown={onBarKeydown}>
    {#each menus as m (m.id)}
      {@const label = m.labelKey ? $t(m.labelKey) : (m.label ?? '')}
      <div class="menu-wrap">
        <button
          type="button"
          class="title"
          class:open={openKey === m.id}
          role="menuitem"
          aria-haspopup="menu"
          aria-expanded={openKey === m.id}
          on:click={() => toggle(m.id)}
          on:mouseenter={() => hoverTitle(m.id)}
        >
          {label}
        </button>
        {#if openKey === m.id}
          <div class="dropdown" role="menu" aria-label={label}>
            {#each m.items as item (item.id)}
              {#if item.kind === 'separator'}
                <div class="sep" role="separator"></div>
              {:else if item.kind === 'submenu'}
                <!-- svelte-ignore a11y-no-static-element-interactions -->
                <div
                  class="sub-wrap"
                  on:mouseenter={() => (openSub = item.id)}
                  on:mouseleave={() => (openSub = null)}
                >
                  <button
                    type="button"
                    class="item"
                    role="menuitem"
                    aria-haspopup="menu"
                    aria-expanded={openSub === item.id}
                    on:click={() => (openSub = openSub === item.id ? null : item.id)}
                    on:focus={() => (openSub = item.id)}
                  >
                    {#if $prefs.menuBarIcons}
                      <span class="icon"><MenuGlyph iconName={item.iconName} iconNodes={item.iconNodes} size={15} stroke={1.7} /></span>
                    {/if}
                    <span class="label">{item.label ?? ''}</span>
                    <span class="chevron"><IconChevronRight size={15} stroke={1.7} /></span>
                  </button>
                  {#if openSub === item.id && item.items}
                    <div class="flyout" role="menu" aria-label={item.label ?? ''}>
                      {#each item.items as sub (sub.id)}
                        {#if sub.kind === 'separator'}
                          <div class="sep" role="separator"></div>
                        {:else}
                          <button
                            type="button"
                            class="item"
                            class:danger={sub.danger}
                            role="menuitem"
                            disabled={isDisabled(sub, $openMessageId)}
                            on:click={() => run(sub)}
                          >
                            {#if $prefs.menuBarIcons}
                              <span class="icon">
                                {#if sub.kind === 'custom'}
                                  <MenuGlyph iconName={sub.iconName} iconNodes={sub.iconNodes} size={15} stroke={1.7} />
                                {:else if sub.icon}
                                  <ThemedIcon name={sub.iconName ?? ''} icon={sub.icon} size={15} stroke={1.7} />
                                {/if}
                              </span>
                            {/if}
                            <span class="label">{itemLabel(sub, $t)}</span>
                            {#if hintFor(sub, $bindings)}
                              <span class="hint">{hintFor(sub, $bindings)}</span>
                            {/if}
                          </button>
                        {/if}
                      {/each}
                    </div>
                  {/if}
                </div>
              {:else}
                <button
                  type="button"
                  class="item"
                  class:danger={item.danger}
                  role="menuitem"
                  disabled={isDisabled(item, $openMessageId)}
                  on:click={() => run(item)}
                >
                  {#if $prefs.menuBarIcons}
                    <span class="icon">
                      {#if item.kind === 'custom'}
                        <MenuGlyph iconName={item.iconName} iconNodes={item.iconNodes} size={15} stroke={1.7} />
                      {:else if item.icon}
                        <ThemedIcon name={item.iconName ?? ''} icon={item.icon} size={15} stroke={1.7} />
                      {/if}
                    </span>
                  {/if}
                  <span class="label">{itemLabel(item, $t)}</span>
                  {#if hintFor(item, $bindings)}
                    <span class="hint">{hintFor(item, $bindings)}</span>
                  {/if}
                </button>
              {/if}
            {/each}
          </div>
        {/if}
      </div>
    {/each}
  </nav>
{/if}

<style>
  .menubar {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    height: 30px;
    padding: 0 var(--space-2);
    background: var(--surface-sunken);
    border-bottom: var(--hairline) solid var(--border-subtle);
    user-select: none;
    position: relative;
  }

  /* only while a menu is open: above the scrim, so the other titles still
     take hover (roll-over to the neighboring menu, like a native bar) and
     clicks. at rest the bar stays in normal stacking order and never covers
     overlays like the settings screen. */
  .menubar.raised {
    z-index: 220;
  }

  .menu-wrap {
    position: relative;
  }

  .title {
    padding: var(--space-1) var(--space-3);
    border: none;
    background: transparent;
    color: var(--text-secondary);
    font-size: var(--fz-label);
    border-radius: var(--radius-control);
    cursor: default;
  }

  .title:hover,
  .title.open {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .scrim {
    position: fixed;
    inset: 0;
    z-index: 219;
  }

  .dropdown {
    position: absolute;
    top: calc(100% + var(--space-1));
    left: 0;
    z-index: 220;
    min-width: 220px;
    padding: var(--space-1);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
  }

  .sub-wrap {
    position: relative;
  }

  .flyout {
    position: absolute;
    top: calc(-1 * var(--space-1));
    left: 100%;
    z-index: 221;
    min-width: 200px;
    padding: var(--space-1);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
  }

  .item {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    padding: var(--space-2) var(--space-3);
    border: none;
    background: transparent;
    color: var(--text-primary);
    cursor: pointer;
    text-align: left;
    font-size: var(--fz-label);
    border-radius: var(--radius-control);
  }

  .item:hover:not(:disabled),
  .item:focus-visible {
    background: var(--surface-hover);
    outline: none;
  }

  .item:disabled {
    color: var(--text-tertiary);
    cursor: default;
  }

  .item.danger:not(:disabled) {
    color: var(--danger);
  }

  .icon {
    display: inline-flex;
    color: var(--text-tertiary);
    flex-shrink: 0;
  }

  .item.danger:not(:disabled) .icon {
    color: var(--danger);
  }

  .item:hover:not(:disabled) .icon {
    color: inherit;
  }

  .label {
    flex: 1;
    white-space: nowrap;
  }

  .chevron {
    display: inline-flex;
    color: var(--text-tertiary);
    flex-shrink: 0;
  }

  .hint {
    margin-left: var(--space-4);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    white-space: nowrap;
  }

  .sep {
    height: var(--hairline);
    margin: var(--space-1) var(--space-2);
    background: var(--border-subtle);
  }
</style>
