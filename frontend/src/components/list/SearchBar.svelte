<script lang="ts">
  // the search field at the top of the message list. it debounces input and emits
  // the query; an empty query (and no date filter) means "show the normal list".
  // a date-filter dropdown narrows results to a preset window or a custom range,
  // emitted alongside the query so the list re-runs the ranked search.
  import { createEventDispatcher } from 'svelte'
  import { IconSearch, IconX, IconCalendar, IconCheck } from '@tabler/icons-svelte'
  import { prefs } from '../../stores/prefs'
  import { shortcutLabel } from '../../lib/i18n'
  import { emptyFilter, type SearchFilter } from '../../stores/messages'

  export let value: string = ''

  // the search shortcut, shown as a hint chip when the user enabled hints and the
  // field is empty (so it never overlaps the clear button or typed text).
  const searchHint = shortcutLabel('mod+f')

  const dispatch = createEventDispatcher<{ search: string; filter: SearchFilter }>()
  let timer: ReturnType<typeof setTimeout> | undefined

  // the active date filter and which preset produced it (for the dropdown ui).
  let filter: SearchFilter = emptyFilter
  let activePreset = 'any'
  let showFilter = false
  // custom range bound to the two date inputs (yyyy-mm-dd strings).
  let customFrom = ''
  let customTo = ''

  $: filterActive = filter.afterUnix > 0 || filter.beforeUnix > 0

  const day = 86400

  // presets are computed relative to now so "last 7 days" always means the most
  // recent week.
  const presets: { key: string; label: string; days: number }[] = [
    { key: 'any', label: 'Any time', days: 0 },
    { key: '7', label: 'Last 7 days', days: 7 },
    { key: '30', label: 'Last 30 days', days: 30 },
    { key: '365', label: 'Last 12 months', days: 365 },
  ]

  function onInput(event: Event): void {
    const next = (event.currentTarget as HTMLInputElement).value
    clearTimeout(timer)
    timer = setTimeout(() => dispatch('search', next), 220)
  }

  function clear(): void {
    clearTimeout(timer)
    dispatch('search', '')
  }

  function applyFilter(next: SearchFilter, preset: string): void {
    filter = next
    activePreset = preset
    dispatch('filter', next)
  }

  function pickPreset(key: string, days: number): void {
    if (days === 0) {
      applyFilter(emptyFilter, 'any')
    } else {
      const now = Math.floor(Date.now() / 1000)
      applyFilter({ afterUnix: now - days * day, beforeUnix: 0 }, key)
    }
    showFilter = false
  }

  function applyCustom(): void {
    const after = customFrom ? Math.floor(Date.parse(customFrom) / 1000) : 0
    // include the whole "to" day by adding a day minus a second.
    const to = customTo ? Math.floor(Date.parse(customTo) / 1000) + day - 1 : 0
    if (!after && !to) {
      return
    }
    applyFilter({ afterUnix: after || 0, beforeUnix: to || 0 }, 'custom')
    showFilter = false
  }

  function clearFilter(): void {
    customFrom = ''
    customTo = ''
    applyFilter(emptyFilter, 'any')
    showFilter = false
  }
</script>

<div class="bar">
  <div class="search">
    <IconSearch size={15} stroke={1.6} class="search-icon" />
    <input
      type="search"
      placeholder="Search mail"
      aria-label="Search mail"
      {value}
      on:input={onInput}
    />
    {#if value}
      <button type="button" class="clear" aria-label="Clear search" on:click={clear}>
        <IconX size={14} stroke={1.8} />
      </button>
    {:else if $prefs.showShortcutHints}
      <kbd class="hint">{searchHint}</kbd>
    {/if}
  </div>

  <div class="filter-wrap">
    <button
      type="button"
      class="filter-btn"
      class:active={filterActive}
      aria-label="Filter by date"
      aria-expanded={showFilter}
      title="Filter by date"
      on:click={() => (showFilter = !showFilter)}
    >
      <IconCalendar size={16} stroke={1.7} />
    </button>

    {#if showFilter}
      <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
      <div class="scrim" on:click={() => (showFilter = false)}></div>
      <div class="menu" role="menu">
        {#each presets as p (p.key)}
          <button type="button" class="opt" role="menuitemradio" aria-checked={activePreset === p.key} on:click={() => pickPreset(p.key, p.days)}>
            <span class="tick">{#if activePreset === p.key}<IconCheck size={13} stroke={2.2} />{/if}</span>
            {p.label}
          </button>
        {/each}

        <div class="custom">
          <span class="custom-label">Custom range</span>
          <label class="date">
            <span>From</span>
            <input type="date" bind:value={customFrom} />
          </label>
          <label class="date">
            <span>To</span>
            <input type="date" bind:value={customTo} />
          </label>
          <div class="custom-actions">
            <button type="button" class="ghost" on:click={clearFilter}>Clear</button>
            <button type="button" class="primary" on:click={applyCustom}>Apply</button>
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .bar {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .search {
    flex: 1;
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0 var(--space-3);
    height: var(--control-height);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-tertiary);
  }

  .search:focus-within {
    border-color: var(--accent);
  }

  input[type='search'] {
    flex: 1;
    min-width: 0;
    border: none;
    background: transparent;
    outline: none;
    font-size: var(--fz-list);
  }

  /* hide the native search clear so we control the affordance. */
  input::-webkit-search-cancel-button {
    display: none;
  }

  .clear {
    display: inline-flex;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: 2px;
    border-radius: var(--radius-control);
  }

  .clear:hover {
    color: var(--text-primary);
  }

  .hint {
    flex-shrink: 0;
    padding: 1px var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-tertiary);
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    line-height: 1.4;
  }

  .filter-wrap {
    position: relative;
    flex-shrink: 0;
  }

  .filter-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: var(--control-height);
    width: var(--control-height);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-secondary);
    cursor: pointer;
  }

  .filter-btn:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  /* an active filter is shown by the accent so it is obvious results are scoped. */
  .filter-btn.active {
    border-color: var(--accent);
    color: var(--accent);
  }

  .scrim {
    position: fixed;
    inset: 0;
    z-index: 40;
  }

  .menu {
    position: absolute;
    top: calc(100% + var(--space-1));
    right: 0;
    z-index: 41;
    width: 232px;
    padding: var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
  }

  .opt {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    border: none;
    background: transparent;
    color: var(--text-primary);
    cursor: pointer;
    text-align: left;
    padding: var(--space-2);
    border-radius: var(--radius-control);
    font-size: var(--fz-label);
  }

  .opt:hover {
    background: var(--surface-hover);
  }

  .tick {
    display: inline-flex;
    width: 14px;
    color: var(--accent);
  }

  .custom {
    margin-top: var(--space-2);
    padding-top: var(--space-2);
    border-top: var(--hairline) solid var(--border-subtle);
  }

  .custom-label {
    display: block;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    margin: 0 var(--space-2) var(--space-2);
  }

  .date {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    padding: 0 var(--space-2) var(--space-2);
  }

  .date span {
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }

  .date input {
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    padding: var(--space-1) var(--space-2);
    font-size: var(--fz-label);
    color: var(--text-primary);
  }

  .custom-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    padding: var(--space-1) var(--space-2) 0;
  }

  .ghost,
  .primary {
    padding: var(--space-1) var(--space-3);
    border-radius: var(--radius-control);
    border: var(--hairline) solid var(--border-default);
    cursor: pointer;
    font-size: var(--fz-label);
  }

  .ghost {
    background: var(--surface-raised);
    color: var(--text-primary);
  }

  .primary {
    background: var(--accent);
    color: var(--accent-fg);
    border-color: transparent;
  }
</style>
