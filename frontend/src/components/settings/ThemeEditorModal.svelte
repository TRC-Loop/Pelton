<script lang="ts">
  // the palette editor for custom color themes (#57): a name, a light/dark
  // base and grouped color fields for the themeable tokens. every change
  // previews live on the whole app; Save writes the palette as an installed
  // theme through the backend, Cancel lets the parent restore the active
  // theme. fields left empty fall back to the base's built-in value.
  import { createEventDispatcher } from 'svelte'
  import { IconX } from '@tabler/icons-svelte'
  import { saveCustomTheme } from '../../lib/api'
  import type { ThemeInfo } from '../../lib/types'
  import { applyUserTheme } from '../../theme/usertheme'
  import { isValidHex, normalizeHex } from '../../theme/accent'
  import { toastError, errorMessage } from '../../stores/toast'
  import { t } from '../../lib/i18n'

  /** id of the installed theme being edited; '' creates a new theme. */
  export let id = ''
  /** prefilled name (empty for a new theme). */
  export let name = ''
  /** prefilled base. */
  export let base: 'light' | 'dark' = 'light'
  /** prefilled token map (the seed a preset or installed theme provides). */
  export let tokens: Record<string, string> = {}

  const dispatch = createEventDispatcher<{ saved: ThemeInfo; close: void }>()

  // the editable token surface, grouped for the form. semantic -bg partners
  // are derived on save, not edited.
  const groups: { key: string; tokens: string[] }[] = [
    { key: 'surfaces', tokens: ['surface-base', 'surface-raised', 'surface-overlay', 'surface-sunken', 'surface-hover'] },
    { key: 'text', tokens: ['text-primary', 'text-secondary', 'text-tertiary'] },
    { key: 'borders', tokens: ['border-subtle', 'border-default', 'border-strong'] },
    { key: 'status', tokens: ['success', 'warning', 'danger'] },
  ]
  const editorTokens = groups.flatMap((g) => g.tokens)
  const statusTokens = groups[3].tokens

  // the built-in palettes from tokens.css, as starting values for a new theme
  // so every field begins on what the base actually looks like.
  const baseDefaults: Record<'light' | 'dark', Record<string, string>> = {
    light: {
      'surface-base': '#f6f6f7',
      'surface-raised': '#ffffff',
      'surface-overlay': '#ffffff',
      'surface-sunken': '#efeff1',
      'surface-hover': '#ececef',
      'text-primary': '#1a1b1e',
      'text-secondary': '#4a4d54',
      'text-tertiary': '#797d86',
      'border-subtle': 'rgba(0, 0, 0, 0.06)',
      'border-default': 'rgba(0, 0, 0, 0.12)',
      'border-strong': 'rgba(0, 0, 0, 0.22)',
      success: '#1a7f4b',
      warning: '#9a6700',
      danger: '#c0392b',
    },
    dark: {
      'surface-base': '#111214',
      'surface-raised': '#1a1c1f',
      'surface-overlay': '#222428',
      'surface-sunken': '#161719',
      'surface-hover': '#25282d',
      'text-primary': '#ecedee',
      'text-secondary': '#b4b7bd',
      'text-tertiary': '#7e828b',
      'border-subtle': 'rgba(255, 255, 255, 0.06)',
      'border-default': 'rgba(255, 255, 255, 0.12)',
      'border-strong': 'rgba(255, 255, 255, 0.22)',
      success: '#3fb27a',
      warning: '#d9a441',
      danger: '#e5675a',
    },
  }

  function seedValues(): Record<string, string> {
    const source = Object.keys(tokens).length ? tokens : baseDefaults[base]
    const values: Record<string, string> = {}
    for (const token of editorTokens) {
      values[token] = source[token] ?? ''
    }
    return values
  }

  let values = seedValues()
  // any field edit stops base switches from re-seeding the form.
  let touched = Object.keys(tokens).length > 0
  let saving = false

  function setBase(next: 'light' | 'dark'): void {
    base = next
    if (!touched) {
      values = seedValues()
    }
  }

  // safeValue mirrors the backend's token-value check so the live preview
  // never injects something the save would reject.
  function safeValue(value: string): boolean {
    return (
      value.length > 0 &&
      value.length <= 300 &&
      !/[;{}<>@\\]/.test(value) &&
      !value.toLowerCase().includes('url(')
    )
  }

  // buildTokens assembles the save payload: seed tokens the editor does not
  // manage stay as they are, edited fields replace their token (or drop it
  // when cleared), and each status color gets its derived background.
  function buildTokens(): Record<string, string> {
    const result: Record<string, string> = {}
    for (const [token, value] of Object.entries(tokens)) {
      if (!editorTokens.includes(token) && !statusTokens.some((s) => token === s + '-bg')) {
        result[token] = value
      }
    }
    for (const token of editorTokens) {
      const value = values[token].trim()
      if (value && safeValue(value)) {
        result[token] = value
      }
    }
    for (const token of statusTokens) {
      const value = values[token].trim()
      if (value && safeValue(value)) {
        const mix = base === 'dark' ? '16%' : '12%'
        result[token + '-bg'] = `color-mix(in srgb, ${value} ${mix}, transparent)`
      }
    }
    return result
  }

  // live preview: re-inject the draft on every change. the parameters only
  // exist so the reactive statement tracks them.
  function previewDraft(_values: Record<string, string>, draftBase: string): void {
    applyUserTheme({ id: 'palette-editor-draft', base: draftBase, tokens: buildTokens(), css: '', icons: {} })
  }
  $: previewDraft(values, base)

  // wellColor gives the color well a hex to show; non-hex values (rgba,
  // color-mix) keep the text input authoritative.
  function wellColor(value: string): string {
    return isValidHex(value) ? normalizeHex(value) : '#888888'
  }

  function onWellInput(token: string, event: Event): void {
    values = { ...values, [token]: (event.currentTarget as HTMLInputElement).value }
    touched = true
  }

  function onTextInput(token: string, event: Event): void {
    values = { ...values, [token]: (event.currentTarget as HTMLInputElement).value }
    touched = true
  }

  async function save(): Promise<void> {
    if (!name.trim()) {
      toastError($t('themeEditor.nameMissing'))
      return
    }
    saving = true
    try {
      const info = await saveCustomTheme({ id, name: name.trim(), base, tokens: buildTokens() })
      dispatch('saved', info)
    } catch (err) {
      toastError(errorMessage(err))
      saving = false
    }
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      dispatch('close')
    }
  }
