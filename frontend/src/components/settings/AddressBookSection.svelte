<script lang="ts">
  // the address book manager. it lists every harvested contact (from sent and
  // received mail) with how often it has been used, and lets the user remove
  // any of them. the book is capped and self-prunes; this is the manual override.
  import { onMount } from 'svelte'
  import { IconTrash, IconSearch } from '@tabler/icons-svelte'
  import { listAddresses, deleteAddress } from '../../lib/api'
  import { errorMessage, toastError } from '../../stores/toast'
  import type { AddressBookEntry } from '../../lib/types'
  import { t } from '../../lib/i18n'

  let entries: AddressBookEntry[] = []
  let query = ''
  let loading = true

  onMount(load)

  async function load(): Promise<void> {
    loading = true
    try {
      entries = await listAddresses()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      loading = false
    }
  }

  async function remove(email: string): Promise<void> {
    try {
      await deleteAddress(email)
      entries = entries.filter((e) => e.email !== email)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  $: filtered = query.trim()
    ? entries.filter(
        (e) =>
          e.email.toLowerCase().includes(query.toLowerCase()) ||
          e.name.toLowerCase().includes(query.toLowerCase()),
      )
    : entries
</script>

<h3>{$t('addressBook.title')}</h3>
<p class="hint">{$t('addressBook.hint')}</p>

<div class="search">
  <IconSearch size={15} stroke={1.7} />
  <input type="search" placeholder={$t('addressBook.searchPlaceholder')} bind:value={query} />
</div>

{#if loading}
  <p class="empty">{$t('addressBook.loading')}</p>
{:else if filtered.length === 0}
  <p class="empty">{query ? $t('addressBook.noMatches') : $t('addressBook.empty')}</p>
{:else}
  <ul class="list">
    {#each filtered as e (e.email)}
      <li>
        <div class="who">
          <span class="name">{e.name || e.email}</span>
          {#if e.name}<span class="addr">{e.email}</span>{/if}
        </div>
        <span class="uses" title={$t('addressBook.timesUsed')}>{e.useCount}×</span>
        <button type="button" class="del" aria-label={`${$t('addressBook.remove')} ${e.email}`} on:click={() => remove(e.email)}>
          <IconTrash size={15} stroke={1.7} />
        </button>
      </li>
    {/each}
  </ul>
{/if}

<style>
  h3 {
    margin: 0 0 var(--space-3);
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .hint {
    margin: 0 0 var(--space-4);
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .search {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-tertiary);
    margin-bottom: var(--space-3);
  }
  .search input {
    flex: 1;
    border: none;
    background: transparent;
    color: var(--text-primary);
    font: inherit;
    outline: none;
  }

  .empty {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    padding: var(--space-3) 0;
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    max-height: 420px;
    overflow-y: auto;
  }

  li {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-1);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .who {
    display: flex;
    flex-direction: column;
    min-width: 0;
    flex: 1;
  }
  .name {
    font-size: var(--fz-label);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .addr {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .uses {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    font-variant-numeric: tabular-nums;
    flex-shrink: 0;
  }

  .del {
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: var(--space-1);
    border-radius: var(--radius-control);
    flex-shrink: 0;
  }
  .del:hover {
    background: var(--surface-hover);
    color: var(--danger);
  }
</style>
