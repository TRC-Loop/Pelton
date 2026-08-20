<script lang="ts">
  // Import from another mail client (#196). Two things live here because they
  // are one job for the user: bringing account settings across so they do not
  // retype six servers, and bringing mail files across so local-only mail is
  // not lost. Everything is user-initiated: a picker, or a profile Pelton found
  // where Thunderbird installs them. Nothing reads the network.
  import { createEventDispatcher, onDestroy, onMount } from 'svelte'
  import { fade, scale } from 'svelte/transition'
  import { IconX, IconMailbox, IconFileImport, IconFolderSearch } from '@tabler/icons-svelte'
  import Spinner from '../common/Spinner.svelte'
  import {
    findThunderbirdProfiles,
    chooseThunderbirdProfile,
    importThunderbirdAccounts,
    chooseMailFiles,
    importMailFiles,
    importThunderbirdFolders,
  } from '../../lib/api'
  import { onImportProgress, type ImportProgressEvent } from '../../lib/events'
  import { loadSidebar } from '../../stores/accounts'
  import { toastError, toastSuccess, errorMessage } from '../../stores/toast'
  import { formatBytes } from '../../lib/format'
  import type { ThunderbirdProfile } from '../../lib/types'
  import { get } from 'svelte/store'
  import { t } from '../../lib/i18n'

  const dispatch = createEventDispatcher<{ close: void }>()

  let profiles: ThunderbirdProfile[] = []
  let scanning = true
  // which profile's contents the list below is showing.
  let selected = 0
  // the picked accounts and folders, keyed by address and by path so the
  // choices survive switching profiles.
  let pickedAccounts = new Set<string>()
  let pickedFolders = new Set<string>()

  // the running import, or null when nothing is going.
  let progress: ImportProgressEvent | null = null
  $: running = progress?.running ?? false

  $: profile = profiles[selected]
  // a pop3 account cannot be recreated here, so it is shown greyed rather than
  // hidden: silently dropping it looks like Pelton missed it.
  $: importableAccounts = (profile?.accounts ?? []).filter((a) => a.kind === 'imap' && !a.exists)
  $: percent =
    progress && progress.bytesTotal > 0 ? Math.min(100, Math.round((progress.bytesDone / progress.bytesTotal) * 100)) : 0

  const stop = onImportProgress((event) => {
    progress = event
    if (!event.running) {
      finishImport(event)
    }
  })
  onDestroy(stop)

  onMount(async () => {
    try {
      profiles = await findThunderbirdProfiles()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      scanning = false
    }
  })

  function finishImport(event: ImportProgressEvent): void {
    if (event.error) {
      toastError(event.error)
      return
    }
    toastSuccess(`${event.imported} ${get(t)('import.messagesImported')}`)
    pickedFolders = new Set()
    // the Local Folders section appears the moment the first import lands.
    void loadSidebar()
  }

  async function pickProfile(): Promise<void> {
    try {
      const found = await chooseThunderbirdProfile()
      if (found.length === 0) {
        return
      }
      profiles = [...found, ...profiles]
      selected = 0
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  function toggle(set: Set<string>, key: string): Set<string> {
    const next = new Set(set)
    if (next.has(key)) {
      next.delete(key)
    } else {
      next.add(key)
    }
    return next
  }

  async function runAccountImport(): Promise<void> {
    if (!profile || pickedAccounts.size === 0) {
      return
    }
    try {
      const created = await importThunderbirdAccounts(profile.path, [...pickedAccounts])
      pickedAccounts = new Set()
      profiles = await findThunderbirdProfiles()
      await loadSidebar()
      toastSuccess(`${created} ${get(t)('import.accountsAdded')}`)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function runFolderImport(): Promise<void> {
    if (pickedFolders.size === 0) {
      return
    }
    try {
      await importThunderbirdFolders([...pickedFolders])
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function runFileImport(): Promise<void> {
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

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape' && !running) {
      dispatch('close')
    }
  }
</script>

<svelte:window on:keydown={onKeydown} />

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="backdrop" transition:fade={{ duration: 120 }} on:click={() => !running && dispatch('close')}></div>
<div
  class="dialog"
  role="dialog"
  aria-modal="true"
  aria-label={$t('import.title')}
  transition:scale={{ duration: 150, start: 0.94 }}
>
  <header>
    <h2>{$t('import.title')}</h2>
    <button
      type="button"
      class="close"
      aria-label={$t('detail.attachments.close')}
      disabled={running}
      on:click={() => dispatch('close')}
    >
      <IconX size={16} stroke={1.8} />
    </button>
  </header>

  <p class="hint">{$t('import.intro')}</p>

  <div class="body">
    <section class="block">
      <div class="block-head">
        <IconMailbox size={15} stroke={1.7} />
        <h3>{$t('import.thunderbird.title')}</h3>
      </div>

      {#if scanning}
        <Spinner label={$t('import.scanning')} />
      {:else if profiles.length === 0}
        <p class="sub-hint">{$t('import.thunderbird.none')}</p>
      {:else}
        {#if profiles.length > 1}
          <label class="field">
            <span class="field-label">{$t('import.thunderbird.profile')}</span>
            <select bind:value={selected}>
              {#each profiles as p, i}
                <option value={i}>{p.name}</option>
              {/each}
            </select>
          </label>
        {/if}

        {#if importableAccounts.length > 0}
          <p class="sub-hint">{$t('import.thunderbird.accountsHint')}</p>
          {#each importableAccounts as account (account.email)}
            <label class="check">
              <input
                type="checkbox"
                checked={pickedAccounts.has(account.email)}
                on:change={() => (pickedAccounts = toggle(pickedAccounts, account.email))}
              />
              <span class="row-main">
                <span class="row-title">{account.email}</span>
                <span class="row-meta">{account.imapHost}</span>
              </span>
            </label>
          {/each}
          <button type="button" class="action-btn" disabled={pickedAccounts.size === 0} on:click={runAccountImport}>
            {$t('import.thunderbird.addAccounts')}
          </button>
        {:else}
          <p class="sub-hint">{$t('import.thunderbird.noAccounts')}</p>
        {/if}

        {#if profile && profile.localFolders.length > 0}
          <p class="sub-hint">{$t('import.thunderbird.foldersHint')}</p>
          <div class="scroll-list">
            {#each profile.localFolders as folder (folder.path)}
              <label class="check">
                <input
                  type="checkbox"
                  checked={pickedFolders.has(folder.path)}
                  disabled={running}
                  on:change={() => (pickedFolders = toggle(pickedFolders, folder.path))}
                />
                <span class="row-main">
                  <span class="row-title">{folder.name}</span>
                  <span class="row-meta">{formatBytes(folder.sizeBytes)}</span>
                </span>
              </label>
            {/each}
          </div>
          <button
            type="button"
            class="action-btn"
            disabled={running || pickedFolders.size === 0}
            on:click={runFolderImport}
          >
            {$t('import.thunderbird.importFolders')}
          </button>
        {/if}
      {/if}

      <button type="button" class="link-btn" on:click={pickProfile}>
        <IconFolderSearch size={14} stroke={1.7} />
        {$t('import.thunderbird.choose')}
      </button>
    </section>

    <section class="block">
      <div class="block-head">
        <IconFileImport size={15} stroke={1.7} />
        <h3>{$t('import.files.title')}</h3>
      </div>
      <p class="sub-hint">{$t('import.files.hint')}</p>
      <button type="button" class="action-btn" disabled={running} on:click={runFileImport}>
        {$t('import.files.choose')}
      </button>
    </section>
  </div>

  {#if progress}
    <div class="progress" role="status">
      <div class="progress-line">
        <span>{running ? progress.folder || $t('import.importing') : $t('import.finished')}</span>
        <span class="row-meta">
          {progress.imported}
          {$t('import.messagesImported')}{#if running && progress.bytesTotal > 0}&nbsp;· {percent}%{/if}
        </span>
      </div>
      {#if running}
        <div class="bar"><div class="fill" style:width="{percent}%"></div></div>
      {/if}
      {#if !running && progress.skipped > 0}
        <p class="sub-hint">{$t('import.skipped')}: {progress.skipped}</p>
      {/if}
      {#if !running && progress.failed > 0}
        <p class="sub-hint">{$t('import.failed')}: {progress.failed}</p>
      {/if}
    </div>
  {/if}

  <div class="actions">
    <button type="button" class="action-btn primary" disabled={running} on:click={() => dispatch('close')}>
      {$t('detail.attachments.close')}
    </button>
  </div>
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
    width: min(520px, calc(100vw - 2 * var(--space-5)));
    max-height: calc(100vh - 2 * var(--space-5));
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
  h3 {
    margin: 0;
    font-size: var(--fz-body);
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
  .close:hover:not(:disabled) {
    background: var(--surface-hover);
    color: var(--text-primary);
  }
  .close:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .body {
    /* min-height:0 so a profile with many folders scrolls inside the card
       instead of pushing the actions off the bottom of the screen. */
    min-height: 0;
    overflow-y: auto;
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

  .hint {
    margin: 0;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .sub-hint {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    line-height: 1.4;
  }

  .field {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    align-self: stretch;
  }
  .field-label {
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }
  select {
    flex: 1;
    min-width: 0;
    height: var(--control-height);
    padding: 0 var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-primary);
    font-size: var(--fz-label);
  }

  .scroll-list {
    align-self: stretch;
    max-height: 180px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .check {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    align-self: stretch;
    font-size: var(--fz-label);
    color: var(--text-primary);
    cursor: pointer;
  }
  .check input {
    accent-color: var(--accent);
    flex-shrink: 0;
  }

  .row-main {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3);
    flex: 1;
    min-width: 0;
  }
  .row-title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .row-meta {
    flex-shrink: 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .progress {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-sunken);
  }
  .progress-line {
    display: flex;
    justify-content: space-between;
    gap: var(--space-3);
    font-size: var(--fz-label);
    color: var(--text-primary);
  }
  .bar {
    height: 4px;
    border-radius: 2px;
    background: var(--surface-raised);
    overflow: hidden;
  }
  .fill {
    height: 100%;
    background: var(--accent);
    transition: width 0.2s ease;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
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
  .action-btn:hover:not(:disabled) {
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

  .link-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    font-size: var(--fz-meta);
    cursor: pointer;
  }
  .link-btn:hover {
    color: var(--text-primary);
  }
</style>
