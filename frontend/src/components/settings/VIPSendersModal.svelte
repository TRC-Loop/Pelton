<script lang="ts">
  // "Manage VIP senders" modal: lists every address the user has marked as VIP,
  // lets them add one by typing it and revoke any entry. New mail from a VIP
  // raises a native notification even when general new-mail notifications are
  // off. Reached from the Notifications settings.
  import { createEventDispatcher } from 'svelte'
  import { fade, scale } from 'svelte/transition'
  import { IconX, IconTrash, IconStarFilled, IconPlus } from '@tabler/icons-svelte'
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

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      dispatch('close')
    }
  }
</script>

<svelte:window on:keydown={onKeydown} />

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="backdrop" transition:fade={{ duration: 120 }} on:click={() => dispatch('close')}></div>
<div
  class="dialog"
  role="dialog"
  aria-modal="true"
  aria-label={$t('vip.title')}
  transition:scale={{ duration: 150, start: 0.94 }}
>
  <header>
    <h2>{$t('vip.title')}</h2>
    <button type="button" class="close" aria-label={$t('detail.attachments.close')} on:click={() => dispatch('close')}>
      <IconX size={16} stroke={1.8} />
    </button>
  </header>

  <p class="hint">{$t('vip.hint')}</p>

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

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 300;
    background: var(--scrim, rgba(0, 0, 0, 0.4));
    backdrop-filter: blur(2px);
  }

  .dialog {
    position: fixed;
    z-index: 301;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: min(440px, calc(100vw - 2 * var(--space-5)));
    max-height: 72vh;
    padding: var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  h2 {
    margin: 0;
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }
  .close {
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: var(--space-1);
    border-radius: var(--radius-control);
  }
  .close:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .hint {
    margin: 0;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
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
    cursor: pointer;
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
    cursor: pointer;
    padding: var(--space-2);
    border-radius: var(--radius-control);
  }
  .remove:hover {
    background: var(--danger-bg, var(--surface-hover));
    color: var(--danger);
  }
</style>
