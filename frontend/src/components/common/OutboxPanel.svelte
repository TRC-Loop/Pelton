<script lang="ts">
  // the popover that opens from the status bar's outbox indicator. it lists every
  // queued, sending and failed message with its recipients and state, an
  // indeterminate progress bar while sending, and the error for failed ones. it
  // is read-only; the worker drives the actual state.
  import { IconSend, IconClock, IconAlertTriangle, IconX } from '@tabler/icons-svelte'
  import { createEventDispatcher } from 'svelte'
  import { outbox } from '../../stores/outbox'
  import type { OutboxRow } from '../../lib/types'
  import { t } from '../../lib/i18n'

  const dispatch = createEventDispatcher<{ close: void }>()

  function recipientLabel(row: OutboxRow, tFn: (key: string) => string): string {
    if (row.recipients.length === 0) {
      return tFn('common.outboxPanel.noRecipients')
    }
    if (row.recipients.length === 1) {
      return row.recipients[0]
    }
    return `${row.recipients[0]} +${row.recipients.length - 1}`
  }
</script>

<div class="panel" role="dialog" aria-label={$t('common.outboxPanel.title')}>
  <header>
    <span class="title">{$t('common.outboxPanel.title')}</span>
    <button type="button" class="close" aria-label={$t('common.outboxPanel.close')} on:click={() => dispatch('close')}>
      <IconX size={14} stroke={1.8} />
    </button>
  </header>

  {#if $outbox.length === 0}
    <p class="empty">{$t('common.outboxPanel.empty')}</p>
  {:else}
    <ul>
      {#each $outbox as row (row.id)}
        <li class="item" class:failed={row.state === 'failed'}>
          <span class="state-icon">
            {#if row.state === 'sending'}
              <IconSend size={15} stroke={1.7} />
            {:else if row.state === 'failed'}
              <IconAlertTriangle size={15} stroke={1.7} />
            {:else}
              <IconClock size={15} stroke={1.7} />
            {/if}
          </span>
          <span class="body">
            <span class="to">{recipientLabel(row, $t)}</span>
            {#if row.state === 'sending'}
              <span class="bar"><span class="fill"></span></span>
            {:else if row.state === 'failed'}
              <span class="err">{row.lastError || $t('common.outboxPanel.sendFailed')}</span>
            {:else}
              <span class="muted">{$t('common.outboxPanel.queued')}</span>
            {/if}
          </span>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .panel {
    width: 320px;
    max-height: 360px;
    display: flex;
    flex-direction: column;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
    overflow: hidden;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-3) var(--space-4);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .title {
    font-size: var(--fz-label);
    font-weight: var(--fw-semibold);
    color: var(--text-secondary);
  }

  .close {
    display: inline-flex;
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

  .empty {
    margin: 0;
    padding: var(--space-5);
    text-align: center;
    color: var(--text-tertiary);
    font-size: var(--fz-label);
  }

  ul {
    list-style: none;
    margin: 0;
    padding: var(--space-2);
    overflow-y: auto;
  }

  .item {
    display: flex;
    gap: var(--space-3);
    padding: var(--space-3);
    border-radius: var(--radius-control);
  }

  .item + .item {
    border-top: var(--hairline) solid var(--border-subtle);
  }

  .state-icon {
    color: var(--text-tertiary);
    flex-shrink: 0;
    margin-top: 1px;
  }

  .item.failed .state-icon {
    color: var(--danger);
  }

  .body {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    min-width: 0;
    flex: 1;
  }

  .to {
    font-size: var(--fz-label);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .muted {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .err {
    font-size: var(--fz-meta);
    color: var(--danger);
    word-break: break-word;
  }

  /* indeterminate progress while sending. */
  .bar {
    position: relative;
    height: 3px;
    border-radius: 999px;
    background: var(--surface-sunken);
    overflow: hidden;
  }

  .fill {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 40%;
    border-radius: 999px;
    background: var(--accent);
    animation: slide 1.1s ease-in-out infinite;
  }

  @keyframes slide {
    0% {
      left: -40%;
    }
    100% {
      left: 100%;
    }
  }
</style>