</script>

<svelte:window on:keydown={onKeydown} />

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="overlay" on:click={() => dispatch('close')}>
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" aria-modal="true" aria-label={$t('themeEditor.title')} tabindex="-1" on:click|stopPropagation>
    <header>
      <span class="m-title">{$t('themeEditor.title')}</span>
      <button type="button" class="m-close" aria-label={$t('about.close')} on:click={() => dispatch('close')}>
        <IconX size={18} stroke={1.8} />
      </button>
    </header>

    <div class="m-body">
      <div class="top-row">
        <label class="field name-field">
          <span class="label">{$t('themeEditor.nameLabel')}</span>
          <input type="text" bind:value={name} placeholder={$t('themeEditor.namePlaceholder')} maxlength="60" />
        </label>
        <div class="field">
          <span class="label">{$t('themeEditor.baseLabel')}</span>
          <div class="segmented">
            <button type="button" class:on={base === 'light'} on:click={() => setBase('light')}>{$t('themeEditor.baseLight')}</button>
            <button type="button" class:on={base === 'dark'} on:click={() => setBase('dark')}>{$t('themeEditor.baseDark')}</button>
          </div>
        </div>
      </div>
      <p class="hint">{$t('themeEditor.hint')}</p>

      {#each groups as group (group.key)}
        <p class="group-heading">{$t('themeEditor.group.' + group.key)}</p>
        <div class="token-grid">
          {#each group.tokens as token (token)}
            <label class="token-row">
              <span class="token-label">{$t('themeEditor.token.' + token)}</span>
              <span class="inputs">
                <input
                  type="color"
                  class="well"
                  value={wellColor(values[token])}
                  on:input={(e) => onWellInput(token, e)}
                />
                <input
                  type="text"
                  class="value mono"
                  value={values[token]}
                  spellcheck="false"
                  on:input={(e) => onTextInput(token, e)}
                />
              </span>
            </label>
          {/each}
        </div>
      {/each}
    </div>

    <footer>
      <button type="button" class="ghost-btn" on:click={() => dispatch('close')}>{$t('themeEditor.cancel')}</button>
      <button type="button" class="primary-btn" disabled={saving} on:click={save}>{$t('themeEditor.save')}</button>
    </footer>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    z-index: 140;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-5);
    background: var(--scrim, rgba(0, 0, 0, 0.4));
  }

  .modal {
    width: 100%;
    max-width: 560px;
    max-height: 84vh;
    display: flex;
    flex-direction: column;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
    overflow: hidden;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4) var(--space-5);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .m-title {
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .m-close {
    display: inline-flex;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    padding: var(--space-1);
    border-radius: var(--radius-control);
  }

  .m-close:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .m-body {
    padding: var(--space-4) var(--space-5);
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .top-row {
    display: flex;
    gap: var(--space-4);
    align-items: flex-end;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .name-field {
    flex: 1;
  }

  .label {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .field input[type='text'] {
    height: var(--control-height);
    padding: 0 var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
  }

  .segmented {
    display: inline-flex;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    overflow: hidden;
  }

  .segmented button {
    padding: 0 var(--space-4);
    height: calc(var(--control-height) - 2px);
    border: none;
    background: transparent;
    color: var(--text-secondary);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .segmented button.on {
    background: var(--selection-bg-strong);
    color: var(--text-primary);
  }

  .hint {
    margin: var(--space-1) 0 0;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .group-heading {
    margin: var(--space-3) 0 0;
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }

  .token-grid {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .token-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .token-label {
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }

  .inputs {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
  }

  .well {
    width: 28px;
    height: 24px;
    padding: 0;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: transparent;
    cursor: pointer;
  }

  .value {
    width: 200px;
    height: 24px;
    padding: 0 var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-meta);
  }

  .mono {
    font-family: var(--font-mono);
  }

  footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    padding: var(--space-4) var(--space-5);
    border-top: var(--hairline) solid var(--border-subtle);
  }

  .ghost-btn {
    padding: var(--space-2) var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .ghost-btn:hover {
    background: var(--surface-hover);
  }

  .primary-btn {
    padding: var(--space-2) var(--space-4);
    border: none;
    border-radius: var(--radius-control);
    background: var(--accent);
    color: var(--accent-fg);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .primary-btn:disabled {
    opacity: 0.6;
    cursor: default;
  }
</style>
