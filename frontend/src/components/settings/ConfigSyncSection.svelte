<script lang="ts">
  // settings sync: mirrors settings (and optionally local message metadata or
  // the whole offline cache) through a folder the user points at, such as a
  // Nextcloud or Dropbox desktop-sync directory. Pelton never talks to a cloud
  // provider itself; it only reads and writes plain files there and lets
  // whatever already syncs that folder do the transport.
  import { onDestroy, onMount } from 'svelte'
  import { IconRefresh } from '@tabler/icons-svelte'
  import {
    getConfigSyncStatus,
    configureConfigSync,
    disableConfigSync,
    triggerConfigSync,
    pickConfigSyncFolder,
    type ConfigSyncStatus,
  } from '../../lib/api'
  import { onConfigSyncStatus } from '../../lib/events'
  import { toastError, toastSuccess, errorMessage } from '../../stores/toast'
  import { formatFullDate } from '../../lib/format'
  import { get } from 'svelte/store'
  import { t } from '../../lib/i18n'

  let status: ConfigSyncStatus | null = null
  let loading = true
  let setupOpen = false
  let syncing = false

  onMount(async () => {
    try {
      status = await getConfigSyncStatus()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      loading = false
    }
  })

  const off = onConfigSyncStatus((e) => {
    status = e
  })
  onDestroy(off)

  async function disable(): Promise<void> {
    try {
      status = await disableConfigSync()
      toastSuccess(get(t)('configSync.disabledToast'))
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function syncNow(): Promise<void> {
    syncing = true
    try {
      status = await triggerConfigSync()
      if (status.lastError) {
        toastError(status.lastError)
      } else {
        toastSuccess(get(t)('configSync.syncedToast'))
      }
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      syncing = false
    }
  }

  function onConfigured(next: ConfigSyncStatus): void {
    status = next
    setupOpen = false
  }
</script>

<div class="section">
  {#if loading}
    <p class="hint">{$t('configSync.loading')}</p>
  {:else if !status?.enabled}
    <p class="hint">{$t('configSync.introHint')}</p>
    <button type="button" class="primary" on:click={() => (setupOpen = true)}>{$t('configSync.setup')}</button>
  {:else}
    <div class="status">
      <div class="row">
        <span class="row-label">{$t('configSync.folder')}</span>
        <span class="row-value mono">{status.path}</span>
      </div>
      <div class="row">
        <span class="row-label">{$t('configSync.mode')}</span>
        <span class="row-value">{status.mode === 'copy' ? $t('configSync.modeCopyFull') : $t('configSync.modeReadonlyFull')}</span>
      </div>
      <div class="row">
        <span class="row-label">{$t('configSync.scope')}</span>
        <span class="row-value">
          {[
            status.syncSettings ? $t('configSync.scopeSettings') : null,
            status.emailScope === 'metadata'
              ? $t('configSync.scopeMetadata')
              : status.emailScope === 'full'
                ? $t('configSync.scopeFullCache')
                : null,
          ]
            .filter(Boolean)
            .join(', ') || $t('configSync.scopeNothing')}
        </span>
      </div>
      <div class="row">
        <span class="row-label">{$t('configSync.lastSynced')}</span>
        <span class="row-value">
          {status.lastSyncUnix ? formatFullDate(new Date(status.lastSyncUnix * 1000).toISOString()) : $t('configSync.neverYet')}
        </span>
      </div>
      {#if status.lastError}
        <div class="row">
          <span class="row-label">{$t('configSync.lastError')}</span>
          <span class="row-value error">{status.lastError}</span>
        </div>
      {/if}
    </div>

    <div class="actions">
      <button type="button" class="action-btn" on:click={syncNow} disabled={syncing}>
        <IconRefresh size={14} stroke={1.8} class={syncing ? 'spin' : ''} />
        {$t('configSync.syncNow')}
      </button>
      <button type="button" class="action-btn" on:click={() => (setupOpen = true)}>{$t('configSync.changeSetup')}</button>
      <button type="button" class="action-btn danger" on:click={disable}>{$t('configSync.disable')}</button>
    </div>
  {/if}
</div>

{#if setupOpen}
  {#await import('./ConfigSyncSetupModal.svelte') then m}
    <svelte:component
      this={m.default}
      current={status}
      on:close={() => (setupOpen = false)}
      on:configured={(e) => onConfigured(e.detail)}
    />
  {/await}
{/if}

<style>
  .section {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .hint {
    margin: 0;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .primary {
    align-self: flex-start;
    padding: var(--space-3) var(--space-5);
    border: none;
    border-radius: var(--radius-control);
    background: var(--accent);
    color: var(--accent-fg);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    cursor: pointer;
  }

  .primary:hover {
    filter: brightness(1.05);
  }

  .status {
    display: flex;
    flex-direction: column;
  }

  .row {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-3) 0;
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .row-label {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }

  .row-value {
    font-size: var(--fz-label);
    color: var(--text-primary);
    text-align: right;
  }

  .row-value.mono {
    font-family: var(--font-mono);
    word-break: break-all;
  }

  .row-value.error {
    color: var(--danger);
  }

  .actions {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
  }

  .action-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .action-btn:hover {
    background: var(--surface-hover);
  }

  .action-btn:disabled {
    opacity: 0.6;
    cursor: default;
  }

  .action-btn.danger {
    color: var(--danger);
    border-color: var(--danger);
  }

  .action-btn.danger:hover {
    background: color-mix(in srgb, var(--danger) 12%, transparent);
  }

  :global(.spin) {
    /* svg transform-origin defaults differ across the webviews wails embeds
       per platform (webkit vs webview2); pin it explicitly so the icon spins
       in place instead of wobbling around an off-center pivot on some os. */
    transform-box: border-box;
    transform-origin: 50% 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
