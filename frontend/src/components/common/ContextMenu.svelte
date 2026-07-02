<script lang="ts">
  // the custom right-click menu, mounted once at the app root. it positions
  // itself at the stored coordinates, flips to stay on screen, and closes on any
  // outside interaction. it is the only context menu in the app; the webview's
  // default is suppressed globally.
  import { tick } from 'svelte'
  import { IconX } from '@tabler/icons-svelte'
  import { contextMenu, closeContextMenu, type MenuItem } from '../../stores/contextmenu'
  import { flagColors } from '../../theme/flagcolors'
  import { t } from '../../lib/i18n'

  let menuEl: HTMLDivElement
  let left = 0
  let top = 0

  // reposition when the menu opens, keeping it inside the viewport.
  $: if ($contextMenu) {
    void place($contextMenu.x, $contextMenu.y)
  }

  // the app applies an interface zoom via css `zoom` on <html>. under zoom,
  // pointer clientX/Y stay in unscaled screen pixels while a fixed element is
  // positioned in the zoomed layout space, so the raw coordinates land the menu
  // in the wrong place. convert cursor and viewport into layout space by dividing
  // by the scale (a no-op at 100%). offsetWidth/Height are already layout-space.
  function uiScale(): number {
    const raw = getComputedStyle(document.documentElement).getPropertyValue('--ui-scale')
    const n = parseFloat(raw)
    return n > 0 ? n : 1
  }

  async function place(x: number, y: number): Promise<void> {
    const scale = uiScale()
    const mx = x / scale
    const my = y / scale
    left = mx
    top = my
    await tick()
    if (!menuEl) {
      return
    }
    const pad = 8
    const vw = window.innerWidth / scale
    const vh = window.innerHeight / scale
    const w = menuEl.offsetWidth
    const h = menuEl.offsetHeight
    if (mx + w + pad > vw) {
      left = Math.max(pad, vw - w - pad)
    }
    if (my + h + pad > vh) {
      top = Math.max(pad, vh - h - pad)
    }
  }

  function run(item: MenuItem): void {
    closeContextMenu()
    item.action()
  }

  // pickColor fires the swatch row's callback and closes the menu.
  function pickColor(onPick: (color: number) => void, color: number): void {
    closeContextMenu()
    onPick(color)
  }
</script>

{#if $contextMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="scrim" on:click={closeContextMenu} on:contextmenu|preventDefault={closeContextMenu}></div>
  <div class="menu" bind:this={menuEl} style={`left:${left}px;top:${top}px`} role="menu">
    {#each $contextMenu.entries as entry}
      {#if entry === 'separator'}
        <div class="sep" role="separator"></div>
      {:else if 'kind' in entry}
        <div class="swatches" role="group" aria-label={$t('common.contextMenu.flagColor')}>
          <button
            type="button"
            class="swatch clear"
            class:active={entry.current === 0}
            title={$t('common.contextMenu.noColor')}
            aria-label={$t('common.contextMenu.noColor')}
            on:click={() => pickColor(entry.onPick, 0)}
          >
            <IconX size={12} stroke={2} />
          </button>
          {#each flagColors as c}
            <button
              type="button"
              class="swatch"
              class:active={entry.current === c.index}
              style={`--sw:${c.hex}`}
              title={c.name}
              aria-label={c.name}
              on:click={() => pickColor(entry.onPick, c.index)}
            ></button>
          {/each}
        </div>
      {:else}
        <button type="button" class="item" class:danger={entry.danger} role="menuitem" on:click={() => run(entry)}>
          {#if entry.icon}
            <span class="icon"><svelte:component this={entry.icon} size={15} stroke={1.7} /></span>
          {/if}
          <span class="label">{entry.label}</span>
        </button>
      {/if}
    {/each}
  </div>
{/if}

<style>
  .scrim {
    position: fixed;
    inset: 0;
    z-index: 250;
  }

  .menu {
    position: fixed;
    z-index: 251;
    min-width: 180px;
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

  .item:hover {
    background: var(--surface-hover);
  }

  .item.danger {
    color: var(--danger);
  }

  .icon {
    display: inline-flex;
    color: var(--text-tertiary);
    flex-shrink: 0;
  }

  .item.danger .icon {
    color: var(--danger);
  }

  .item:hover .icon {
    color: inherit;
  }

  .sep {
    height: var(--hairline);
    margin: var(--space-1) var(--space-2);
    background: var(--border-subtle);
  }

  .swatches {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    flex-wrap: wrap;
  }

  .swatch {
    width: 18px;
    height: 18px;
    border-radius: 999px;
    border: var(--hairline) solid var(--border-default);
    background: var(--sw, transparent);
    cursor: pointer;
    padding: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--text-tertiary);
    transition: transform 0.08s ease;
  }

  .swatch:hover {
    transform: scale(1.18);
  }

  .swatch.active {
    box-shadow: 0 0 0 2px var(--accent);
  }

  .swatch.clear {
    background: var(--surface-hover);
    border-color: var(--border-default);
    color: var(--text-secondary);
  }

  /* the clear (no color) swatch is active by default; keep its ring subtle so the
     X does not read as an odd outlined glyph. */
  .swatch.clear.active {
    box-shadow: none;
    border-color: var(--text-tertiary);
  }
</style>
