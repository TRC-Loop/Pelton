<script lang="ts">
  // what happened to a mailbox that failed to sync, and a button to try again
  // (#322). Opened from the mark next to the account in the sidebar and from
  // the status bar, which is why it lives here rather than inside either of
  // them: one dialog, two ways in.
  //
  // It answers the three questions the old behaviour left open: which mailbox,
  // what went wrong, and how long it has been that way.
  import { IconCloudOff, IconRefresh } from '@tabler/icons-svelte'
  import Modal from './Modal.svelte'
  import { failureDetail, closeSyncFailure, retrySync } from '../../stores/syncfailures'
  import { formatRelative } from '../../lib/format'
  import { t } from '../../lib/i18n'

  let busy = false

  // the reason is a coarse class from the backend; anything unrecognized reads
  // as the generic sentence rather than as a missing string.
  const reasons = ['auth', 'network', 'credentials', 'other']
  $: reasonKey = $failureDetail && reasons.includes($failureDetail.reason) ? $failureDetail.reason : 'other'

  // how long it has been broken. Never having synced at all is its own line:
  // "last synced never" reads like a bug.
  $: lastOk = $failureDetail && $failureDetail.lastOk !== '' ? Date.parse($failureDetail.lastOk) : NaN
  $: lastOkLabel = Number.isNaN(lastOk)
    ? $t('syncFailure.neverSynced')
    : $t('syncFailure.lastSynced').replace('{when}', formatRelative(lastOk, $t))

  async function retry(): Promise<void> {
    if (!$failureDetail || busy) {
      return
    }
    busy = true
    const ok = await retrySync($failureDetail.accountId)
    busy = false
    if (ok) {
      closeSyncFailure()
    }
    // on a repeat failure the backend has already pushed the new state, so the
    // dialog is showing this attempt's reason by the time this returns.
  }
</script>

{#if $failureDetail}
  <Modal title={$t('syncFailure.title')} size="small" on:close={closeSyncFailure}>
    <span slot="icon"><IconCloudOff size={16} stroke={1.8} /></span>

    <p class="who">{$failureDetail.email}</p>
    <p class="reason">{$t(`syncFailure.reason.${reasonKey}`)}</p>
    <p class="when">{lastOkLabel}</p>
    {#if $failureDetail.detail !== ''}
      <p class="detail-label">{$t('syncFailure.detailLabel')}</p>
      <pre class="detail">{$failureDetail.detail}</pre>
    {/if}

    <svelte:fragment slot="footer">
      <button type="button" class="cancel" on:click={closeSyncFailure}>
        {$t('syncFailure.close')}
      </button>
      <button type="button" class="go" disabled={busy} on:click={retry}>
        <IconRefresh size={14} stroke={1.8} />
        {busy ? $t('syncFailure.retrying') : $t('syncFailure.retry')}
      </button>
    </svelte:fragment>
  </Modal>
{/if}

<style>
  .who {
    margin: 0;
    font-size: var(--fz-body);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }

  .reason {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--danger);
  }

  .when {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .detail-label {
    margin: var(--space-2) 0 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .detail {
    margin: 0;
    padding: var(--space-2) var(--space-3);
    max-height: 120px;
    overflow: auto;
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
    background: var(--surface-raised);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    white-space: pre-wrap;
    word-break: break-word;
  }

  .cancel {
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
    color: var(--text-secondary);
    background: transparent;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    cursor: var(--cursor-action);
  }
  .cancel:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .go {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
    color: var(--accent-fg);
    background: var(--accent);
    border: none;
    border-radius: var(--radius-control);
    cursor: var(--cursor-action);
  }
  .go:disabled {
    opacity: 0.6;
    cursor: default;
  }
</style>
