<script lang="ts">
  // the mailbox manager: lists configured accounts and lets the user edit an
  // account's display name and server settings, or delete it outright. Deleting
  // is destructive (it drops the cached mail and the keyring secret), so it goes
  // through an inline confirm. Email is not editable here, it identifies the
  // account; changing it is a re-add.
  import { onMount } from 'svelte'
  import { IconPencil, IconTrash, IconCheck, IconX, IconPlus } from '@tabler/icons-svelte'
  import { listAccounts, updateAccount, deleteAccount } from '../../lib/api'
  import { refreshSidebar } from '../../stores/accounts'
  import { errorMessage, toastError } from '../../stores/toast'
  import type { Account } from '../../lib/types'
  import { t } from '../../lib/i18n'

  let accounts: Account[] = []
  let loading = true
  let editingId: number | null = null
  let confirmingId: number | null = null
  let saving = false
  // the working copy of the account being edited, so cancelling discards edits.
  let draft: Account | null = null
  // the add-mailbox wizard is code-split like the other settings modals, so
  // it only loads once the user actually asks to add a mailbox.
  let wizardOpen = false

  onMount(load)

  function onMailboxAdded(): void {
    wizardOpen = false
    void load()
    void refreshSidebar()
  }

  async function load(): Promise<void> {
    loading = true
    try {
      accounts = await listAccounts()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      loading = false
    }
  }

  function startEdit(account: Account): void {
    confirmingId = null
    editingId = account.id
    draft = { ...account }
  }

  function cancelEdit(): void {
    editingId = null
    draft = null
  }

  async function save(): Promise<void> {
    if (!draft) {
      return
    }
    saving = true
    try {
      const updated = await updateAccount({
        id: draft.id,
        displayName: draft.displayName,
        username: draft.username,
        imapHost: draft.imapHost,
        imapPort: draft.imapPort,
        smtpHost: draft.smtpHost,
        smtpPort: draft.smtpPort,
      })
      accounts = accounts.map((a) => (a.id === updated.id ? updated : a))
      void refreshSidebar()
      cancelEdit()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      saving = false
    }
  }

  async function confirmDelete(id: number): Promise<void> {
    try {
      await deleteAccount(id)
      accounts = accounts.filter((a) => a.id !== id)
      confirmingId = null
      void refreshSidebar()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }
</script>

<div class="head">
  <div>
    <h3>{$t('settingsPanel.category.mailboxes')}</h3>
    <p class="hint">{$t('mailboxes.hint')}</p>
  </div>
  <button type="button" class="add-btn" on:click={() => (wizardOpen = true)}>
    <IconPlus size={14} stroke={2} />
    {$t('mailboxes.add')}
  </button>
</div>

{#if loading}
  <p class="empty">{$t('mailboxes.loading')}</p>
{:else if accounts.length === 0}
  <p class="empty">{$t('mailboxes.empty')}</p>
{:else}
  <ul class="list">
    {#each accounts as account (account.id)}
      <li>
        {#if editingId === account.id && draft}
          <div class="edit">
            <span class="email">{account.email}</span>
            <label class="field">
              <span>{$t('wizard.field.displayName')}</span>
              <input type="text" bind:value={draft.displayName} />
            </label>
            <label class="field">
              <span>{$t('wizard.field.username')}</span>
              <input type="text" bind:value={draft.username} placeholder={account.email} />
            </label>
            <div class="servers">
              <label class="field"><span>{$t('wizard.field.imapHost')}</span><input type="text" bind:value={draft.imapHost} /></label>
              <label class="field narrow"><span>{$t('wizard.field.port')}</span><input type="number" bind:value={draft.imapPort} /></label>
            </div>
            <div class="servers">
              <label class="field"><span>{$t('wizard.field.smtpHost')}</span><input type="text" bind:value={draft.smtpHost} /></label>
              <label class="field narrow"><span>{$t('wizard.field.port')}</span><input type="number" bind:value={draft.smtpPort} /></label>
            </div>
            <p class="server-hint">{$t('mailboxes.serverChangeHint')}</p>
            <div class="edit-actions">
              <button type="button" class="ghost" on:click={cancelEdit}>{$t('mailboxes.cancel')}</button>
              <button type="button" class="primary" disabled={saving} on:click={save}>
                <IconCheck size={14} stroke={2} />
                {saving ? $t('mailboxes.saving') : $t('mailboxes.save')}
              </button>
            </div>
          </div>
        {:else}
          <div class="who">
            <span class="name">{account.displayName || account.email}</span>
            {#if account.displayName}<span class="addr">{account.email}</span>{/if}
          </div>
          {#if confirmingId === account.id}
            <div class="confirm">
              <span class="warn">{$t('mailboxes.deleteConfirm')}</span>
              <button type="button" class="danger" on:click={() => confirmDelete(account.id)}>{$t('action.delete')}</button>
              <button type="button" class="ghost" on:click={() => (confirmingId = null)}>{$t('mailboxes.cancel')}</button>
            </div>
          {:else}
            <button type="button" class="icon" aria-label={`${$t('mailboxes.edit')} ${account.email}`} on:click={() => startEdit(account)}>
              <IconPencil size={15} stroke={1.7} />
            </button>
            <button type="button" class="icon del" aria-label={`${$t('action.delete')} ${account.email}`} on:click={() => ((confirmingId = account.id), (editingId = null))}>
              <IconTrash size={15} stroke={1.7} />
            </button>
          {/if}
        {/if}
      </li>
    {/each}
  </ul>
{/if}

{#if wizardOpen}
  {#await import('../wizard/AddMailboxWizard.svelte') then m}
    <svelte:component this={m.default} on:close={() => (wizardOpen = false)} on:added={onMailboxAdded} />
  {/await}
{/if}

<style>
  .head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .add-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    flex-shrink: 0;
    padding: var(--space-2) var(--space-4);
    border: none;
    border-radius: var(--radius-control);
    background: var(--accent);
    color: var(--accent-fg);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    cursor: pointer;
  }
  .add-btn:hover {
    filter: brightness(1.05);
  }

  h3 {
    margin: 0 0 var(--space-3);
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .hint {
    margin: 0 0 var(--space-4);
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .empty {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    padding: var(--space-3) 0;
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
  }

  li {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-1);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .who {
    display: flex;
    flex-direction: column;
    min-width: 0;
    flex: 1;
  }
  .name {
    font-size: var(--fz-label);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .addr {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .icon {
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: var(--space-1);
    border-radius: var(--radius-control);
    flex-shrink: 0;
  }
  .icon:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }
  .icon.del:hover {
    color: var(--danger);
  }

  .confirm {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
    justify-content: flex-end;
  }
  .warn {
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .edit {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    width: 100%;
  }
  .email {
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .field span {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }
  .field input {
    height: var(--control-height);
    padding: 0 var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-primary);
    font-size: var(--fz-list);
  }
  .field input:focus {
    border-color: var(--accent);
    outline: none;
  }

  .servers {
    display: flex;
    gap: var(--space-2);
  }
  .servers .field {
    flex: 1;
  }
  .servers .field.narrow {
    flex: 0 0 88px;
  }

  .server-hint {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .edit-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-top: var(--space-1);
  }

  .primary,
  .ghost,
  .danger {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4);
    border-radius: var(--radius-control);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    cursor: pointer;
    border: var(--hairline) solid var(--border-default);
  }
  .primary {
    background: var(--accent);
    color: var(--accent-fg);
    border-color: transparent;
  }
  .primary:disabled {
    opacity: 0.6;
    cursor: default;
  }
  .ghost {
    background: transparent;
    color: var(--text-secondary);
  }
  .ghost:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }
  .danger {
    background: var(--danger);
    color: var(--accent-fg);
    border-color: transparent;
  }
</style>
