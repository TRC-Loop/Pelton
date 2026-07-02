<script lang="ts">
  // the settings-sync setup modal: pick a mode, a folder, and a scope, then
  // confirm. reopening it with an existing setup pre-fills the current values
  // so it doubles as the "change setup" flow.
  import { createEventDispatcher } from 'svelte'
  import { IconX, IconFolder, IconCheck } from '@tabler/icons-svelte'
  import { configureConfigSync, pickConfigSyncFolder, type ConfigSyncStatus } from '../../lib/api'
  import { toastError, errorMessage } from '../../stores/toast'
  import { get } from 'svelte/store'
  import { t } from '../../lib/i18n'

  export let current: ConfigSyncStatus | null

  const dispatch = createEventDispatcher<{ close: void; configured: ConfigSyncStatus }>()

  type ModeChoice = 'copy' | 'readonly'
  type EmailChoice = 'off' | 'metadata' | 'full'

  let mode: ModeChoice = (current?.mode as ModeChoice) || 'copy'
  let path = current?.path || ''
  let syncSettings = current ? current.syncSettings : true
  let emailScope: EmailChoice = (current?.emailScope as EmailChoice) || 'off'
  let saving = false

  async function browse(): Promise<void> {
    try {
      const picked = await pickConfigSyncFolder()
      if (picked) {
        path = picked
      }
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      dispatch('close')
    }
  }

  async function save(): Promise<void> {
    if (!path) {
      toastError(get(t)('configSync.errChooseFolder'))
      return
    }
    if (!syncSettings && emailScope === 'off') {
      toastError(get(t)('configSync.errPickScope'))
      return
    }
    saving = true
    try {
      const status = await configureConfigSync(mode, path, syncSettings, emailScope)
      dispatch('configured', status)
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      saving = false
    }
  }
</script>

<svelte:window on:keydown={onKeydown} />

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="overlay" on:click={() => dispatch('close')}>
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions a11y-no-noninteractive-element-interactions -->
  <div class="modal" role="dialog" aria-modal="true" aria-label={current?.enabled ? $t('configSync.changeTitle') : $t('configSync.setup')} tabindex="-1" on:click|stopPropagation>
    <header>
      <span class="title">{current?.enabled ? $t('configSync.changeTitle') : $t('configSync.setup')}</span>
      <button type="button" class="close" aria-label={$t('configSync.close')} on:click={() => dispatch('close')}>
        <IconX size={18} stroke={1.8} />
      </button>
    </header>

    <div class="body">
      <section>
        <span class="label">{$t('configSync.mode')}</span>
        <div class="options">
          <button type="button" class="option" class:active={mode === 'copy'} on:click={() => (mode = 'copy')}>
            <span class="opt-title">{$t('configSync.modeCopy')} {#if mode === 'copy'}<IconCheck size={14} stroke={2.4} />{/if}</span>
            <span class="opt-sub">{$t('configSync.modeCopyDesc')}</span>
          </button>
          <button type="button" class="option" class:active={mode === 'readonly'} on:click={() => (mode = 'readonly')}>
            <span class="opt-title">{$t('configSync.modeReadonly')} {#if mode === 'readonly'}<IconCheck size={14} stroke={2.4} />{/if}</span>
            <span class="opt-sub">{$t('configSync.modeReadonlyDesc')}</span>
          </button>
        </div>
      </section>

      <section>
        <span class="label">{$t('configSync.folder')}</span>
        <div class="folder-row">
          <span class="folder-path">{path || $t('configSync.noFolder')}</span>
          <button type="button" class="browse" on:click={browse}>
            <IconFolder size={14} stroke={1.8} />
            {$t('configSync.browse')}
          </button>
        </div>
        <p class="sub-hint">{$t('configSync.folderHint')}</p>
      </section>

      <section>
        <span class="label">{$t('configSync.scope')}</span>
        <label class="checkline">
          <input type="checkbox" bind:checked={syncSettings} />
          {$t('configSync.scopeSettings')}
        </label>
        <div class="email-scope">
          <label class="checkline">
            <input type="checkbox" checked={emailScope !== 'off'} on:change={(e) => (emailScope = e.currentTarget.checked ? 'metadata' : 'off')} />
            {$t('configSync.scopeEmail')}
          </label>
          {#if emailScope !== 'off'}
            <div class="sub-options">
              <label class="radioline">
                <input type="radio" name="emailscope" value="metadata" bind:group={emailScope} />
                <span>
                  <span class="opt-title">{$t('configSync.metadataOnlyTitle')}</span>
                  <span class="opt-sub">{$t('configSync.metadataOnlyDesc')}</span>
                </span>
              </label>
              <label class="radioline">
                <input type="radio" name="emailscope" value="full" bind:group={emailScope} />
                <span>
                  <span class="opt-title">{$t('configSync.scopeFullCache')}</span>
                  <span class="opt-sub">{$t('configSync.fullCacheDesc')}</span>
                </span>
              </label>
            </div>
          {/if}
        </div>
      </section>
    </div>

    <footer>
      <button type="button" class="ghost" on:click={() => dispatch('close')}>{$t('configSync.cancel')}</button>
      <button type="button" class="primary" on:click={save} disabled={saving}>
        {saving ? $t('configSync.saving') : current?.enabled ? $t('configSync.save') : $t('configSync.setupButton')}
      </button>
    </footer>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    z-index: 150;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-5);
    background: var(--scrim, rgba(0, 0, 0, 0.4));
  }

  .modal {
    width: 100%;
    max-width: 560px;
    max-height: 86vh;
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
    flex-shrink: 0;
  }

  .title {
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .close {
    display: inline-flex;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    padding: var(--space-1);
    border-radius: var(--radius-control);
  }

  .close:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .body {
    padding: var(--space-4) var(--space-5);
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
  }

  .label {
    display: block;
    margin-bottom: var(--space-2);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }

  .options {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .option {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-raised);
    text-align: left;
    cursor: pointer;
  }

  .option.active {
    border-color: var(--accent);
    box-shadow: 0 0 0 1px var(--accent);
  }

  .opt-title {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }

  .opt-sub {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    line-height: 1.45;
  }

  .folder-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .folder-path {
    flex: 1;
    min-width: 0;
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .browse {
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
    flex-shrink: 0;
  }

  .browse:hover {
    background: var(--surface-hover);
  }

  .sub-hint {
    margin: var(--space-2) 0 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    line-height: 1.45;
  }

  .checkline {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fz-label);
    color: var(--text-primary);
    cursor: pointer;
    padding: var(--space-2) 0;
  }

  .email-scope {
    margin-top: var(--space-2);
  }

  .sub-options {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: var(--space-2) 0 0 var(--space-6);
  }

  .radioline {
    display: flex;
    align-items: flex-start;
    gap: var(--space-2);
    cursor: pointer;
  }

  .radioline input {
    margin-top: 3px;
  }

  .radioline .opt-title {
    display: block;
    font-size: var(--fz-label);
  }

  footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-3);
    padding: var(--space-4) var(--space-5);
    border-top: var(--hairline) solid var(--border-subtle);
    flex-shrink: 0;
  }

  .primary,
  .ghost {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-3) var(--space-5);
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

  .primary:hover {
    filter: brightness(1.05);
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
</style>
