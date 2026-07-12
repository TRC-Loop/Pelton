<script lang="ts">
  // the Themes settings category: a gallery of the built-in default plus
  // every installed .peltontheme, with import/export/delete. activating a
  // card applies immediately; the import flow goes through ThemeImportModal
  // (read-before-import, remote-reference choice).
  import { onMount } from 'svelte'
  import { IconFileImport, IconRefresh, IconTrash, IconUpload, IconAlertTriangle, IconWorld } from '@tabler/icons-svelte'
  import { listThemes, previewThemeImport, deleteTheme, exportTheme } from '../../lib/api'
  import type { ThemeInfo, ThemeImportPreview } from '../../lib/types'
  import { prefs, setThemeId } from '../../stores/prefs'
  import { toastInfo, toastError, errorMessage } from '../../stores/toast'
  import { t } from '../../lib/i18n'
  import ThemeImportModal from './ThemeImportModal.svelte'

  let themes: ThemeInfo[] = []
  let importPreview: ThemeImportPreview | null = null
  // id of the theme whose delete button is in its confirm step.
  let confirmingDelete = ''

  async function reload(): Promise<void> {
    try {
      themes = await listThemes()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }
  onMount(reload)

  async function activate(id: string): Promise<void> {
    try {
      await setThemeId(id)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function startImport(): Promise<void> {
    try {
      const preview = await previewThemeImport()
      if (!preview.canceled) {
        importPreview = preview
      }
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function onInstalled(event: CustomEvent<ThemeInfo>): Promise<void> {
    importPreview = null
    await reload()
    await activate(event.detail.id)
  }

  async function onExport(id: string): Promise<void> {
    try {
      const path = await exportTheme(id)
      if (path) {
        toastInfo($t('themes.exported'))
      }
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function onDelete(id: string): Promise<void> {
    if (confirmingDelete !== id) {
      confirmingDelete = id
      return
    }
    confirmingDelete = ''
    try {
      await deleteTheme(id)
      if ($prefs.themeId === id) {
        await setThemeId('')
      }
      await reload()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }
</script>

<p class="hint">{$t('themes.hint')}</p>

<div class="toolbar">
  <button type="button" class="action-btn" on:click={startImport}>
    <IconFileImport size={15} stroke={1.6} />
    {$t('themes.import')}
  </button>
  <button type="button" class="action-btn" on:click={reload} title={$t('themes.reloadHint')}>
    <IconRefresh size={15} stroke={1.6} />
    {$t('themes.reload')}
  </button>
</div>

<div class="gallery">
  <button type="button" class="card" class:active={$prefs.themeId === ''} on:click={() => activate('')}>
    <div class="swatch default-swatch"><span class="half light"></span><span class="half dark"></span></div>
    <div class="meta">
      <span class="name">{$t('themes.defaultName')}</span>
      <span class="sub">{$t('themes.defaultDesc')}</span>
    </div>
  </button>

  {#each themes as theme (theme.id)}
    <div class="card" class:active={$prefs.themeId === theme.id}>
      <button type="button" class="card-body" on:click={() => activate(theme.id)}>
        {#if theme.preview}
          <img class="preview" src={theme.preview} alt="" draggable="false" />
        {:else}
          <div class="swatch" class:dark-base={theme.base === 'dark'}></div>
        {/if}
        <div class="meta">
          <span class="name">{theme.name}</span>
          <span class="sub">
            {#if theme.author}{$t('themes.by').replace('{author}', theme.author)}{/if}
            {#if theme.version}&nbsp;· v{theme.version}{/if}
          </span>
          {#if theme.compatWarning}
            <span class="badge warn" title={theme.compatWarning}>
              <IconAlertTriangle size={12} stroke={1.8} />
              {$t('themes.compatBadge')}
            </span>
          {/if}
          {#if theme.remoteRefs?.length}
            <span class="badge warn" title={theme.remoteRefs.join('\n')}>
              <IconWorld size={12} stroke={1.8} />
              {$t('themes.remoteBadge')}
            </span>
          {/if}
        </div>
      </button>
      <div class="card-actions">
        <button type="button" class="icon-btn" title={$t('themes.export')} on:click={() => onExport(theme.id)}>
          <IconUpload size={14} stroke={1.6} />
        </button>
        <button
          type="button"
          class="icon-btn danger"
          title={$t('themes.delete')}
          on:click={() => onDelete(theme.id)}
          on:mouseleave={() => (confirmingDelete = '')}
        >
          {#if confirmingDelete === theme.id}
            <span class="confirm-text">{$t('themes.deleteConfirm')}</span>
          {:else}
            <IconTrash size={14} stroke={1.6} />
          {/if}
        </button>
      </div>
    </div>
  {/each}
</div>

{#if importPreview}
  <ThemeImportModal preview={importPreview} on:installed={onInstalled} on:close={() => (importPreview = null)} />
{/if}

<style>
  .hint {
    margin: 0 0 var(--space-4);
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .toolbar {
    display: flex;
    gap: var(--space-2);
    margin-bottom: var(--space-4);
  }

  .action-btn {
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
  }

  .action-btn:hover {
    background: var(--surface-hover);
  }

  .gallery {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
    gap: var(--space-3);
  }

  .card {
    display: flex;
    flex-direction: column;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-raised);
    overflow: hidden;
    text-align: left;
    padding: 0;
  }

  button.card {
    cursor: pointer;
  }

  .card.active {
    outline: 2px solid var(--accent);
    outline-offset: -1px;
  }

  .card-body {
    display: flex;
    flex-direction: column;
    border: none;
    background: none;
    padding: 0;
    cursor: pointer;
    text-align: left;
  }

  .preview,
  .swatch {
    width: 100%;
    height: 72px;
    object-fit: cover;
    display: block;
  }

  .swatch {
    background: var(--surface-sunken);
  }

  .swatch.dark-base {
    background: #1a1c1f;
  }

  .default-swatch {
    display: flex;
  }

  .half {
    flex: 1;
  }

  .half.light {
    background: #f6f6f7;
  }

  .half.dark {
    background: #111214;
  }

  .meta {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-3);
  }

  .name {
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }

  .sub {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .badge {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    align-self: flex-start;
    padding: 1px var(--space-2);
    border-radius: var(--radius-control);
    font-size: var(--fz-meta);
  }

  .badge.warn {
    color: var(--warning);
    background: var(--warning-bg);
  }

  .card-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-1);
    padding: 0 var(--space-2) var(--space-2);
  }

  .icon-btn {
    display: inline-flex;
    align-items: center;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-control);
  }

  .icon-btn:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .icon-btn.danger:hover {
    color: var(--danger);
    background: var(--danger-bg);
  }

  .confirm-text {
    font-size: var(--fz-meta);
    color: var(--danger);
  }
</style>
