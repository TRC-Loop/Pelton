<script lang="ts">
  // editable keyboard shortcuts. each app action shows its current combo and can
  // be rebound by clicking Change and pressing the new keys; a search box filters
  // the list, and each row (and the whole set) can be reset to the default.
  // capturing is exact: the recorded combo includes every held modifier.
  import { IconSearch, IconX, IconRotateClockwise } from '@tabler/icons-svelte'
  import { shortcuts as registry, eventToCombo, type ShortcutAction } from '../../lib/shortcuts'
  import { bindings, recording, setBinding, resetBinding, resetAll, conflictsFor } from '../../stores/shortcuts'
  import { t, shortcutLabel } from '../../lib/i18n'

  let query = ''

  // the action currently being recorded, plus any conflict message to show.
  let recordingAction: ShortcutAction | null = null
  let conflictMsg = ''

  // filter rows by their localized label (and the action key as a fallback).
  $: rows = registry.filter((s) => {
    const q = query.trim().toLowerCase()
    if (!q) {
      return true
    }
    return t(s.labelKey).toLowerCase().includes(q) || s.action.includes(q)
  })

  function startRecord(action: ShortcutAction): void {
    recordingAction = action
    conflictMsg = ''
    recording.set(true)
  }

  function stopRecord(): void {
    recordingAction = null
    conflictMsg = ''
    recording.set(false)
  }

  // onKey captures the new combo. Escape cancels; a bare modifier is ignored so
  // the user can build a combo; a duplicate is rejected with an inline message.
  function onKey(event: KeyboardEvent): void {
    if (recordingAction === null) {
      return
    }
    event.preventDefault()
    event.stopPropagation()
    if (event.key === 'Escape') {
      stopRecord()
      return
    }
    const combo = eventToCombo(event)
    if (!combo) {
      return
    }
    const clash = conflictsFor(recordingAction, combo)
    if (clash) {
      conflictMsg = `Already used by “${t(registryLabel(clash))}”`
      return
    }
    setBinding(recordingAction, combo)
    stopRecord()
  }

  function registryLabel(action: ShortcutAction): string {
    return registry.find((s) => s.action === action)?.labelKey ?? action
  }
</script>

<svelte:window on:keydown={onKey} />

<div class="head-row">
  <div class="search">
    <IconSearch size={15} stroke={1.6} />
    <input type="search" placeholder="Search shortcuts" aria-label="Search shortcuts" bind:value={query} />
    {#if query}
      <button type="button" class="clear" aria-label="Clear" on:click={() => (query = '')}>
        <IconX size={14} stroke={1.8} />
      </button>
    {/if}
  </div>
  <button type="button" class="reset-all" on:click={resetAll} title="Reset all shortcuts to defaults">
    <IconRotateClockwise size={14} stroke={1.7} /> Reset all
  </button>
</div>

<ul class="list">
  {#each rows as sc (sc.action)}
    <li>
      <span class="label">{t(sc.labelKey)}</span>
      <div class="controls">
        {#if recordingAction === sc.action}
          <span class="recording">Press keys… <kbd>Esc</kbd> to cancel</span>
        {:else}
          <kbd>{shortcutLabel($bindings[sc.action])}</kbd>
        {/if}
        <button type="button" class="change" on:click={() => startRecord(sc.action)}>
          {recordingAction === sc.action ? 'Recording' : 'Change'}
        </button>
        <button
          type="button"
          class="reset"
          aria-label="Reset to default"
          title="Reset to default"
          on:click={() => resetBinding(sc.action)}
        >
          <IconRotateClockwise size={14} stroke={1.7} />
        </button>
      </div>
    </li>
  {/each}
  {#if rows.length === 0}
    <li class="empty">No shortcuts match “{query}”.</li>
  {/if}
</ul>

{#if conflictMsg}
  <p class="conflict">{conflictMsg}. Pick a different combination.</p>
{/if}

<style>
  .head-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
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

  .search input {
    flex: 1;
    min-width: 0;
    border: none;
    background: transparent;
    outline: none;
    font-size: var(--fz-label);
    color: var(--text-primary);
  }

  .search input::-webkit-search-cancel-button {
    display: none;
  }

  .clear {
    display: inline-flex;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: 2px;
  }

  .reset-all {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    flex-shrink: 0;
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-secondary);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .reset-all:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-2) 0;
  }

  .empty {
    color: var(--text-tertiary);
    font-size: var(--fz-label);
    justify-content: flex-start;
  }

  .label {
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }

  .controls {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
  }

  kbd {
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    color: var(--text-primary);
    background: var(--surface-sunken);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    padding: 2px var(--space-2);
  }

  .recording {
    font-size: var(--fz-meta);
    color: var(--accent);
  }

  .recording kbd {
    color: var(--text-secondary);
  }

  .change,
  .reset {
    display: inline-flex;
    align-items: center;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-secondary);
    cursor: pointer;
    font-size: var(--fz-meta);
  }

  .change {
    padding: 2px var(--space-3);
  }

  .reset {
    padding: 2px;
  }

  .change:hover,
  .reset:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .conflict {
    margin: var(--space-2) 0 0;
    font-size: var(--fz-label);
    color: var(--warning);
  }
</style>
