<script lang="ts">
  // Import/export replaces the removed folder-based config sync with a plain,
  // local file: the user picks what to export, and on import picks a file and
  // chooses what to bring in. Pelton only reads and writes the file the user
  // chooses; nothing leaves the machine on its own.
  import { IconDownload, IconUpload, IconFileImport } from '@tabler/icons-svelte'
  import { exportData, inspectBackupFile, importData, type BackupInfo } from '../../lib/api'
  import { initPrefs } from '../../stores/prefs'
  import { toastError, toastSuccess, errorMessage } from '../../stores/toast'
  import { formatFullDate } from '../../lib/format'
  import { get } from 'svelte/store'
  import { t } from '../../lib/i18n'

  // export selection.
  let exportSettings = true
  let exportWhitelist = true
  let exporting = false

  // import state: the inspected file and which of its categories to apply.
  let file: BackupInfo | null = null
  let importSettings = true
  let importWhitelist = true
  let importing = false

  $: exportCategories = [
    ...(exportSettings ? ['settings'] : []),
    ...(exportWhitelist ? ['whitelist'] : []),
  ]

  async function runExport(): Promise<void> {
    if (exportCategories.length === 0) {
      toastError(get(t)('importExport.nothingSelected'))
      return
    }
    exporting = true
    try {
      const path = await exportData(exportCategories)
      if (path) {
        toastSuccess(get(t)('importExport.exported'))
      }
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      exporting = false
    }
  }

  async function chooseFile(): Promise<void> {
    try {
      const info = await inspectBackupFile()
      if (!info.path) {
        return
      }
      file = info
      importSettings = info.hasSettings
      importWhitelist = info.hasWhitelist
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function runImport(): Promise<void> {
    if (!file) {
      return
    }
    const categories = [
      ...(importSettings && file.hasSettings ? ['settings'] : []),
      ...(importWhitelist && file.hasWhitelist ? ['whitelist'] : []),
    ]
    if (categories.length === 0) {
      toastError(get(t)('importExport.nothingSelected'))
      return
    }
    importing = true
    try {
      await importData(file.path, categories)
      // settings we just wrote drive the ui; reload them so it updates live.
      await initPrefs()
      toastSuccess(get(t)('importExport.imported'))
      file = null
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      importing = false
    }
  }
</script>

<div class="section">
  <div class="block">
    <div class="block-head">
      <IconUpload size={16} stroke={1.7} />
      <h4>{$t('importExport.exportTitle')}</h4>
    </div>
    <p class="hint">{$t('importExport.exportHint')}</p>
    <label class="check">
      <input type="checkbox" bind:checked={exportSettings} />
      <span>{$t('importExport.category.settings')}</span>
    </label>
    <label class="check">
      <input type="checkbox" bind:checked={exportWhitelist} />
      <span>{$t('importExport.category.whitelist')}</span>
    </label>
    <button type="button" class="action-btn primary" disabled={exporting} on:click={runExport}>
      <IconDownload size={14} stroke={1.8} />
      {$t('importExport.exportButton')}
    </button>
  </div>

  <div class="block">
    <div class="block-head">
      <IconFileImport size={16} stroke={1.7} />
      <h4>{$t('importExport.importTitle')}</h4>
    </div>
    <p class="hint">{$t('importExport.importHint')}</p>
    <button type="button" class="action-btn" on:click={chooseFile}>{$t('importExport.chooseFile')}</button>

    {#if file}
      <div class="file-info">
        <div class="meta-row">
          <span class="meta-label">{$t('importExport.created')}</span>
          <span class="meta-value">{file.createdAt ? formatFullDate(file.createdAt) : '—'}</span>
        </div>
        {#if file.appVersion}
          <div class="meta-row">
            <span class="meta-label">{$t('importExport.appVersion')}</span>
            <span class="meta-value">{file.appVersion}</span>
          </div>
        {/if}
      </div>

      <label class="check" class:disabled={!file.hasSettings}>
        <input type="checkbox" bind:checked={importSettings} disabled={!file.hasSettings} />
        <span>{$t('importExport.category.settings')}{#if file.hasSettings}&nbsp;({file.settingCount}){/if}</span>
      </label>
      <label class="check" class:disabled={!file.hasWhitelist}>
        <input type="checkbox" bind:checked={importWhitelist} disabled={!file.hasWhitelist} />
        <span>{$t('importExport.category.whitelist')}</span>
      </label>

      <button type="button" class="action-btn primary" disabled={importing} on:click={runImport}>
        {$t('importExport.importButton')}
      </button>
    {/if}
  </div>
</div>

<style>
  .section {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  .block {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    align-items: flex-start;
  }

  .block-head {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--text-secondary);
  }
  h4 {
    margin: 0;
    font-size: var(--fz-body);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .hint {
    margin: 0;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .check {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fz-label);
    color: var(--text-primary);
    cursor: pointer;
  }
  .check.disabled {
    color: var(--text-tertiary);
    cursor: default;
  }
  .check input {
    accent-color: var(--accent);
  }

  .action-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    margin-top: var(--space-1);
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
  .action-btn.primary {
    background: var(--accent);
    color: var(--accent-fg);
    border-color: transparent;
  }

  .file-info {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-2) 0;
  }
  .meta-row {
    display: flex;
    justify-content: space-between;
    gap: var(--space-4);
  }
  .meta-label {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }
  .meta-value {
    font-size: var(--fz-label);
    color: var(--text-primary);
  }
</style>
