<script lang="ts">
  // "Manage VIP senders" modal: lists every address the user has marked as VIP,
  // lets them add one by typing it and revoke any entry. New mail from a VIP
  // raises a native notification even when general new-mail notifications are
  // off. Reached from the Notifications settings.
  import { createEventDispatcher } from 'svelte'
  import Modal from '../common/Modal.svelte'
  import { IconTrash, IconStarFilled, IconPlus } from '@tabler/icons-svelte'
  import { vipSenders, addVIP, removeVIP, bareAddress } from '../../stores/vip'
  import { errorMessage, toastError } from '../../stores/toast'
  import { t } from '../../lib/i18n'

  const dispatch = createEventDispatcher<{ close: void }>()

  let input = ''
  let adding = false

  // sorted list of the current VIP addresses for stable rendering.
  $: entries = [...$vipSenders].sort()

  // valid guards the add button: a bare address must have an @ with text either
  // side. Full RFC validation is not the point; the backend normalizes anyway.
  $: candidate = bareAddress(input)
  $: valid = /^[^@\s]+@[^@\s]+$/.test(candidate) && !$vipSenders.has(candidate)

  async function add(): Promise<void> {
    if (!valid || adding) {
      return
    }
    adding = true
    try {
      await addVIP(input)
      input = ''
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      adding = false
    }
  }

  async function remove(address: string): Promise<void> {
    try {
      await removeVIP(address)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }
</script>

<Modal title={$t('vip.title')} hint={$t('vip.hint')} size="small" on:close={() => dispatch('close')}>
  <div class="stack">
    <form class="add" on:submit|preventDefault={add}>
      <input
        type="email"
        bind:value={input}
        placeholder={$t('vip.addPlaceholder')}
        autocomplete="off"
        spellcheck="false"
      />
      <button type="submit" class="add-btn" disabled={!valid || adding} title={$t('vip.add')}>
        <IconPlus size={15} stroke={1.8} />
        <span>{$t('vip.add')}</span>
      </button>
    </form>

    {#if entries.length === 0}
      <p class="empty">{$t('vip.empty')}</p>
    {:else}
      <ul class="list">
        {#each entries as value (value)}
          <li>
            <span class="star"><IconStarFilled size={15} /></span>
            <span class="value">{value}</span>
            <button
              type="button"
              class="remove"
              aria-label={$t('vip.remove')}
              title={$t('vip.remove')}
              on:click={() => remove(value)}
            >
              <IconTrash size={15} stroke={1.7} />
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</Modal>

<style>
  .stack {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .add {
    display: flex;
    gap: var(--space-2);
  }
  .add input {
    flex: 1;
    min-width: 0;
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-input, var(--surface-raised));
    color: var(--text-primary);
    font-size: var(--fz-label);
  }
  .add-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: var(--cursor-action);
  }
  .add-btn:hover:not(:disabled) {
    background: var(--surface-hover);
  }
  .add-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .empty {
    margin: var(--space-4) 0;
    text-align: center;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
  }

  li {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) 0;
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .star {
    flex-shrink: 0;
    color: var(--warning);
    display: inline-flex;
  }

  .value {
    flex: 1;
    min-width: 0;
    font-size: var(--fz-label);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .remove {
    flex-shrink: 0;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: var(--cursor-action);
    padding: var(--space-2);
    border-radius: var(--radius-control);
  }
  .remove:hover {
    background: var(--danger-bg, var(--surface-hover));
    color: var(--danger);
  }
</style>
