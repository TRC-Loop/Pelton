<script lang="ts">
  // shown in place of the body when a protected message could not be opened.
  // each state gets its own next step, because "locked" and "no key" need
  // completely different things from the reader: one is a passphrase they know,
  // the other is a key they have to import.
  import { createEventDispatcher } from 'svelte'
  import { IconLock, IconKeyOff, IconAlertTriangle } from '@tabler/icons-svelte'
  import { unlockMessage } from '../../lib/api'
  import { errorMessage } from '../../stores/toast'
  import { t } from '../../lib/i18n'
  import type { PGPState } from '../../lib/types'

  export let state: PGPState
  export let messageId: number

  const dispatch = createEventDispatcher<{ opened: void }>()

  let passphrase = ''
  let busy = false
  let failed = ''

  // a new message clears whatever was typed for the previous one, so a
  // passphrase never carries across to a message it was not meant for.
  let seededFor = -1
  $: if (messageId !== seededFor) {
    seededFor = messageId
    passphrase = ''
    failed = ''
    busy = false
  }

  async function unlock(): Promise<void> {
    if (busy || passphrase === '') {
      return
    }
    busy = true
    failed = ''
    try {
      await unlockMessage(messageId, passphrase)
      // clear it the moment it is no longer needed rather than leaving it in
      // component state for the lifetime of the pane.
      passphrase = ''
      dispatch('opened')
    } catch (err) {
      failed = errorMessage(err)
    } finally {
      busy = false
    }
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter') {
      void unlock()
    }
  }
</script>

<div class="notice">
  <span class="icon">
    {#if state === 'locked'}
      <IconLock size={18} stroke={1.7} />
    {:else if state === 'nokey'}
      <IconKeyOff size={18} stroke={1.7} />
    {:else}
      <IconAlertTriangle size={18} stroke={1.7} />
    {/if}
  </span>

  <div class="body">
    <p class="headline">{$t(`detail.pgp.${state}.title`)}</p>
    <p class="explain">{$t(`detail.pgp.${state}.body`)}</p>

    {#if state === 'locked'}
      <div class="unlock">
        <input
          type="password"
          bind:value={passphrase}
          on:keydown={onKeydown}
          placeholder={$t('detail.pgp.passphrase')}
          autocomplete="off"
          spellcheck="false"
          disabled={busy}
        />
        <button type="button" on:click={unlock} disabled={busy || passphrase === ''}>
          {$t('detail.pgp.unlock')}
        </button>
      </div>
      {#if failed !== ''}
        <p class="failed">{failed}</p>
      {/if}
    {/if}
  </div>
</div>

<style>
  .notice {
    display: flex;
    gap: var(--space-3);
    padding: var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-sunken);
  }

  .icon {
    display: inline-flex;
    flex-shrink: 0;
    color: var(--text-tertiary);
  }

  .body {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    min-width: 0;
  }

  .headline {
    margin: 0;
    font-size: var(--fz-body);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }

  .explain {
    margin: 0;
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }

  .unlock {
    display: flex;
    gap: var(--space-2);
    margin-top: var(--space-2);
  }

  .unlock input {
    flex: 1;
    min-width: 0;
    padding: var(--space-2) var(--space-3);
    font-family: var(--font-ui);
    font-size: var(--fz-body);
    color: var(--text-primary);
    background: var(--surface-raised);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
  }
  .unlock input:focus {
    outline: none;
    border-color: var(--accent);
  }

  .unlock button {
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--accent-fg);
    background: var(--accent);
    border: none;
    border-radius: var(--radius-control);
    cursor: var(--cursor-action);
  }
  .unlock button:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .failed {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--danger);
  }
</style>
