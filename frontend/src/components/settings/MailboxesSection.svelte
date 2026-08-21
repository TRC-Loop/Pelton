<script lang="ts">
  // the mailbox manager: lists configured accounts and lets the user edit an
  // account's display name and server settings, or delete it outright. Deleting
  // is destructive (it drops the cached mail and the keyring secret), so it goes
  // through an inline confirm. Email is not editable here, it identifies the
  // account; changing it is a re-add.
  import { onMount } from 'svelte'
  import { IconPencil, IconTrash, IconCheck, IconPlus, IconAlertTriangle } from '@tabler/icons-svelte'
  import ToggleSwitch from '../common/ToggleSwitch.svelte'
  import Modal from '../common/Modal.svelte'
  import {
    listAccounts,
    updateAccount,
    deleteAccount,
    getLogStatus,
    deleteLogs,
    chooseArchiveExportFolder,
    previewArchiveExportName,
  } from '../../lib/api'
  import { refreshSidebar } from '../../stores/accounts'
  import { missingPassword, askForPassword, refreshMissingPasswords } from '../../stores/passwordprompt'
  import { errorMessage, toastError, pushAction } from '../../stores/toast'
  import type { Account, TLSMode } from '../../lib/types'
  import { t } from '../../lib/i18n'

  let accounts: Account[] = []
  let loading = true
  let editingId: number | null = null
  let confirmingId: number | null = null
  let saving = false
  // the working copy of the account being edited, so cancelling discards edits.
  let draft: Account | null = null
  // the password is not part of the account row (it lives in the keyring and is
  // never sent back), so it is drafted separately. Empty means "leave it".
  let passwordDraft = ''
  // the add-mailbox wizard is code-split like the other settings modals, so
  // it only loads once the user actually asks to add a mailbox.
  let wizardOpen = false
  // the sample file name the export settings show, rendered by the backend so
  // the ui never has a second copy of the naming rules.
  let namePreview = ''
  // the draft as it opened, so closing knows whether there is anything to lose.
  let opened = ''

  $: dirty = draft !== null && (JSON.stringify(draft) !== opened || passwordDraft !== '')

  // the subfolder modes, in the order the picker shows them.
  const subfolderModes = ['none', 'year', 'month'] as const

  // how this mailbox starts a new message: unprotected, signed, or signed and
  // encrypted whenever every recipient has a key.
  const pgpDefaults = ['', 'sign', 'auto'] as const

  function setPGPDefault(value: string): void {
    if (draft) {
      draft.pgpDefault = value
    }
  }

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
      await refreshMissingPasswords()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      loading = false
    }
  }

  function startEdit(account: Account): void {
    confirmingId = null
    editingId = account.id
    // the backend already reports the security an account actually connects
    // with, resolving the empty value an older account carries, so the control
    // shows the truth and saving pins it.
    draft = { ...account }
    opened = JSON.stringify(draft)
    passwordDraft = ''
    void refreshPreview()
  }

  // refreshPreview renders the current template through the backend. A failure
  // only costs the preview line, so it falls back to empty rather than a toast.
  async function refreshPreview(): Promise<void> {
    if (!draft) {
      return
    }
    try {
      namePreview = await previewArchiveExportName(draft.exportNameTemplate, draft.exportSubfolders)
    } catch {
      namePreview = ''
    }
  }

  function setExportOnArchive(on: boolean): void {
    if (draft) {
      draft.exportOnArchive = on
    }
  }

  function setSubfolders(mode: string): void {
    if (draft) {
      draft.exportSubfolders = mode
      void refreshPreview()
    }
  }

  // pickExportFolder opens the native directory picker. Choosing a folder turns
  // the export on, since that is plainly what the click meant.
  async function pickExportFolder(): Promise<void> {
    if (!draft) {
      return
    }
    try {
      const dir = await chooseArchiveExportFolder()
      if (dir && draft) {
        draft.exportDir = dir
        draft.exportOnArchive = true
      }
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // the security setters take the null check out of the template: inside an
  // event handler the draft's narrowing from the enclosing block is lost.
  function setIMAPTLS(mode: TLSMode): void {
    if (draft) {
      draft.imapTls = mode
    }
  }

  function setSMTPTLS(mode: TLSMode): void {
    if (draft) {
      draft.smtpTls = mode
    }
  }

  function cancelEdit(): void {
    editingId = null
    draft = null
    passwordDraft = ''
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
        password: passwordDraft,
        imapTls: draft.imapTls as TLSMode,
        smtpTls: draft.smtpTls as TLSMode,
        exportOnArchive: draft.exportOnArchive,
        exportDir: draft.exportDir,
        exportSubfolders: draft.exportSubfolders,
        exportNameTemplate: draft.exportNameTemplate,
        pgpDefault: draft.pgpDefault,
      })
      accounts = accounts.map((a) => (a.id === updated.id ? updated : a))
      if (passwordDraft !== '') {
        void refreshMissingPasswords()
      }
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
      void offerLogCleanup()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // a log written while a mailbox existed is still a record of using it, so
  // deleting the mailbox offers to take the logs with it (#211). Nothing is
  // removed without the click: the logs may be the reason logging was on.
  async function offerLogCleanup(): Promise<void> {
    let status
    try {
      status = await getLogStatus()
    } catch {
      return
    }
    if (status.sizeBytes === 0) {
      return
    }
    pushAction(
      'info',
      $t('mailboxes.deleteLogsOffer'),
      {
        label: $t('settingsPanel.button.deleteLogs'),
        run: () => {
          deleteLogs().catch((err) => toastError(errorMessage(err)))
        },
      },
      8000,
    )
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
        <div class="who">
          <span class="name">{account.displayName || account.email}</span>
          {#if account.displayName}<span class="addr">{account.email}</span>{/if}
        </div>
        {#if $missingPassword.has(account.id)}
          <button
            type="button"
            class="icon warn-icon"
            title={$t('mailboxes.passwordPrompt.marker')}
            aria-label={$t('mailboxes.passwordPrompt.marker')}
            on:click={() => askForPassword(account)}
          >
            <IconAlertTriangle size={15} stroke={1.8} />
          </button>
        {/if}
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
      </li>
    {/each}
  </ul>
{/if}

{#if draft}
  <Modal title={draft.email} hint={$t('mailboxes.serverChangeHint')} size="large" {dirty} on:close={cancelEdit}>
    <div class="form">
      <label class="field">
        <span>{$t('wizard.field.displayName')}</span>
        <input type="text" bind:value={draft.displayName} />
      </label>
      <label class="field">
        <span>{$t('wizard.field.username')}</span>
        <input type="text" bind:value={draft.username} placeholder={draft.email} />
      </label>
      <div class="servers">
        <label class="field"><span>{$t('wizard.field.imapHost')}</span><input type="text" bind:value={draft.imapHost} /></label>
        <label class="field narrow"><span>{$t('wizard.field.port')}</span><input type="number" bind:value={draft.imapPort} /></label>
      </div>
      <div class="servers">
        <span class="field">
          <span>{$t('wizard.advanced.imapSecurity')}</span>
          <div class="seg" role="radiogroup" aria-label={$t('wizard.advanced.imapSecurity')}>
            <button type="button" class:on={draft.imapTls === 'ssl'} on:click={() => setIMAPTLS('ssl')}>SSL / TLS</button>
            <button type="button" class:on={draft.imapTls === 'starttls'} on:click={() => setIMAPTLS('starttls')}>STARTTLS</button>
          </div>
        </span>
      </div>
      <div class="servers">
        <label class="field"><span>{$t('wizard.field.smtpHost')}</span><input type="text" bind:value={draft.smtpHost} /></label>
        <label class="field narrow"><span>{$t('wizard.field.port')}</span><input type="number" bind:value={draft.smtpPort} /></label>
      </div>
      <div class="servers">
        <span class="field">
          <span>{$t('wizard.advanced.smtpSecurity')}</span>
          <div class="seg" role="radiogroup" aria-label={$t('wizard.advanced.smtpSecurity')}>
            <button type="button" class:on={draft.smtpTls === 'ssl'} on:click={() => setSMTPTLS('ssl')}>SSL / TLS</button>
            <button type="button" class:on={draft.smtpTls === 'starttls'} on:click={() => setSMTPTLS('starttls')}>STARTTLS</button>
          </div>
        </span>
      </div>
      <label class="field">
        <span>{$t('wizard.field.password')}</span>
        <input
          type="password"
          bind:value={passwordDraft}
          autocomplete="off"
          placeholder={$t(
            $missingPassword.has(draft.id)
              ? 'mailboxes.passwordMissing'
              : 'mailboxes.passwordUnchanged',
          )}
        />
      </label>
      {#if $missingPassword.has(draft.id)}
        <p class="server-hint warn">{$t('mailboxes.passwordNeededHint')}</p>
      {/if}

      <div class="toggle">
        <span>{$t('mailboxes.export.toggle')}</span>
        <ToggleSwitch
          checked={draft.exportOnArchive}
          disabled={!draft.exportDir}
          label={$t('mailboxes.export.toggle')}
          on:change={(e) => setExportOnArchive(e.detail)}
        />
      </div>
      <p class="server-hint">{$t('mailboxes.export.hint')}</p>
      <div class="folder-row">
        <span class="path" class:unset={!draft.exportDir}>
          {draft.exportDir || $t('mailboxes.export.noFolder')}
        </span>
        <button type="button" class="ghost" on:click={pickExportFolder}>
          {$t('mailboxes.export.choose')}
        </button>
      </div>
      <span class="field">
        <span>{$t('mailboxes.export.subfolders')}</span>
        <div class="seg" role="radiogroup" aria-label={$t('mailboxes.export.subfolders')}>
          {#each subfolderModes as mode (mode)}
            <button type="button" class:on={draft.exportSubfolders === mode} on:click={() => setSubfolders(mode)}>
              {$t(`mailboxes.export.subfolders.${mode}`)}
            </button>
          {/each}
        </div>
      </span>
      <label class="field">
        <span>{$t('mailboxes.export.template')}</span>
        <input
          type="text"
          bind:value={draft.exportNameTemplate}
          on:input={refreshPreview}
          placeholder="{'{date}'}_{'{subject}'}"
        />
      </label>
      <p class="server-hint">{$t('mailboxes.export.placeholders')}</p>
      <span class="field">
        <span>{$t('mailboxes.pgpDefault')}</span>
        <div class="seg" role="radiogroup" aria-label={$t('mailboxes.pgpDefault')}>
          {#each pgpDefaults as mode (mode)}
            <button type="button" class:on={draft.pgpDefault === mode} on:click={() => setPGPDefault(mode)}>
              {$t(`mailboxes.pgpDefault.${mode || 'off'}`)}
            </button>
          {/each}
        </div>
      </span>
      <p class="server-hint">{$t('mailboxes.pgpDefaultHint')}</p>

      {#if namePreview}
        <p class="server-hint preview">{$t('mailboxes.export.preview')} <code>{namePreview}</code></p>
      {/if}
    </div>

    <svelte:fragment slot="footer">
      <button type="button" class="ghost" on:click={cancelEdit}>{$t('mailboxes.cancel')}</button>
      <button type="button" class="primary" disabled={saving} on:click={save}>
        <IconCheck size={14} stroke={2} />
        {saving ? $t('mailboxes.saving') : $t('mailboxes.save')}
      </button>
    </svelte:fragment>
  </Modal>
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
    cursor: var(--cursor-action);
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
    cursor: var(--cursor-action);
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

  /* the mailbox cannot sync, so this one is coloured at rest rather than on
     hover like the edit and delete buttons next to it. */
  .warn-icon {
    color: var(--warning);
  }
  .warn-icon:hover {
    color: var(--warning);
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

  /* the security picker, matching the one in the add-mailbox wizard. */
  .seg {
    display: inline-flex;
    align-self: flex-start;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    overflow: hidden;
  }
  .seg button {
    border: none;
    background: var(--surface-raised);
    color: var(--text-secondary);
    cursor: pointer;
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
  }
  .seg button + button {
    border-left: var(--hairline) solid var(--border-default);
  }
  .seg button.on {
    background: var(--accent);
    color: var(--accent-fg);
  }
  .servers .field.narrow {
    flex: 0 0 88px;
  }

  .server-hint {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  /* the export-on-archive block: a switch, the chosen folder, and a preview of
     the file name the template produces. */
  .toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    font-size: var(--fz-label);
    color: var(--text-primary);
  }

  .folder-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .path {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    direction: rtl;
    text-align: left;
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .path.unset {
    color: var(--text-tertiary);
    direction: ltr;
  }

  .preview code {
    font-family: var(--font-mono);
    color: var(--text-secondary);
    overflow-wrap: anywhere;
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
    cursor: var(--cursor-action);
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
