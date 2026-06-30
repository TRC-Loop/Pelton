<script lang="ts">
  // one selectable row in the sidebar: an optional disclosure caret, an icon, a
  // label, and an unread count. active rows use the accent selection background;
  // unread is shown by weight and the count, never by accent color.
  import { createEventDispatcher } from 'svelte'
  import { IconChevronRight } from '@tabler/icons-svelte'
  import { prefs } from '../../stores/prefs'

  export let label: string
  export let count: number = 0
  export let active: boolean = false
  export let depth: number = 0
  export let expandable: boolean = false
  export let expanded: boolean = false

  const dispatch = createEventDispatcher<{ select: void; toggle: void }>()

  // indent nested folders by depth. the base inset keeps the caret aligned.
  $: indent = `calc(var(--space-3) + ${depth} * var(--space-4))`
  // one vertical guide line per ancestor level when the user enables them, each
  // aligned under that ancestor's caret.
  $: guides = $prefs.sidebarIndentGuides ? Array.from({ length: depth }, (_, i) => i) : []
</script>

<div class="row" class:active style={`padding-left:${indent}`}>
  {#each guides as level (level)}
    <span
      class="guide"
      aria-hidden="true"
      style={`left:calc(var(--space-3) + ${level} * var(--space-4) + 9px)`}
    ></span>
  {/each}
  {#if expandable}
    <button
      type="button"
      class="caret"
      class:open={expanded}
      aria-label={expanded ? 'Collapse' : 'Expand'}
      on:click|stopPropagation={() => dispatch('toggle')}
    >
      <IconChevronRight size={14} stroke={1.8} />
    </button>
  {:else}
    <span class="caret-spacer" aria-hidden="true"></span>
  {/if}

  <button type="button" class="main" class:unread={count > 0} on:click={() => dispatch('select')}>
    <span class="icon" aria-hidden="true"><slot /></span>
    <span class="label">{label}</span>
    {#if count > 0}
      <span class="count" aria-label={`${count} unread`}>{count}</span>
    {/if}
  </button>
</div>

<style>
  .row {
    position: relative;
    display: flex;
    align-items: center;
    gap: var(--space-1);
    padding-right: var(--space-3);
    border-radius: var(--radius-control);
  }

  /* vertical indent guide for a nested folder's ancestor level. a fixed 1px keeps
     it crisp at the fractional x-position the calc lands on (a sub-pixel hairline
     there renders nearly invisible), and border-default reads clearly in both
     themes. */
  .guide {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 1px;
    background: var(--border-default);
    pointer-events: none;
  }

  .row:not(.active):hover {
    background: var(--surface-hover);
  }

  .row.active {
    background: var(--selection-bg);
  }

  .caret,
  .caret-spacer {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 22px;
    flex-shrink: 0;
  }

  .caret {
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    border-radius: var(--radius-control);
  }

  .caret :global(svg) {
    transition: transform 0.12s ease;
  }

  .caret.open :global(svg) {
    transform: rotate(90deg);
  }

  .main {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex: 1;
    min-width: 0;
    padding: var(--row-pad-y) var(--space-2);
    border: none;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    text-align: left;
    font-size: var(--fz-list);
    line-height: 1.2;
  }

  .row.active .main {
    color: var(--text-primary);
  }

  .main.unread {
    color: var(--text-primary);
    font-weight: var(--fw-semibold);
  }

  .icon {
    display: inline-flex;
    color: var(--text-tertiary);
    flex-shrink: 0;
  }

  .row.active .icon {
    color: var(--accent);
  }

  .label {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .count {
    flex-shrink: 0;
    font-size: var(--fz-meta);
    font-weight: var(--fw-medium);
    color: var(--text-tertiary);
    background: var(--surface-sunken);
    border-radius: 999px;
    padding: 0 var(--space-2);
    min-width: 18px;
    text-align: center;
  }
</style>
