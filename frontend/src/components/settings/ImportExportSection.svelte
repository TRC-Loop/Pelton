<script lang="ts">
  // Import/export replaces the removed folder-based config sync with a plain,
  // local file: the user picks what to export, and on import picks a file and
  // chooses what to bring in. Pelton only reads and writes the file the user
  // chooses; nothing leaves the machine on its own.
  import { IconDownload, IconUpload, IconFileImport } from '@tabler/icons-svelte'
  import { inspectBackupFile, importData, type BackupInfo } from '../../lib/api'
  import { initPrefs } from '../../stores/prefs'
  import { toastError, toastSuccess, errorMessage } from '../../stores/toast'
  import { formatFullDate } from '../../lib/format'
  import { get } from 'svelte/store'
  import { t } from '../../lib/i18n'

  let exportOpen = false

  // import state: the inspected file and which of its categories to apply.
  let file: BackupInfo | null = null
  let importSettings = true
  let importWhitelist = true
  let importMailboxes = false
  let importSignatures = true
  let importing = false

  // credential password: only relevant when the picked file actually carries
  // encrypted mailbox credentials and the user wants them restored too.
  let importPasswords = false
  let credentialPassword = ''

  async function chooseFile(): Promise<void> {
    try {
      const info = await inspectBackupFile()
      if (!info.path) {
        return
      }
      file = info
      importSettings = info.hasSettings
      importWhitelist = info.hasWhitelist
      importMailboxes = false
      importSignatures = info.hasSignatures
      importPasswords = false
      credentialPassword = ''
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
      ...(importMailboxes && file.hasMailboxes ? ['mailboxes'] : []),
      ...(importSignatures && file.hasSignatures ? ['signatures'] : []),
    ]
    if (categories.length === 0) {
      toastError(get(t)('importExport.nothingSelected'))
      return
    }
    if (importPasswords && credentialPassword.length === 0) {
      toastError(get(t)('importExport.passwordInvalid'))
      return
    }
    importing = true
    try {
      await importData(file.path, categories, importPasswords ? credentialPassword : '')
      // settings we just wrote drive the ui; reload them so it updates live.
      await initPrefs()
      toastSuccess(get(t)('importExport.imported'))
      file = null
      credentialPassword = ''
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
    <button type="button" class="action-btn primary" on:click={() => (exportOpen = true)}>
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
      <label class="check" class:disabled={!file.hasMailboxes}>
        <input type="checkbox" bind:checked={importMailboxes} disabled={!file.hasMailboxes} />
        <span>{$t('importExport.category.mailboxes')}{#if file.hasMailboxes}&nbsp;({file.mailboxCount}){/if}</span>
      </label>
      {#if importMailboxes && file.hasMailboxes}
        <p class="sub-hint">{$t('importExport.mailboxesHint')}</p>
        {#if file.hasEncryptedCredentials}
          <label class="check sub">
            <input type="checkbox" bind:checked={importPasswords} />
            <span>{$t('importExport.restorePasswords')}</span>
          </label>
          {#if importPasswords}
            <input
              type="password"
              class="pw-input"
              placeholder={$t('importExport.passwordPlaceholder')}
              autocomplete="current-password"
              bind:value={credentialPassword}
            />
          {/if}
        {/if}
      {/if}
      <label class="check" class:disabled={!file.hasSignatures}>
        <input type="checkbox" bind:checked={importSignatures} disabled={!file.hasSignatures} />
        <span>{$t('importExport.category.signatures')}{#if file.hasSignatures}&nbsp;({file.signatureCount}){/if}</span>
      </label>

      <button
        type="button"
        class="action-btn primary"
        disabled={importing || (importPasswords && credentialPassword.length === 0)}
        on:click={runImport}
      >
        {$t('importExport.importButton')}
      </button>
    {/if}
  </div>
</div>

<!-- the export modal is code-split so its logic loads only on demand. -->
{#if exportOpen}
  {#await import('./ExportModal.svelte') then m}
    <svelte:component this={m.default} on:close={() => (exportOpen = false)} />
  {/await}
{/if}

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

  .sub-hint {
    margin: 0 0 0 calc(var(--space-2) + 14px);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    line-height: 1.4;
  }

  .check.sub {
    margin-left: calc(var(--space-2) + 14px);
  }

  .pw-input {
    margin-left: calc(var(--space-2) + 14px);
    height: var(--control-height);
    padding: 0 var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-primary);
    font-size: var(--fz-list);
  }
  .pw-input:focus {
    border-color: var(--accent);
    outline: none;
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
