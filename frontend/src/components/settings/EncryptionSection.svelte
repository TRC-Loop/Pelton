<script lang="ts">
  // the Encryption settings (#192): the OpenPGP keys Pelton has imported, and
  // which private key each account signs with.
  //
  // Keys live in a directory next to the database, readable only by the user
  // who owns them. They are deliberately not part of the settings backup: a
  // private key belongs in a backup you make on purpose, not in a file people
  // mail around, so the only way one leaves is the export button here.
  import { onMount } from 'svelte'
  import {
    IconKey,
    IconLock,
    IconLockOpen,
    IconUpload,
    IconDownload,
    IconTrash,
    IconAlertTriangle,
  } from '@tabler/icons-svelte'
  import {
    listPGPKeys,
    importPGPKey,
    deletePGPKey,
    exportPGPKey,
    forgetPGPPassphrase,
    getAccountPGPKey,
    setAccountPGPKey,
    getSetting,
    setSetting,
    SettingKeys,
  } from '../../lib/api'
  import ToggleSwitch from '../common/ToggleSwitch.svelte'
  import { sidebar } from '../../stores/accounts'
  import { openPassphrase } from '../../stores/passphrase'
  import { errorMessage, toastError, toastSuccess, toastInfo } from '../../stores/toast'
  import { formatDateTimeMedium } from '../../lib/format'
  import { t } from '../../lib/i18n'
  import type { PGPKey } from '../../lib/types'

  let keys: PGPKey[] = []
  let loading = true
  let busy = false
  // the signing key each account is pinned to, by account id. '' means the
  // address is matched against the keys' user ids instead.
  let accountKeys: Record<number, string> = {}
  // whether search may look inside encrypted mail. read from the backend rather
  // than the prefs store, since it is not a display preference.
  let indexDecrypted = false
  let searchBusy = false

  $: accounts = $sidebar.data?.accounts ?? []
  $: signingKeys = keys.filter((k) => k.hasPrivate)

  onMount(async () => {
    await reload()
    try {
      indexDecrypted = (await getSetting(SettingKeys.indexDecrypted)).value === 'true'
    } catch {
      // never set: the default is off, which is what the toggle already shows.
    }
    loading = false
  })

  // onIndexDecrypted rebuilds the search index in the background either way,
  // so switching this off removes the plaintext already written rather than
  // only stopping new mail from being added.
  async function onIndexDecrypted(next: boolean): Promise<void> {
    searchBusy = true
    try {
      await setSetting(SettingKeys.indexDecrypted, next ? 'true' : 'false')
      indexDecrypted = next
      toastInfo($t('encryption.searchRebuilding'))
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      searchBusy = false
    }
  }

  async function reload(): Promise<void> {
    try {
      keys = await listPGPKeys()
    } catch (err) {
      toastError(errorMessage(err))
      return
    }
    const pairs = await Promise.all(
      accounts.map(async (a) => [a.id, await getAccountPGPKey(a.id).catch(() => '')] as const),
    )
    accountKeys = Object.fromEntries(pairs)
  }

  async function onImport(): Promise<void> {
    busy = true
    try {
      const added = await importPGPKey()
      if (added.length > 0) {
        toastSuccess($t('encryption.imported').replace('{n}', String(added.length)))
        await reload()
      }
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      busy = false
    }
  }

  async function onExport(key: PGPKey, includePrivate: boolean): Promise<void> {
    try {
      const path = await exportPGPKey(key.fingerprint, includePrivate)
      if (path) {
        toastSuccess($t('encryption.exported').replace('{path}', path))
      }
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function onDelete(key: PGPKey): Promise<void> {
    // deleting a private key destroys the only copy of it unless the user
    // exported one, and every message encrypted to it becomes unreadable.
    const prompt = key.hasPrivate ? 'encryption.confirmDeletePrivate' : 'encryption.confirmDelete'
    if (!window.confirm($t(prompt).replace('{name}', label(key)))) {
      return
    }
    try {
      await deletePGPKey(key.fingerprint)
      await reload()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  function onUnlock(key: PGPKey): void {
    openPassphrase(key, () => void reload())
  }

  async function onForget(key: PGPKey): Promise<void> {
    try {
      await forgetPGPPassphrase(key.fingerprint)
      toastInfo($t('encryption.forgotten'))
      await reload()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function onAccountKey(accountId: number, fingerprint: string): Promise<void> {
    try {
      await setAccountPGPKey(accountId, fingerprint)
      accountKeys = { ...accountKeys, [accountId]: fingerprint }
    } catch (err) {
      toastError(errorMessage(err))
      await reload()
    }
  }

  /** A key date, rendered as a plain day: the time of day means nothing here. */
  function keyDate(iso: string): string {
    return formatDateTimeMedium(new Date(iso), false)
  }

  /** The name to show for a key: its user id, falling back to the key id. */
  function label(key: PGPKey): string {
    if (key.name && key.email) {
      return `${key.name} <${key.email}>`
    }
    return key.email || key.name || shortId(key.fingerprint)
  }

  /** The last sixteen hex digits, which is how gpg prints a long key id. */
  function shortId(fingerprint: string): string {
    return fingerprint.length > 16 ? fingerprint.slice(-16) : fingerprint
  }

  /** Groups the fingerprint into fours so it can be read aloud and compared. */
  function spaced(fingerprint: string): string {
    return (fingerprint.match(/.{1,4}/g) ?? []).join(' ')
  }
</script>

<div class="section">
  <p class="lead">{$t('encryption.lead')}</p>

  <div class="actions">
    <button type="button" class="primary" disabled={busy} on:click={onImport}>
      <IconUpload size={15} stroke={1.7} />
      {$t('encryption.import')}
    </button>
  </div>

  {#if loading}
    <p class="hint">{$t('common.loading')}</p>
  {:else if keys.length === 0}
    <p class="hint">{$t('encryption.empty')}</p>
  {:else}
    <ul class="keys">
      {#each keys as key (key.fingerprint)}
        <li class="key" class:expired={key.expired}>
          <span class="glyph" class:private={key.hasPrivate}>
            {#if key.locked && !key.unlocked && !key.remembered}
              <IconLock size={16} stroke={1.7} />
            {:else if key.hasPrivate}
              <IconLockOpen size={16} stroke={1.7} />
            {:else}
              <IconKey size={16} stroke={1.7} />
            {/if}
          </span>

          <div class="body">
            <div class="title">
              <span class="name">{label(key)}</span>
              {#if key.hasPrivate}
                <span class="tag">{$t('encryption.tag.private')}</span>
              {/if}
              {#if key.expired}
                <span class="tag danger">{$t('encryption.tag.expired')}</span>
              {:else if key.remembered}
                <span class="tag">{$t('encryption.tag.remembered')}</span>
              {:else if key.unlocked}
                <span class="tag">{$t('encryption.tag.unlocked')}</span>
              {/if}
            </div>

            <div class="meta">
              <code>{spaced(key.fingerprint)}</code>
            </div>
            <div class="meta">
              {key.algorithm}{key.bits ? ` ${key.bits}` : ''}
              &middot; {$t('encryption.created').replace('{date}', keyDate(key.created))}
              {#if key.expires}
                &middot; {$t(key.expired ? 'encryption.expiredOn' : 'encryption.expiresOn').replace(
                  '{date}',
                  keyDate(key.expires),
                )}
              {/if}
            </div>
            {#if key.emails.length > 1}
              <div class="meta">{key.emails.join(', ')}</div>
            {/if}
          </div>

          <div class="row-actions">
            {#if key.hasPrivate && key.locked}
              {#if key.unlocked || key.remembered}
                <button type="button" class="ghost" on:click={() => void onForget(key)}>
                  {$t('encryption.forget')}
                </button>
              {:else}
                <button type="button" class="ghost" on:click={() => onUnlock(key)}>
                  {$t('encryption.unlock')}
                </button>
              {/if}
            {/if}
            <button
              type="button"
              class="ghost"
              title={$t('encryption.exportPublic')}
              aria-label={$t('encryption.exportPublic')}
              on:click={() => void onExport(key, false)}
            >
              <IconDownload size={15} stroke={1.7} />
            </button>
            {#if key.hasPrivate}
              <button type="button" class="ghost" on:click={() => void onExport(key, true)}>
                {$t('encryption.exportPrivate')}
              </button>
            {/if}
            <button
              type="button"
              class="ghost danger"
              title={$t('action.delete')}
              aria-label={$t('action.delete')}
              on:click={() => void onDelete(key)}
            >
              <IconTrash size={15} stroke={1.7} />
            </button>
          </div>
        </li>
      {/each}
    </ul>

    <div class="note">
      <span class="note-icon"><IconAlertTriangle size={16} stroke={1.8} /></span>
      <p>{$t('encryption.backupNote')}</p>
    </div>
  {/if}

  {#if signingKeys.length > 0 && accounts.length > 0}
    <h4>{$t('encryption.signingTitle')}</h4>
    <p class="hint">{$t('encryption.signingHint')}</p>
    {#each accounts as account (account.id)}
      <div class="account">
        <span class="row-label">{account.email}</span>
        <select
          value={accountKeys[account.id] ?? ''}
          on:change={(e) => void onAccountKey(account.id, e.currentTarget.value)}
        >
          <option value="">{$t('encryption.signingAuto')}</option>
          {#each signingKeys as key (key.fingerprint)}
            <option value={key.fingerprint}>{label(key)} ({shortId(key.fingerprint)})</option>
          {/each}
        </select>
      </div>
    {/each}
  {/if}

  <h4>{$t('encryption.searchTitle')}</h4>
  <div class="account">
    <span class="row-label">{$t('encryption.searchLabel')}</span>
    <ToggleSwitch
      checked={indexDecrypted}
      label={$t('encryption.searchLabel')}
      disabled={searchBusy}
      on:change={(e) => void onIndexDecrypted(e.detail)}
    />
  </div>
  <p class="hint">{$t('encryption.searchHint')}</p>
</div>

<style>
  .section {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .lead,
  .hint {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  h4 {
    margin: var(--space-3) 0 0;
    font-size: var(--fz-body);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .actions {
    display: flex;
    gap: var(--space-2);
  }

  .primary {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--accent-fg);
    background: var(--accent);
    border: none;
    border-radius: var(--radius-control);
    cursor: pointer;
  }
  .primary:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .keys {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .key {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
  }
  .key.expired {
    opacity: 0.7;
  }

  .glyph {
    display: inline-flex;
    flex-shrink: 0;
    margin-top: 2px;
    color: var(--text-tertiary);
  }
  .glyph.private {
    color: var(--accent);
  }

  .body {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .title {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: var(--space-2);
  }

  .name {
    font-size: var(--fz-body);
    color: var(--text-primary);
  }

  .tag {
    padding: 0 var(--space-1);
    border-radius: var(--radius-control);
    background: var(--surface-hover);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }
  .tag.danger {
    background: var(--danger-bg);
    color: var(--danger);
  }

  .meta {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    overflow-wrap: anywhere;
  }

  .meta code {
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
  }

  .row-actions {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    gap: var(--space-1);
  }

  .ghost {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1) var(--space-2);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
    background: transparent;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    cursor: pointer;
  }
  .ghost:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }
  .ghost.danger {
    color: var(--danger);
  }
  .ghost.danger:hover {
    background: var(--danger-bg);
    color: var(--danger);
  }

  .note {
    display: flex;
    gap: var(--space-3);
    padding: var(--space-3);
    border-radius: var(--radius-control);
    background: var(--warning-bg);
  }

  .note-icon {
    display: inline-flex;
    flex-shrink: 0;
    color: var(--warning);
  }

  .note p {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .account {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .row-label {
    font-size: var(--fz-body);
    color: var(--text-primary);
  }

  .account select {
    max-width: 60%;
    padding: var(--space-1) var(--space-2);
    font-family: var(--font-ui);
    font-size: var(--fz-meta);
    color: var(--text-primary);
    background: var(--surface-raised);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
  }
</style>
