<script lang="ts">
  // the in-app menu bar: the top row of the ui on Windows/Linux (where the
  // native menu bar cannot follow the app theme and, on Linux, crashed GTK on
  // rebuild), opt-in on macOS. it mirrors the native menu's structure and emits
  // the same action strings the native menu did, so App.svelte handles both
  // through one dispatcher. labels come from the frontend i18n and live-update
  // on a language change; item state follows the open-message selection
  // directly instead of round-tripping through the backend.
  import { createEventDispatcher, type ComponentType } from 'svelte'
  import {
    IconInfoCircle,
    IconSettings,
    IconEyeOff,
    IconPower,
    IconPencil,
    IconFileTypePdf,
    IconRefresh,
    IconPlus,
    IconMailbox,
    IconArrowBackUp,
    IconMailOpened,
    IconMail,
    IconFlag,
    IconArchive,
    IconTrash,
    IconMaximize,
    IconBatteryEco,
  } from '@tabler/icons-svelte'
  import ThemedIcon from './ThemedIcon.svelte'
  import { t, isMac, shortcutLabel } from '../../lib/i18n'
  import { prefs } from '../../stores/prefs'
  import { bindings } from '../../stores/shortcuts'
  import { openMessageId } from '../../stores/selection'

  const dispatch = createEventDispatcher<{ action: string }>()

  interface Item {
    action: string
    labelKey: string
    icon: ComponentType
    iconName: string
    // hint is a fixed combo shown as the shortcut hint; unset items look their
    // hint up in the rebindable shortcut map under their action name.
    hint?: string
    // needsMessage items are disabled while no message is open.
    needsMessage?: boolean
    danger?: boolean
  }
  type Entry = Item | 'separator'

  interface MenuDef {
    key: string
    label: string
    entries: Entry[]
  }

  // the app menu's Hide item only exists on macOS (WindowHide there is the
  // standard Cmd+H behavior; on Windows/Linux the titlebar minimize covers it).
  // rebuilt reactively so labels follow a live language change.
  $: menus = buildMenus($t)

  function buildMenus(tFn: (key: string) => string): MenuDef[] {
    return [
    {
      key: 'app',
      label: 'Pelton',
      entries: [
        { action: 'about', labelKey: 'menu.about', icon: IconInfoCircle, iconName: 'info-circle' },
        'separator',
        { action: 'preferences', labelKey: 'menu.preferences', icon: IconSettings, iconName: 'settings' },
        'separator',
        ...(isMac
          ? [{ action: 'hide-window', labelKey: 'menu.hide', icon: IconEyeOff, iconName: 'eye-off', hint: 'mod+h' } as Item]
          : []),
        { action: 'quit', labelKey: 'menu.quit', icon: IconPower, iconName: 'power', hint: isMac ? 'mod+q' : undefined },
      ],
    },
    {
      key: 'file',
      label: tFn('menu.file'),
      entries: [
        { action: 'compose', labelKey: 'menu.compose', icon: IconPencil, iconName: 'pencil' },
        'separator',
        { action: 'export-pdf', labelKey: 'menu.exportPdf', icon: IconFileTypePdf, iconName: 'file-type-pdf' },
      ],
    },
    {
      key: 'mailbox',
      label: tFn('menu.mailbox'),
      entries: [
        { action: 'sync', labelKey: 'menu.sync', icon: IconRefresh, iconName: 'refresh' },
        'separator',
        { action: 'add-mailbox', labelKey: 'menu.addMailbox', icon: IconPlus, iconName: 'plus' },
        { action: 'open-mailboxes', labelKey: 'menu.manageMailboxes', icon: IconMailbox, iconName: 'mailbox' },
      ],
    },
    {
      key: 'mail',
      label: tFn('menu.mail'),
      entries: [
        { action: 'undo', labelKey: 'menu.undo', icon: IconArrowBackUp, iconName: 'arrow-back-up', hint: 'mod+z' },
        'separator',
        { action: 'mark-read', labelKey: 'menu.markRead', icon: IconMailOpened, iconName: 'mail-opened', needsMessage: true },
        { action: 'mark-unread', labelKey: 'menu.markUnread', icon: IconMail, iconName: 'mail', needsMessage: true },
        { action: 'flag', labelKey: 'menu.flag', icon: IconFlag, iconName: 'flag', needsMessage: true },
        { action: 'archive', labelKey: 'menu.archive', icon: IconArchive, iconName: 'archive', needsMessage: true },
        { action: 'delete-message', labelKey: 'menu.deleteMessage', icon: IconTrash, iconName: 'trash', needsMessage: true, danger: true },
      ],
    },
    {
      key: 'view',
      label: tFn('menu.view'),
      entries: [
        { action: 'toggle-fullscreen', labelKey: 'menu.toggleFullscreen', icon: IconMaximize, iconName: 'maximize' },
        'separator',
        { action: 'toggle-low-power', labelKey: 'menu.lowPower', icon: IconBatteryEco, iconName: 'battery-eco' },
      ],
    },
    ]
  }

  let openKey: string | null = null
  let barEl: HTMLElement

  // hintFor resolves an item's shortcut hint: a fixed combo if the item
  // declares one, otherwise the live (user-rebindable) binding for its action.
  function hintFor(item: Item, map: Record<string, string>): string {
    const combo = item.hint ?? map[item.action] ?? ''
    return combo ? shortcutLabel(combo) : ''
  }

  function isDisabled(item: Item, openId: number | null): boolean {
    return !!item.needsMessage && openId === null
  }

  function toggle(key: string): void {
    openKey = openKey === key ? null : key
  }

  // while a menu is open, hovering another top-level title switches to it, the
  // way native menu bars roll.
  function hoverTitle(key: string): void {
    if (openKey !== null) {
      openKey = key
    }
  }

  function close(): void {
    openKey = null
  }

  function run(item: Item): void {
    if (isDisabled(item, $openMessageId)) {
      return
    }
    close()
    dispatch('action', item.action)
  }

  // dropdownItems lists the focusable buttons of the open dropdown for the
  // arrow-key navigation below.
  function dropdownItems(): HTMLButtonElement[] {
    return Array.from(barEl?.querySelectorAll<HTMLButtonElement>('.dropdown .item:not([disabled])') ?? [])
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
    const keys = menus.map((m) => m.key)
    const next = (keys.indexOf(openKey) + delta + keys.length) % keys.length
    openKey = keys[next]
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

{#if openKey !== null}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="scrim" on:click={close} on:contextmenu|preventDefault={close}></div>
{/if}

<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
<nav class="menubar" aria-label="Pelton" bind:this={barEl} on:keydown={onBarKeydown}>
  {#each menus as m (m.key)}
    <div class="menu-wrap">
      <button
        type="button"
        class="title"
        class:open={openKey === m.key}
        role="menuitem"
        aria-haspopup="menu"
        aria-expanded={openKey === m.key}
        on:click={() => toggle(m.key)}
        on:mouseenter={() => hoverTitle(m.key)}
      >
        {m.label}
      </button>
      {#if openKey === m.key}
        <div class="dropdown" role="menu" aria-label={m.label}>
          {#each m.entries as entry}
            {#if entry === 'separator'}
              <div class="sep" role="separator"></div>
            {:else}
              <button
                type="button"
                class="item"
                class:danger={entry.danger}
                role="menuitem"
                disabled={isDisabled(entry, $openMessageId)}
                on:click={() => run(entry)}
              >
                {#if $prefs.menuBarIcons}
                  <span class="icon"><ThemedIcon name={entry.iconName} icon={entry.icon} size={15} stroke={1.7} /></span>
                {/if}
                <span class="label">{$t(entry.labelKey)}</span>
                {#if hintFor(entry, $bindings)}
                  <span class="hint">{hintFor(entry, $bindings)}</span>
                {/if}
              </button>
            {/if}
          {/each}
        </div>
      {/if}
    </div>
  {/each}
</nav>

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
    /* above the scrim, so while a menu is open the other titles still take
       hover (roll-over to the neighboring menu, like a native bar) and
       clicks; the scrim only catches clicks outside the bar. */
    position: relative;
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
