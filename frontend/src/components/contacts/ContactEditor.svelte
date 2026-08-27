<script lang="ts">
  // the contact editor (#168). It edits the fields Pelton shows; everything
  // else on the vCard is kept untouched by the save, which is what the line at
  // the bottom says.
  import { createEventDispatcher } from 'svelte'
  import { IconPlus, IconMinus } from '@tabler/icons-svelte'
  import Modal from '../common/Modal.svelte'
  import { addressBooks, saveContact } from '../../stores/contacts'
  import { t } from '../../lib/i18n'
  import type { ContactDraft } from '../../lib/types'

  /** The contact being edited. id 0 creates a new one. */
  export let draft: ContactDraft

  const dispatch = createEventDispatcher<{ close: void }>()

  let saving = false
  const opened = JSON.stringify(draft)

  $: dirty = JSON.stringify(draft) !== opened
  // no default address book: a contact filed in the wrong one is a contact on
  // the wrong device, so a new contact cannot be saved until a book is chosen.
  $: writable = $addressBooks.filter((b) => !b.readOnly)
  $: canSave =
    !saving &&
    draft.bookId !== 0 &&
    (draft.fullName.trim() !== '' || draft.emails.some((e) => e.value.trim() !== ''))

  function addEmail(): void {
    draft.emails = [...draft.emails, { value: '', label: '' }]
  }

  function removeEmail(index: number): void {
    draft.emails = draft.emails.filter((_, i) => i !== index)
  }

  function addPhone(): void {
    draft.phones = [...draft.phones, { value: '', label: '' }]
  }

  function removePhone(index: number): void {
    draft.phones = draft.phones.filter((_, i) => i !== index)
  }

  async function save(): Promise<void> {
    saving = true
    const saved = await saveContact(draft)
    saving = false
    if (saved) {
      dispatch('close')
    }
    // a conflict leaves the editor open behind the conflict dialog, so
    // whichever way it is resolved the typed version is still here.
  }
</script>

<Modal
  title={draft.id === 0 ? $t('contacts.editor.new') : $t('contacts.editor.edit')}
  size="medium"
  {dirty}
  on:close={() => dispatch('close')}
>
  <div class="form">
    <label class="field">
      <span>{$t('contacts.field.book')}</span>
      <select bind:value={draft.bookId} disabled={draft.id !== 0}>
        <option value={0} disabled>{$t('contacts.field.bookPlaceholder')}</option>
        {#each writable as book (book.id)}
          <option value={book.id}>{book.name || $t('contacts.book.untitled')}</option>
        {/each}
      </select>
    </label>

    <label class="field">
      <span>{$t('contacts.field.name')}</span>
      <input type="text" bind:value={draft.fullName} />
    </label>

    <div class="pair">
      <label class="field">
        <span>{$t('contacts.field.organization')}</span>
        <input type="text" bind:value={draft.organization} />
      </label>
      <label class="field">
        <span>{$t('contacts.field.title')}</span>
        <input type="text" bind:value={draft.title} />
      </label>
    </div>

    <span class="group-label">{$t('contacts.field.emails')}</span>
    {#each draft.emails as email, i (i)}
      <div class="value-row">
        <input type="email" bind:value={email.value} placeholder="name@example.com" />
        <input type="text" bind:value={email.label} placeholder={$t('contacts.field.labelPlaceholder')} class="label-input" />
        <button type="button" class="icon" aria-label={$t('contacts.field.removeEmail')} on:click={() => removeEmail(i)}>
          <IconMinus size={14} stroke={1.8} />
        </button>
      </div>
    {/each}
    <button type="button" class="add" on:click={addEmail}>
      <IconPlus size={13} stroke={2} />
      {$t('contacts.field.addEmail')}
    </button>

    <span class="group-label">{$t('contacts.field.phones')}</span>
    {#each draft.phones as phone, i (i)}
      <div class="value-row">
        <input type="tel" bind:value={phone.value} />
        <input type="text" bind:value={phone.label} placeholder={$t('contacts.field.labelPlaceholder')} class="label-input" />
        <button type="button" class="icon" aria-label={$t('contacts.field.removePhone')} on:click={() => removePhone(i)}>
          <IconMinus size={14} stroke={1.8} />
        </button>
      </div>
    {/each}
    <button type="button" class="add" on:click={addPhone}>
      <IconPlus size={13} stroke={2} />
      {$t('contacts.field.addPhone')}
    </button>

    <label class="field">
      <span>{$t('contacts.field.note')}</span>
      <textarea rows="3" bind:value={draft.note}></textarea>
    </label>

    <p class="hint">{$t('contacts.editor.keepsRest')}</p>

    <div class="actions">
      <button type="button" class="ghost" on:click={() => dispatch('close')}>{$t('mailboxes.cancel')}</button>
      <button type="button" class="primary" disabled={!canSave} on:click={save}>
        {saving ? $t('contacts.editor.saving') : $t('action.save')}
      </button>
    </div>
  </div>
</Modal>

<style>
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

  .group-label {
    margin-top: var(--space-2);
  }

  .pair {
    display: flex;
    gap: var(--space-3);
  }

  .pair .field {
    flex: 1;
  }

  input,
  select,
  textarea {
    padding: 0 var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font: inherit;
    font-size: var(--fz-body);
  }

  input,
  select {
    height: var(--control-height);
  }

  textarea {
    padding: var(--space-2);
    resize: vertical;
    font-family: var(--font-ui);
  }

  .value-row {
    display: flex;
    gap: var(--space-2);
  }

  .value-row input {
    flex: 1;
  }

  .value-row .label-input {
    flex: 0 0 120px;
  }

  .add {
    align-self: flex-start;
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    border: none;
    background: transparent;
    color: var(--accent);
    font: inherit;
    font-size: var(--fz-meta);
    padding: 0;
    cursor: var(--cursor-action);
  }

  .hint {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-top: var(--space-2);
  }

  button.primary,
  button.ghost {
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

  button.primary {
    background: var(--accent);
    color: var(--accent-fg);
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
    color: var(--danger);
  }
</style>
