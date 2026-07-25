<script lang="ts">
  // the saved View builder: name, icon, color, the search query (text + from/to/
  // subject), a relative date window, scope toggles and an account scope. used
  // both for creating a view from scratch and for editing an existing one (and
  // for "save this search as a view", where the parent seeds the query fields).
  import { createEventDispatcher } from 'svelte'
  import { IconX } from '@tabler/icons-svelte'
  import type { View, Account } from '../../lib/types'
  import { viewIconNames, viewIconComponent, viewColors, viewColorCss } from '../../lib/viewicons'
  import { saveView } from '../../lib/api'
  import { upsertView } from '../../stores/views'
  import { toastError, errorMessage } from '../../stores/toast'
  import { t } from '../../lib/i18n'

  // value is the working view. the parent passes a blank view (id 0) to create,
  // or an existing view to edit; either way it is copied so edits stay local
  // until save.
  export let value: View
  export let accounts: Account[] = []

  const dispatch = createEventDispatcher<{ saved: View; close: void }>()

  // local copy so cancelling discards edits.
  let draft: View = { ...value }
  let saving = false

  // relative date window presets, in days. 0 means no bound.
  const windows: { days: number; key: string }[] = [
    { days: 0, key: 'views.window.any' },
    { days: 1, key: 'views.window.day' },
    { days: 7, key: 'views.window.week' },
    { days: 30, key: 'views.window.month' },
    { days: 90, key: 'views.window.quarter' },
    { days: 365, key: 'views.window.year' },
  ]

  $: canSave = draft.name.trim() !== '' && !saving

  async function save(): Promise<void> {
    if (!canSave) {
      return
    }
    saving = true
    try {
      const saved = await saveView(draft)
      upsertView(saved)
      dispatch('saved', saved)
      dispatch('close')
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      saving = false
    }
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions a11y-no-noninteractive-element-interactions -->
<div class="overlay" on:click={() => dispatch('close')}>
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" aria-modal="true" aria-label={$t('views.editorTitle')} tabindex="-1" on:click|stopPropagation>
    <header>
      <span class="m-title">{draft.id === 0 ? $t('views.newTitle') : $t('views.editTitle')}</span>
      <button type="button" class="m-close" aria-label={$t('about.close')} on:click={() => dispatch('close')}>
        <IconX size={18} stroke={1.8} />
      </button>
    </header>

    <div class="m-body">
      <label class="field">
        <span class="label">{$t('views.nameLabel')}</span>
        <input type="text" bind:value={draft.name} placeholder={$t('views.namePlaceholder')} maxlength="60" />
      </label>

      <div class="field">
        <span class="label">{$t('views.iconLabel')}</span>
        <div class="icon-grid">
          {#each viewIconNames as name (name)}
            <button
              type="button"
              class="icon-cell"
              class:on={draft.icon === name || (draft.icon === '' && name === 'bookmark')}
              title={name}
              style={`color:${viewColorCss(draft.color)}`}
              on:click={() => (draft.icon = name)}
            >
              <svelte:component this={viewIconComponent(name)} size={18} stroke={1.7} />
            </button>
          {/each}
        </div>
      </div>

      <div class="field">
        <span class="label">{$t('views.colorLabel')}</span>
        <div class="swatches">
          {#each viewColors as color (color.value)}
            <button
              type="button"
              class="swatch"
              class:on={draft.color === color.value}
              style={`--sw:${color.css}`}
              aria-label={color.value || $t('views.colorNone')}
              on:click={() => (draft.color = color.value)}
            ></button>
          {/each}
        </div>
      </div>

      <div class="section-head">{$t('views.filterHeading')}</div>

      <label class="field">
        <span class="label">{$t('views.queryText')}</span>
        <input type="text" bind:value={draft.queryText} placeholder={$t('views.queryTextPlaceholder')} />
      </label>

      <div class="two">
        <label class="field">
          <span class="label">{$t('views.queryFrom')}</span>
          <input type="text" bind:value={draft.queryFrom} placeholder={$t('views.queryFromPlaceholder')} />
        </label>
        <label class="field">
          <span class="label">{$t('views.queryTo')}</span>
          <input type="text" bind:value={draft.queryTo} placeholder={$t('views.queryToPlaceholder')} />
        </label>
      </div>

      <label class="field">
        <span class="label">{$t('views.querySubject')}</span>
        <input type="text" bind:value={draft.querySubject} placeholder={$t('views.querySubjectPlaceholder')} />
      </label>

      <div class="two">
        <label class="field">
          <span class="label">{$t('views.dateWindow')}</span>
          <select bind:value={draft.withinDays}>
            {#each windows as w (w.days)}
              <option value={w.days}>{$t(w.key)}</option>
            {/each}
          </select>
        </label>
        <label class="field">
          <span class="label">{$t('views.accountScope')}</span>
          <select bind:value={draft.accountId}>
            <option value={0}>{$t('views.allAccounts')}</option>
            {#each accounts as acc (acc.id)}
              <option value={acc.id}>{acc.email}</option>
            {/each}
          </select>
        </label>
      </div>

      <div class="toggles">
        <label class="check">
          <input type="checkbox" bind:checked={draft.unreadOnly} />
          <span>{$t('views.unreadOnly')}</span>
        </label>
        <label class="check">
          <input type="checkbox" bind:checked={draft.flaggedOnly} />
          <span>{$t('views.flaggedOnly')}</span>
        </label>
        <label class="check">
          <input type="checkbox" bind:checked={draft.hasAttachment} />
          <span>{$t('views.hasAttachment')}</span>
        </label>
      </div>
    </div>

    <footer>
      <button type="button" class="btn ghost" on:click={() => dispatch('close')}>{$t('views.cancel')}</button>
      <button type="button" class="btn primary" disabled={!canSave} on:click={save}>{$t('views.save')}</button>
    </footer>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: var(--overlay-scrim, rgba(0, 0, 0, 0.4));
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 60;
    padding: var(--space-4);
  }

  .modal {
    display: flex;
    flex-direction: column;
    width: min(560px, 100%);
    max-height: min(88vh, 720px);
    background: var(--surface-base);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    box-shadow: var(--shadow-overlay, 0 20px 60px rgba(0, 0, 0, 0.3));
    overflow: hidden;
  }

  header,
  footer {
    display: flex;
    align-items: center;
    padding: var(--space-3) var(--space-4);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  footer {
    border-bottom: none;
    border-top: var(--hairline) solid var(--border-subtle);
    justify-content: flex-end;
    gap: var(--space-2);
  }

  .m-title {
    flex: 1;
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .m-close {
    display: inline-flex;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    border-radius: var(--radius-control);
    padding: var(--space-1);
  }

  .m-close:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .m-body {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-4);
    overflow-y: auto;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .two {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
  }

  .label {
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }

  input[type='text'],
  select {
    width: 100%;
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-primary);
    font-size: var(--fz-label);
    outline: none;
  }

  input[type='text']:focus,
  select:focus {
    border-color: var(--accent);
  }

  .section-head {
    margin-top: var(--space-1);
    font-size: var(--fz-meta);
    font-weight: var(--fw-semibold);
    color: var(--text-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .icon-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(34px, 1fr));
    gap: var(--space-1);
  }

  .icon-cell {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    aspect-ratio: 1;
    border: var(--hairline) solid transparent;
    background: transparent;
    color: var(--text-secondary);
    border-radius: var(--radius-control);
    cursor: pointer;
  }

  .icon-cell:hover {
    background: var(--surface-hover);
  }

  .icon-cell.on {
    border-color: var(--accent);
    background: var(--selection-bg);
  }

  .swatches {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }

  .swatch {
    width: 24px;
    height: 24px;
    border-radius: 999px;
    border: 2px solid transparent;
    background: var(--sw);
    cursor: pointer;
  }

  .swatch.on {
    border-color: var(--text-primary);
  }

  .toggles {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .check {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fz-label);
    color: var(--text-secondary);
    cursor: pointer;
  }

  .btn {
    padding: var(--space-2) var(--space-4);
    border-radius: var(--radius-control);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    cursor: pointer;
    border: var(--hairline) solid var(--border-default);
  }

  .btn.ghost {
    background: var(--surface-raised);
    color: var(--text-secondary);
  }

  .btn.ghost:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .btn.primary {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--accent-contrast, #fff);
  }

  .btn.primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
