<script lang="ts">
  // the custom right-click menu, mounted once at the app root. it positions
  // itself at the stored coordinates, flips to stay on screen, and closes on any
  // outside interaction. it is the only context menu in the app; the webview's
  // default is suppressed globally.
  import { tick } from 'svelte'
  import { contextMenu, closeContextMenu, type MenuItem } from '../../stores/contextmenu'

  let menuEl: HTMLDivElement
  let left = 0
  let top = 0

  // reposition when the menu opens, keeping it inside the viewport.
  $: if ($contextMenu) {
    void place($contextMenu.x, $contextMenu.y)
  }

  async function place(x: number, y: number): Promise<void> {
    left = x
    top = y
    await tick()
    if (!menuEl) {
      return
    }
    const rect = menuEl.getBoundingClientRect()
    const pad = 8
    if (x + rect.width + pad > window.innerWidth) {
      left = Math.max(pad, window.innerWidth - rect.width - pad)
    }
    if (y + rect.height + pad > window.innerHeight) {
      top = Math.max(pad, window.innerHeight - rect.height - pad)
    }
  }

  function run(item: MenuItem): void {
    closeContextMenu()
    item.action()
  }
</script>

{#if $contextMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="scrim" on:click={closeContextMenu} on:contextmenu|preventDefault={closeContextMenu}></div>
  <div class="menu" bind:this={menuEl} style={`left:${left}px;top:${top}px`} role="menu">
    {#each $contextMenu.entries as entry}
      {#if entry === 'separator'}
        <div class="sep" role="separator"></div>
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
</style>
