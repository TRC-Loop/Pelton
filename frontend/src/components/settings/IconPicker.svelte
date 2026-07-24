<script lang="ts">
  // IconPicker lets a custom menu entry choose an icon. Icons from the active
  // theme are offered first (they restyle live with the theme), then the full
  // Tabler set, searchable by name. The Tabler geometry is a ~2MB dataset, so it
  // is lazy-imported only when the picker mounts, keeping it out of the main
  // bundle; the chosen tabler icon's node geometry is returned so the always-on
  // bar can render it without that dataset.
  import { createEventDispatcher, onMount } from 'svelte'
  import { IconSearch, IconX } from '@tabler/icons-svelte'
  import MenuGlyph from '../common/MenuGlyph.svelte'
  import { iconOverrides } from '../../theme/icons'
  import { t } from '../../lib/i18n'
  import type { IconNode } from '../../lib/menuactions'

  const dispatch = createEventDispatcher<{
    select: { iconName: string; iconNodes?: IconNode[] }
    clear: void
  }>()

  // how many tabler matches to render at once; more than this and the user is
  // asked to narrow the search, so the grid never mounts thousands of svgs.
  const maxResults = 120

  let nodes: Record<string, IconNode[]> | null = null
  let query = ''

  onMount(async () => {
    nodes = (await import('tabler-nodes-outline')).default
  })

  // normalize strips separators so "mail open" matches "mail-opened".
  function normalize(value: string): string {
    return value.toLowerCase().replace(/[\s-]/g, '')
  }

  $: q = normalize(query)
  $: themeNames = Object.keys($iconOverrides)
    .filter((name) => q === '' || normalize(name).includes(q))
    .sort()
  $: allTabler = nodes ? Object.keys(nodes) : []
  $: tablerMatches = q === '' ? allTabler : allTabler.filter((name) => normalize(name).includes(q))
  $: tablerShown = tablerMatches.slice(0, maxResults)

  function pickTheme(name: string): void {
    dispatch('select', { iconName: name })
  }

  function pickTabler(name: string): void {
    dispatch('select', { iconName: name, iconNodes: nodes?.[name] })
  }
</script>

<div class="picker">
  <div class="search">
    <IconSearch size={15} stroke={1.7} />
    <input
      type="text"
      bind:value={query}
      placeholder={$t('menuBar.icon.search')}
      aria-label={$t('menuBar.icon.search')}
    />
    <button type="button" class="clear" on:click={() => dispatch('clear')}>
      <IconX size={14} stroke={1.7} />
      <span>{$t('menuBar.icon.none')}</span>
    </button>
  </div>

  <div class="scroll">
    {#if themeNames.length > 0}
      <div class="group-label">{$t('menuBar.icon.themeGroup')}</div>
      <div class="grid">
        {#each themeNames as name (name)}
          <button type="button" class="cell" title={name} on:click={() => pickTheme(name)}>
            <MenuGlyph iconName={name} size={20} />
          </button>
        {/each}
      </div>
    {/if}

    <div class="group-label">{$t('menuBar.icon.tablerGroup')}</div>
    {#if nodes === null}
      <div class="hint">{$t('menuBar.icon.loading')}</div>
    {:else}
      <div class="grid">
        {#each tablerShown as name (name)}
          <button type="button" class="cell" title={name} on:click={() => pickTabler(name)}>
            <MenuGlyph iconNodes={nodes[name]} size={20} />
          </button>
        {/each}
      </div>
      {#if tablerMatches.length === 0}
        <div class="hint">{$t('menuBar.icon.noMatches')}</div>
      {:else if tablerMatches.length > tablerShown.length}
        <div class="hint">{$t('menuBar.icon.narrow')}</div>
      {/if}
    {/if}
  </div>
</div>

<style>
  .picker {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .search {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0 var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-tertiary);
  }

  .search input {
    flex: 1;
    padding: var(--space-2) 0;
    border: none;
    background: transparent;
    color: var(--text-primary);
    font-size: var(--fz-label);
    outline: none;
  }

  .clear {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1) var(--space-2);
    border: none;
    background: transparent;
    color: var(--text-secondary);
    font-size: var(--fz-meta);
    border-radius: var(--radius-control);
    cursor: pointer;
  }

  .clear:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .scroll {
    max-height: 240px;
    overflow-y: auto;
  }

  .group-label {
    padding: var(--space-2) var(--space-1) var(--space-1);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(34px, 1fr));
    gap: var(--space-1);
  }

  .cell {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    aspect-ratio: 1;
    border: var(--hairline) solid transparent;
    background: transparent;
    color: var(--text-secondary);
    border-radius: var(--radius-control);
    cursor: pointer;
  }

  .cell:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
    border-color: var(--border-subtle);
  }

  .hint {
    padding: var(--space-2) var(--space-1);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }
</style>
