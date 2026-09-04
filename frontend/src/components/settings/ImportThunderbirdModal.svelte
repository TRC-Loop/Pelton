<script lang="ts">
  // Moving in from Thunderbird (#196, rebuilt for #308). This used to share a
  // dialog with "import some mail files", which are not the same job: one is a
  // migration you do once when you switch clients, the other is loading files
  // and can happen any day.
  //
  // Two steps, because the old single screen made you commit before you could
  // see what you were committing to. Step one lists the profiles found with
  // enough on each card to tell them apart: how many accounts, which addresses,
  // how many mailboxes and how big. Step two is that profile's contents with
  // checkboxes, so what comes across is a choice rather than all or nothing.
  //
  // Nothing here reads the network, and no password is read: Thunderbird keeps
  // its own, and each account is re-authenticated once after it lands.
  import { createEventDispatcher, onDestroy, onMount } from 'svelte'
  import Modal from '../common/Modal.svelte'
  import {
    IconMailbox,
    IconFolderSearch,
    IconChevronLeft,
    IconAlertTriangle,
    IconCheck,
  } from '@tabler/icons-svelte'
  import Spinner from '../common/Spinner.svelte'
  import {
    findThunderbirdProfiles,
    chooseThunderbirdProfile,
    importThunderbirdAccounts,
    importThunderbirdFolders,
  } from '../../lib/api'
  import { onImportProgress, type ImportProgressEvent } from '../../lib/events'
  import { loadSidebar } from '../../stores/accounts'
  import { promptForMissingPasswords } from '../../stores/passwordprompt'
  import { toastError, toastSuccess, errorMessage } from '../../stores/toast'
  import { formatBytes } from '../../lib/format'
  import type { ThunderbirdProfile } from '../../lib/types'
  import { get } from 'svelte/store'
  import { t } from '../../lib/i18n'

  const dispatch = createEventDispatcher<{ close: void }>()

  let profiles: ThunderbirdProfile[] = []
  let scanning = true
  // null while choosing a profile, otherwise the one being looked at.
  let chosen: ThunderbirdProfile | null = null

  // the picked accounts and mailboxes, keyed by address and by path so a step
  // back and forward again keeps what was ticked.
  let pickedAccounts = new Set<string>()
  let pickedFolders = new Set<string>()

  let progress: ImportProgressEvent | null = null
  $: running = progress?.running ?? false

  // a pop3 account cannot be recreated (Pelton speaks imap), and one already
  // configured would be a duplicate. Both are listed rather than hidden, with
  // the reason next to them: a missing account looks like Pelton failed to see
  // it.
  $: accounts = chosen?.accounts ?? []
  $: importable = accounts.filter((a) => a.kind === 'imap' && !a.exists)
  $: folders = chosen?.localFolders ?? []
  $: pickedBytes = folders
    .filter((f) => pickedFolders.has(f.path))
    .reduce((sum, f) => sum + f.sizeBytes, 0)

  $: percent =
    progress && progress.bytesTotal > 0
      ? Math.min(100, Math.round((progress.bytesDone / progress.bytesTotal) * 100))
      : 0

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
      // one profile is not a choice, so it opens straight into its contents.
      if (profiles.length === 1) {
        chosen = profiles[0]
      }
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

  // profileSummary is the line under a profile's name on the chooser: what is
  // in it, in the terms you would use to decide which one you want.
  function profileSummary(p: ThunderbirdProfile): string {
    const accountCount = p.accounts.length
    const folderCount = p.localFolders.length
    const bytes = p.localFolders.reduce((sum, f) => sum + f.sizeBytes, 0)
    const parts = [
      accountCount === 1
        ? $t('import.tb.oneAccount')
        : $t('import.tb.nAccounts').replace('{n}', String(accountCount)),
      folderCount === 1
        ? $t('import.tb.oneMailbox')
        : $t('import.tb.nMailboxes').replace('{n}', String(folderCount)),
    ]
    if (bytes > 0) {
      parts.push(formatBytes(bytes))
    }
    return parts.join(' · ')
  }

  // the addresses on a profile card, capped so a profile with a dozen accounts
  // does not push the next card off the screen.
  function addressLine(p: ThunderbirdProfile): string {
    const shown = p.accounts.slice(0, 3).map((a) => a.email)
    if (p.accounts.length > shown.length) {
      shown.push($t('import.tb.andMore').replace('{n}', String(p.accounts.length - shown.length)))
    }
    return shown.join(', ')
  }

  async function pickProfile(): Promise<void> {
    try {
      const found = await chooseThunderbirdProfile()
      if (found.length === 0) {
        return
      }
      profiles = [...found, ...profiles]
      chosen = found[0]
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  function back(): void {
    chosen = null
    progress = null
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

  function toggleAllAccounts(): void {
    pickedAccounts =
      pickedAccounts.size === importable.length ? new Set() : new Set(importable.map((a) => a.email))
  }

  function toggleAllFolders(): void {
    pickedFolders = pickedFolders.size === folders.length ? new Set() : new Set(folders.map((f) => f.path))
  }

  async function runAccountImport(): Promise<void> {
    if (!chosen || pickedAccounts.size === 0) {
      return
    }
    try {
      const created = await importThunderbirdAccounts(chosen.path, [...pickedAccounts])
      pickedAccounts = new Set()
      // re-reading marks the imported ones as existing, so they drop out of the
      // list rather than being offered a second time.
      profiles = await findThunderbirdProfiles()
      chosen = profiles.find((p) => p.path === chosen?.path) ?? chosen
      await loadSidebar()
      toastSuccess(`${created} ${get(t)('import.accountsAdded')}`)
      // Thunderbird keeps its own passwords, so an imported mailbox arrives
      // without one and cannot sync until it has one. Asking here, while the
      // user is still setting up, is the difference between a mailbox that
      // works and one that quietly never receives anything.
      if (created > 0) {
        await promptForMissingPasswords()
      }
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
</script>

<Modal
  title={$t('import.tb.title')}
  hint={chosen ? chosen.name : $t('import.tb.intro')}
  size="medium"
  closable={!running}
  on:close={() => dispatch('close')}
>
  <span slot="icon"><IconMailbox size={16} stroke={1.8} /></span>

  <div class="body">
    {#if scanning}
      <Spinner label={$t('import.scanning')} />
    {:else if !chosen}
      <!-- step one: which profile. -->
      {#if profiles.length === 0}
        <p class="hint">{$t('import.tb.none')}</p>
      {:else}
        <ul class="profiles">
          {#each profiles as p (p.path)}
            <li>
              <button type="button" class="profile" on:click={() => (chosen = p)}>
                <span class="p-name">{p.name}</span>
                <span class="p-summary">{profileSummary(p)}</span>
                {#if p.accounts.length > 0}
                  <span class="p-addresses">{addressLine(p)}</span>
                {/if}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
      <button type="button" class="link-btn" on:click={pickProfile}>
        <IconFolderSearch size={14} stroke={1.7} />
        {$t('import.tb.choose')}
      </button>
    {:else}
      <!-- step two: what comes across. -->
      <button type="button" class="link-btn back" disabled={running} on:click={back}>
        <IconChevronLeft size={14} stroke={1.8} />
        {$t('import.tb.backToProfiles')}
      </button>

      <section class="block">
        <header class="block-head">
          <h3>{$t('import.tb.accounts')}</h3>
          {#if importable.length > 0}
            <button type="button" class="link-btn" on:click={toggleAllAccounts}>
              {pickedAccounts.size === importable.length ? $t('import.tb.none_') : $t('import.tb.all')}
            </button>
          {/if}
        </header>

        {#if accounts.length === 0}
          <p class="hint">{$t('import.tb.noAccounts')}</p>
        {:else}
          {#each accounts as account (account.email)}
            {@const blocked = account.kind !== 'imap' || account.exists}
            <label class="row" class:blocked>
              <input
                type="checkbox"
                checked={pickedAccounts.has(account.email)}
                disabled={blocked}
                on:change={() => (pickedAccounts = toggle(pickedAccounts, account.email))}
              />
              <span class="row-main">
                <span class="row-title">{account.email}</span>
                <span class="row-meta">
                  {#if account.exists}
                    <IconCheck size={12} stroke={2} /> {$t('import.tb.alreadyHere')}
                  {:else if account.kind !== 'imap'}
                    <IconAlertTriangle size={12} stroke={1.9} />
                    {$t('import.tb.pop3').replace('{kind}', account.kind.toUpperCase())}
                  {:else}
                    {account.imapHost}
                  {/if}
                </span>
              </span>
            </label>
          {/each}
          <button
            type="button"
            class="action-btn"
            disabled={running || pickedAccounts.size === 0}
            on:click={runAccountImport}
          >
            {pickedAccounts.size === 1
              ? $t('import.tb.addOne')
              : $t('import.tb.addN').replace('{n}', String(pickedAccounts.size))}
          </button>
          <p class="hint">{$t('import.tb.passwordNote')}</p>
        {/if}
      </section>

      <section class="block">
        <header class="block-head">
          <h3>{$t('import.tb.mailboxes')}</h3>
          {#if folders.length > 0}
            <button type="button" class="link-btn" disabled={running} on:click={toggleAllFolders}>
              {pickedFolders.size === folders.length ? $t('import.tb.none_') : $t('import.tb.all')}
            </button>
          {/if}
        </header>

        {#if folders.length === 0}
          <p class="hint">{$t('import.tb.noMailboxes')}</p>
        {:else}
          <div class="scroll-list">
            {#each folders as folder (folder.path)}
              <label class="row">
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
            {pickedFolders.size === 1
              ? $t('import.tb.importOne')
              : $t('import.tb.importN')
                  .replace('{n}', String(pickedFolders.size))
                  .replace('{size}', formatBytes(pickedBytes))}
          </button>
        {/if}
      </section>
    {/if}

    {#if progress}
      <div class="progress" role="status" aria-live="polite">
        <div class="progress-line">
          <span class="p-where">
            {#if running}
              {progress.fileTotal > 1
                ? $t('import.progress.mailboxOf')
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
    /* min-height:0 so a profile with many mailboxes scrolls inside the card
       instead of pushing the actions off the bottom of the screen. */
    min-height: 0;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  h3 {
    margin: 0;
    font-size: var(--fz-body);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .hint {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .profiles {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  /* a card rather than a line of text: which profile you want is decided by
     what is in it, so what is in it has to be on the card. */
  .profile {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    width: 100%;
    padding: var(--space-3);
    text-align: left;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-raised);
    cursor: var(--cursor-action);
  }
  .profile:hover {
    border-color: var(--accent);
    background: var(--surface-hover);
  }

  .p-name {
    font-size: var(--fz-body);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }

  .p-summary {
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .p-addresses {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .block {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .block-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-1) 0;
    cursor: var(--cursor-action);
  }

  .row.blocked {
    cursor: default;
    opacity: 0.55;
  }

  .row-main {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .row-title {
    font-size: var(--fz-label);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .row-meta {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .scroll-list {
    max-height: 220px;
    overflow-y: auto;
    padding-right: var(--space-2);
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

  .link-btn {
    align-self: flex-start;
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: 0;
    font-size: var(--fz-meta);
    color: var(--accent);
    background: transparent;
    border: none;
    cursor: var(--cursor-action);
  }
  .link-btn:disabled {
    color: var(--text-tertiary);
    cursor: default;
  }

  .back {
    color: var(--text-secondary);
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
