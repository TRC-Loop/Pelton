<script lang="ts">
  // where CardDAV address books are configured (#168). A book can be found from
  // a mailbox's own domain, which is one click on a server that hosts both, or
  // pointed at by hand, which is the Nextcloud case: contacts and mail are
  // usually not on the same host.
  import { onMount } from 'svelte'
  import { IconPlus, IconTrash, IconPencil, IconRefresh, IconAlertTriangle } from '@tabler/icons-svelte'
  import Modal from '../common/Modal.svelte'
  import ToggleSwitch from '../common/ToggleSwitch.svelte'
  import Spinner from '../common/Spinner.svelte'
  import {
    listAccounts,
    discoverAddressBooks,
    addAddressBook,
    updateAddressBook,
    removeAddressBook,
  } from '../../lib/api'
  import { addressBooks, loadContacts, refreshContacts } from '../../stores/contacts'
  import { prefs, setHarvestAddresses } from '../../stores/prefs'
  import { errorMessage, toastError, toastSuccess } from '../../stores/toast'
  import { accountLabel } from '../../lib/format'
  import { formatRelative } from '../../lib/format'
  import { t } from '../../lib/i18n'
  import type { Account, AddressBookDraft, DiscoveredBook } from '../../lib/types'

  let accounts: Account[] = []
  let loading = true
  // the add/edit form, or null when closed.
  let draft: AddressBookDraft | null = null
  let opened = ''
  // what a discovery run found, and whether one is in flight.
  let found: DiscoveredBook[] = []
  let discovering = false
  let saving = false
  let confirmingId: number | null = null

  $: dirty = draft !== null && JSON.stringify(draft) !== opened

  onMount(async () => {
    try {
      accounts = await listAccounts()
      await loadContacts()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      loading = false
    }
  })

  function startAdd(): void {
    found = []
    draft = { id: 0, accountId: 0, name: '', url: '', collectionPath: '', username: '', password: '' }
    opened = JSON.stringify(draft)
  }

  function startEdit(id: number): void {
    const book = $addressBooks.find((b) => b.id === id)
    if (!book) {
      return
    }
    found = []
    draft = {
      id: book.id,
      accountId: book.accountId,
      name: book.name,
      url: book.url,
      collectionPath: book.collectionPath,
      username: book.username,
      // never prefilled: the stored password is not readable, and an empty one
      // means "leave it alone".
      password: '',
    }
    opened = JSON.stringify(draft)
  }

  async function discover(): Promise<void> {
    if (!draft) {
      return
    }
    discovering = true
    try {
      found = await discoverAddressBooks(draft)
      if (found.length === 0) {
        toastError($t('contacts.books.noneFound'))
      }
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      discovering = false
    }
  }

  function choose(book: DiscoveredBook): void {
    if (!draft || book.exists) {
      return
    }
    draft.url = book.url
    draft.collectionPath = book.collectionPath
    if (draft.name === '') {
      draft.name = book.name
    }
  }

  async function save(): Promise<void> {
    if (!draft) {
      return
    }
    saving = true
    try {
      if (draft.id === 0) {
        await addAddressBook(draft)
      } else {
        await updateAddressBook(draft)
      }
      await loadContacts()
      draft = null
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      saving = false
    }
  }

  async function confirmRemove(id: number): Promise<void> {
    try {
      await removeAddressBook(id)
      confirmingId = null
      await loadContacts()
      toastSuccess($t('contacts.books.removed'))
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  function accountName(id: number): string {
    const account = accounts.find((a) => a.id === id)
    return account ? accountLabel(account) : ''
  }
</script>

<div class="head">
  <div>
    <h3>{$t('settingsPanel.category.contacts')}</h3>
    <p class="hint">{$t('contacts.books.hint')}</p>
  </div>
  <div class="head-actions">
    <button type="button" class="ghost" on:click={refreshContacts}>
      <IconRefresh size={14} stroke={1.8} />
      {$t('contacts.sync')}
    </button>
    <button type="button" class="add-btn" on:click={startAdd}>
      <IconPlus size={14} stroke={2} />
      {$t('contacts.books.add')}
    </button>
  </div>
</div>

{#if loading}
  <p class="empty">{$t('mailboxes.loading')}</p>
{:else if $addressBooks.length === 0}
  <p class="empty">{$t('contacts.books.empty')}</p>
{:else}
  <ul class="list">
    {#each $addressBooks as book (book.id)}
      <li>
        <div class="who">
          <span class="name">{book.name || $t('contacts.book.untitled')}</span>
          <span class="addr">
            {book.url}{book.accountId ? ` · ${accountName(book.accountId)}` : ''}
          </span>
          <span class="addr">
            {$t('contacts.books.count').replace('{n}', String(book.contactCount))}
            {#if book.lastSync}
              · {$t('contacts.books.synced').replace('{when}', formatRelative(Date.parse(book.lastSync), $t))}
            {/if}
          </span>
          {#if book.lastError}
            <span class="err">{book.lastError}</span>
          {/if}
        </div>
        {#if !book.hasPassword}
          <span class="warn-icon" title={$t('contacts.books.needsPassword')}>
            <IconAlertTriangle size={15} stroke={1.8} />
          </span>
        {/if}
        {#if confirmingId === book.id}
          <div class="confirm">
            <span class="warn">{$t('contacts.books.removeConfirm')}</span>
            <button type="button" class="danger" on:click={() => confirmRemove(book.id)}>{$t('action.remove')}</button>
            <button type="button" class="ghost" on:click={() => (confirmingId = null)}>{$t('mailboxes.cancel')}</button>
          </div>
        {:else}
          <button type="button" class="icon" aria-label={$t('mailboxes.edit')} on:click={() => startEdit(book.id)}>
            <IconPencil size={15} stroke={1.7} />
          </button>
          <button type="button" class="icon del" aria-label={$t('action.remove')} on:click={() => (confirmingId = book.id)}>
            <IconTrash size={15} stroke={1.7} />
          </button>
        {/if}
      </li>
    {/each}
  </ul>
{/if}

<div class="toggle">
  <span>{$t('contacts.harvest.toggle')}</span>
  <ToggleSwitch
    checked={$prefs.harvestAddresses}
    label={$t('contacts.harvest.toggle')}
    on:change={(e) => setHarvestAddresses(e.detail)}
  />
</div>
<p class="hint">{$t('contacts.harvest.hint')}</p>

{#if draft}
  <Modal
    title={draft.id === 0 ? $t('contacts.books.add') : $t('contacts.books.edit')}
    hint={$t('contacts.books.formHint')}
    size="medium"
    {dirty}
    on:close={() => (draft = null)}
  >
    <div class="form">
      <label class="field">
        <span>{$t('contacts.books.field.name')}</span>
        <input type="text" bind:value={draft.name} placeholder={$t('contacts.book.untitled')} />
      </label>
      <label class="field">
        <span>{$t('contacts.books.field.url')}</span>
        <input type="text" bind:value={draft.url} placeholder="cloud.example.com" />
      </label>
      <label class="field">
        <span>{$t('contacts.books.field.account')}</span>
        <select bind:value={draft.accountId}>
          <option value={0}>{$t('contacts.books.field.noAccount')}</option>
          {#each accounts.filter((a) => !a.local) as account (account.id)}
            <option value={account.id}>{accountLabel(account)}</option>
          {/each}
        </select>
      </label>
      <p class="hint">{$t('contacts.books.accountHint')}</p>

      <div class="pair">
        <label class="field">
          <span>{$t('contacts.books.field.username')}</span>
          <input type="text" bind:value={draft.username} />
        </label>
        <label class="field">
          <span>{$t('contacts.books.field.password')}</span>
          <input type="password" bind:value={draft.password} placeholder={draft.id === 0 ? '' : $t('contacts.books.field.passwordKept')} />
        </label>
      </div>

      <button type="button" class="ghost" on:click={discover} disabled={discovering}>
        {discovering ? $t('contacts.books.discovering') : $t('contacts.books.discover')}
      </button>

      {#if discovering}
        <div class="loading"><Spinner /></div>
      {:else if found.length > 0}
        <span class="group-label">{$t('contacts.books.found')}</span>
        <ul class="found">
          {#each found as book (book.collectionPath)}
            <li>
              <button type="button" class="found-row" class:taken={book.exists} disabled={book.exists} on:click={() => choose(book)}>
                <span>{book.name || book.collectionPath}</span>
                {#if book.exists}<span class="addr">{$t('contacts.books.alreadyAdded')}</span>{/if}
              </button>
            </li>
          {/each}
        </ul>
      {/if}

      <label class="field">
        <span>{$t('contacts.books.field.collection')}</span>
        <input type="text" bind:value={draft.collectionPath} placeholder="/remote.php/dav/addressbooks/users/me/contacts/" />
      </label>

      <div class="actions">
        <button type="button" class="ghost" on:click={() => (draft = null)}>{$t('mailboxes.cancel')}</button>
        <button type="button" class="primary" disabled={saving || draft.url.trim() === ''} on:click={save}>
          {saving ? $t('contacts.books.saving') : $t('action.save')}
        </button>
      </div>
    </div>
  </Modal>
{/if}

<style>
  .head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
  }

  .head-actions {
    display: flex;
    gap: var(--space-2);
  }

  h3 {
    margin: 0 0 var(--space-1);
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
  }

  .hint {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .list {
    list-style: none;
    margin: 0 0 var(--space-4);
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .list li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2);
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-card);
    background: var(--surface-raised);
  }

  .who {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .name {
    font-size: var(--fz-body);
  }

  .addr {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .err {
    font-size: var(--fz-meta);
    color: var(--danger);
  }

  .warn-icon {
    display: inline-flex;
    color: var(--warning);
  }

  .toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    font-size: var(--fz-label);
    margin-top: var(--space-4);
  }

  .form {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    width: 100%;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .field span,
  .group-label {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .pair {
    display: flex;
    gap: var(--space-3);
  }

  .pair .field {
    flex: 1;
  }

  input,
  select {
    height: var(--control-height);
    padding: 0 var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font: inherit;
    font-size: var(--fz-body);
  }

  .found {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .found-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    width: 100%;
    padding: var(--space-1) var(--space-2);
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-primary);
    font: inherit;
    font-size: var(--fz-label);
    text-align: left;
    cursor: var(--cursor-action);
  }

  .found-row.taken {
    opacity: 0.5;
    cursor: default;
  }

  .loading {
    display: flex;
    justify-content: center;
    padding: var(--space-3);
  }

  .empty {
    margin: 0 0 var(--space-4);
    font-size: var(--fz-body);
    color: var(--text-tertiary);
  }

  .confirm {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fz-meta);
  }

  .warn {
    color: var(--danger);
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-top: var(--space-2);
  }

  button.primary,
  button.ghost,
  button.danger,
  .add-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    height: var(--control-height);
    padding: 0 var(--space-3);
    border-radius: var(--radius-control);
    border: var(--hairline) solid var(--border-default);
    background: var(--surface-raised);
    color: var(--text-primary);
    font: inherit;
    font-size: var(--fz-label);
    cursor: var(--cursor-action);
  }

  button.primary,
  .add-btn {
    background: var(--accent);
    color: var(--accent-fg);
    border-color: transparent;
  }

  button.danger {
    background: var(--danger);
    color: var(--text-inverse);
    border-color: transparent;
  }

  button:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: var(--control-height);
    height: var(--control-height);
    border: none;
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--text-secondary);
    cursor: var(--cursor-action);
  }

  .icon:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .icon.del:hover {
    color: var(--danger);
  }
</style>
