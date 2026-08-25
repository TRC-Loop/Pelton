<script lang="ts">
  // Importing .eml and .mbox files (#308). This used to share a dialog with the
  // Thunderbird migration, which made a job you might do any day look like part
  // of a one-off move between clients.
  //
  // Imported mail lands in Local Folders, an account sync never touches, so
  // nothing here can reach a server or change what is on one.
  import { createEventDispatcher, onDestroy } from 'svelte'
  import Modal from '../common/Modal.svelte'
  import { IconFileImport } from '@tabler/icons-svelte'
  import { chooseMailFiles, importMailFiles } from '../../lib/api'
  import { onImportProgress, type ImportProgressEvent } from '../../lib/events'
  import { loadSidebar } from '../../stores/accounts'
  import { toastError, toastSuccess, errorMessage } from '../../stores/toast'
  import { get } from 'svelte/store'
  import { t } from '../../lib/i18n'

  const dispatch = createEventDispatcher<{ close: void }>()

  let progress: ImportProgressEvent | null = null
  $: running = progress?.running ?? false
  $: percent =
    progress && progress.bytesTotal > 0
      ? Math.min(100, Math.round((progress.bytesDone / progress.bytesTotal) * 100))
      : 0

  const stop = onImportProgress((event) => {
    progress = event
    if (!event.running) {
      finish(event)
    }
  })
  onDestroy(stop)

  function finish(event: ImportProgressEvent): void {
    if (event.error) {
      toastError(event.error)
      return
    }
    toastSuccess(`${event.imported} ${get(t)('import.messagesImported')}`)
    // the Local Folders section appears the moment the first import lands.
    void loadSidebar()
  }

  async function run(): Promise<void> {
    try {
      const paths = await chooseMailFiles()
      if (paths.length === 0) {
        return
      }
      await importMailFiles(paths)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }
</script>

<Modal
  title={$t('import.files.title')}
  hint={$t('import.files.intro')}
  size="small"
  closable={!running}
  on:close={() => dispatch('close')}
>
  <span slot="icon"><IconFileImport size={16} stroke={1.8} /></span>

  <div class="body">
    <p class="hint">{$t('import.files.hint')}</p>
    <button type="button" class="action-btn" disabled={running} on:click={run}>
      {$t('import.files.choose')}
    </button>

    {#if progress}
      <div class="progress" role="status" aria-live="polite">
        <div class="progress-line">
          <span class="p-where">
            {#if running}
              {progress.fileTotal > 1
                ? $t('import.progress.fileOf')
                    .replace('{name}', progress.folder || '')
                    .replace('{i}', String(progress.fileIndex))
                    .replace('{n}', String(progress.fileTotal))
                : progress.folder || $t('import.importing')}
            {:else}
              {$t('import.finished')}
            {/if}
          </span>
          <span class="row-meta">
            {progress.imported.toLocaleString()}
            {$t('import.messagesImported')}{#if running && progress.bytesTotal > 0}&nbsp;· {percent}%{/if}
          </span>
        </div>
        {#if running}
          <div class="bar"><div class="fill" style:width="{percent}%"></div></div>
        {/if}
        {#if !running && progress.skipped > 0}
          <p class="hint">{$t('import.skipped')}: {progress.skipped}</p>
        {/if}
        {#if !running && progress.failed > 0}
          <p class="hint">{$t('import.failed')}: {progress.failed}</p>
        {/if}
      </div>
    {/if}
  </div>

  <svelte:fragment slot="footer">
    <button type="button" class="action-btn primary" disabled={running} on:click={() => dispatch('close')}>
      {$t('modal.close')}
    </button>
  </svelte:fragment>
</Modal>

<style>
  .body {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .hint {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .row-meta {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .action-btn {
    align-self: flex-start;
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
    color: var(--text-primary);
    background: var(--surface-raised);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    cursor: var(--cursor-action);
  }
  .action-btn:hover:not(:disabled) {
    background: var(--surface-hover);
  }
  .action-btn:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .action-btn.primary {
    color: var(--accent-fg);
    background: var(--accent);
    border-color: transparent;
  }

  .progress {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding-top: var(--space-3);
    border-top: var(--hairline) solid var(--border-subtle);
  }

  .progress-line {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3);
    font-size: var(--fz-label);
    color: var(--text-primary);
  }

  .p-where {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .bar {
    height: 4px;
    border-radius: 999px;
    background: var(--surface-sunken);
    overflow: hidden;
  }

  .fill {
    height: 100%;
    background: var(--accent);
    transition: width 120ms linear;
  }
</style>
