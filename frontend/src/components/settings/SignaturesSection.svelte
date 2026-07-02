<script lang="ts">
  // the Signatures settings: write reusable header/footer blocks (markdown or
  // html), then assign a default header and/or footer per mailbox. the assigned
  // defaults are inserted when a new compose opens; the user can still change
  // them per message from the compose footer.
  import { onMount } from 'svelte'
  import { IconPlus, IconPencil, IconTrash } from '@tabler/icons-svelte'
  import { signatures, persistSignature, removeSignature, getAccountSignatures, setAccountSignatures } from '../../stores/signatures'
  import { sidebar } from '../../stores/accounts'
  import { errorMessage, toastError } from '../../stores/toast'
  import type { Signature, AccountSignatures } from '../../lib/types'
  import { get } from 'svelte/store'
  import { t } from '../../lib/i18n'

  $: accounts = $sidebar.data?.accounts ?? []
  $: headers = $signatures.filter((s) => s.kind === 'header')
  $: footers = $signatures.filter((s) => s.kind === 'footer')

  // the editor draft. id 0 means a new block.
  const blank = (): Signature => ({ id: 0, name: '', kind: 'footer', format: 'markdown', content: '' })
  let draft: Signature = blank()
  let editing = false

  function startNew(): void {
    draft = blank()
    editing = true
  }

  function startEdit(s: Signature): void {
    draft = { ...s }
    editing = true
  }

  async function save(): Promise<void> {
    if (!draft.name.trim()) {
      toastError(get(t)('signatures.needName'))
      return
    }
    try {
      await persistSignature(draft)
      editing = false
      draft = blank()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function remove(id: number): Promise<void> {
    try {
      await removeSignature(id)
      // refresh assignments since a deleted block clears any default using it.
      await loadAssignments()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // per-account default assignment, loaded on mount and when accounts change.
  let assignments: Record<number, AccountSignatures> = {}

  async function loadAssignments(): Promise<void> {
    const next: Record<number, AccountSignatures> = {}
    for (const acc of accounts) {
      try {
        next[acc.id] = await getAccountSignatures(acc.id)
      } catch {
        next[acc.id] = { headerId: 0, footerId: 0 }
      }
    }
    assignments = next
  }

  onMount(loadAssignments)
  // reload assignments once accounts arrive (sidebar may still be loading at mount).
  $: if (accounts.length && Object.keys(assignments).length === 0) {
    void loadAssignments()
  }

  async function setAssignment(accountId: number, kind: 'header' | 'footer', id: number): Promise<void> {
    const current = assignments[accountId] ?? { headerId: 0, footerId: 0 }
    const next: AccountSignatures = {
      headerId: kind === 'header' ? id : current.headerId,
      footerId: kind === 'footer' ? id : current.footerId,
    }
    assignments = { ...assignments, [accountId]: next }
    try {
      await setAccountSignatures(accountId, next.headerId, next.footerId)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }
</script>

<h3>{$t('signatures.title')}</h3>
<p class="hint">{$t('signatures.hint')}</p>

<div class="blocks">
  {#if $signatures.length === 0}
    <p class="empty">{$t('signatures.empty')}</p>
  {:else}
    {#each $signatures as s (s.id)}
      <div class="block">
        <div class="block-main">
          <span class="kind {s.kind}">{s.kind}</span>
          <span class="block-name">{s.name}</span>
          <span class="fmt">{s.format}</span>
        </div>
        <div class="block-actions">
          <button type="button" class="icon-btn" aria-label={$t('signatures.edit')} on:click={() => startEdit(s)}>
            <IconPencil size={15} stroke={1.6} />
          </button>
          <button type="button" class="icon-btn danger" aria-label={$t('signatures.delete')} on:click={() => remove(s.id)}>
            <IconTrash size={15} stroke={1.6} />
          </button>
        </div>
      </div>
    {/each}
  {/if}
</div>

{#if editing}
  <div class="editor">
    <div class="editor-row">
      <input class="field" placeholder={$t('signatures.namePlaceholder')} bind:value={draft.name} />
      <select class="field" bind:value={draft.kind}>
        <option value="header">{$t('signatures.kindHeader')}</option>
        <option value="footer">{$t('signatures.kindFooter')}</option>
      </select>
      <select class="field" bind:value={draft.format}>
        <option value="markdown">Markdown</option>
        <option value="html">HTML</option>
      </select>
    </div>
    <textarea class="content" rows="6" placeholder={$t('signatures.contentPlaceholder')} bind:value={draft.content}></textarea>
    <div class="editor-actions">
      <button type="button" class="ghost" on:click={() => (editing = false)}>{$t('signatures.cancel')}</button>
      <button type="button" class="primary" on:click={save}>{$t('signatures.save')}</button>
    </div>
  </div>
{:else}
  <button type="button" class="add" on:click={startNew}>
    <IconPlus size={15} stroke={1.8} />
    {$t('signatures.new')}
  </button>
{/if}

{#if accounts.length > 0}
  <h4>{$t('signatures.perMailboxDefaults')}</h4>
  <div class="assign">
    {#each accounts as acc (acc.id)}
      <div class="assign-row">
        <span class="assign-acc" title={acc.email}>{acc.email}</span>
        <label class="assign-field">
          <span>{$t('signatures.kindHeader')}</span>
          <select
            value={assignments[acc.id]?.headerId ?? 0}
            on:change={(e) => setAssignment(acc.id, 'header', Number(e.currentTarget.value))}
          >
            <option value={0}>{$t('signatures.none')}</option>
            {#each headers as s (s.id)}<option value={s.id}>{s.name}</option>{/each}
          </select>
        </label>
        <label class="assign-field">
          <span>{$t('signatures.kindFooter')}</span>
          <select
            value={assignments[acc.id]?.footerId ?? 0}
            on:change={(e) => setAssignment(acc.id, 'footer', Number(e.currentTarget.value))}
          >
            <option value={0}>{$t('signatures.none')}</option>
            {#each footers as s (s.id)}<option value={s.id}>{s.name}</option>{/each}
          </select>
        </label>
      </div>
    {/each}
  </div>
{/if}

<style>
  h3 {
    margin: 0 0 var(--space-3);
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  h4 {
    margin: var(--space-5) 0 var(--space-3);
    font-size: var(--fz-body);
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
    margin: 0 0 var(--space-3);
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }

  .blocks {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .block {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
  }

  .block-main {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    min-width: 0;
  }

  .kind {
    font-size: var(--fz-meta);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 1px var(--space-2);
    border-radius: 999px;
    background: var(--surface-sunken);
    color: var(--text-tertiary);
  }

  .kind.header {
    color: var(--accent);
  }

  .block-name {
    font-size: var(--fz-body);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .fmt {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .block-actions {
    display: inline-flex;
    gap: var(--space-1);
    flex-shrink: 0;
  }

  .icon-btn {
    display: inline-flex;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: var(--space-1);
    border-radius: var(--radius-control);
  }

  .icon-btn:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .icon-btn.danger:hover {
    color: var(--danger);
  }

  .add {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    margin-top: var(--space-3);
    padding: var(--space-2) var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .add:hover {
    background: var(--surface-hover);
  }

  .editor {
    margin-top: var(--space-3);
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-raised);
  }

  .editor-row {
    display: flex;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }

  .field {
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-base);
    color: var(--text-primary);
    padding: var(--space-2);
    font-size: var(--fz-label);
  }

  .editor-row .field:first-child {
    flex: 1;
  }

  .content {
    width: 100%;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-base);
    color: var(--text-primary);
    padding: var(--space-2);
    font-family: var(--font-mono);
    font-size: var(--fz-label);
    resize: vertical;
  }

  .editor-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-top: var(--space-2);
  }

  .ghost,
  .primary {
    padding: var(--space-2) var(--space-4);
    border-radius: var(--radius-control);
    border: var(--hairline) solid var(--border-default);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .ghost {
    background: var(--surface-raised);
    color: var(--text-primary);
  }

  .primary {
    background: var(--accent);
    color: var(--accent-fg, #fff);
    border-color: transparent;
  }

  .assign {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .assign-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex-wrap: wrap;
  }

  .assign-acc {
    flex: 1;
    min-width: 140px;
    font-size: var(--fz-label);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .assign-field {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .assign-field select {
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-base);
    color: var(--text-primary);
    padding: var(--space-1) var(--space-2);
    font-size: var(--fz-label);
  }
</style>
