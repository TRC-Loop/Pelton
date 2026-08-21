<script lang="ts">
  // the read-before-import step for a .peltontheme container: shows the
  // theme's metadata, every stylesheet's raw content in a read-only viewer,
  // and - when the css references the network - an explicit warning with the
  // user's Allow/Block choice. nothing is installed until the confirm button.
  import { createEventDispatcher } from 'svelte'
  import Modal from '../common/Modal.svelte'
  import { IconAlertTriangle } from '@tabler/icons-svelte'
  import { confirmThemeImport } from '../../lib/api'
  import type { ThemeImportPreview, ThemeInfo } from '../../lib/types'
  import { toastError, errorMessage } from '../../stores/toast'
  import { t } from '../../lib/i18n'

  /** the parsed preview returned by previewThemeImport. */
  export let preview: ThemeImportPreview

  const dispatch = createEventDispatcher<{ installed: ThemeInfo; close: void }>()

  // the parts choice: both on by default, so a plain import stays one click.
  // it is only offered when the theme actually carries both parts.
  let importTokens = true
  let importCSS = true
  $: hasTokens = preview.tokenCount > 0
  $: hasCSS = (preview.cssFiles?.length ?? 0) > 0
  $: showParts = hasTokens && hasCSS
  $: nothingPicked = showParts && !importTokens && !importCSS

  $: remoteRefs = importCSS ? (preview.cssFiles ?? []).flatMap((f) => f.remoteRefs ?? []) : []
  // the remote choice: strip network references by default; allowing them is
  // the explicit opt-in.
  let allowRemote = false
  let installing = false

  async function install(): Promise<void> {
    installing = true
    try {
      const info = await confirmThemeImport(preview.path, allowRemote, importTokens, importCSS)
      dispatch('installed', info)
    } catch (err) {
      toastError(errorMessage(err))
      installing = false
    }
  }

</script>

<Modal title={$t('themes.importTitle')} size="medium" on:close={() => dispatch('close')}>
    <div class="m-body">
      <div class="meta">
        <span class="name">{preview.info.name}</span>
        <span class="sub">
          {#if preview.info.author}{$t('themes.by').replace('{author}', preview.info.author)}{/if}
          {#if preview.info.version}&nbsp;· v{preview.info.version}{/if}
          &nbsp;· {preview.info.base === 'dark' ? $t('themes.baseDark') : $t('themes.baseLight')}
        </span>
        {#if preview.info.description}
          <p class="desc">{preview.info.description}</p>
        {/if}
        {#if preview.info.preview}
          <img class="shot" src={preview.info.preview} alt="" draggable="false" />
        {/if}
      </div>

      {#if preview.updatesExisting}
        <p class="note">{$t('themes.updatesExisting').replace('{version}', preview.installedVersion || '?')}</p>
      {/if}
      {#if preview.info.compatWarning}
        <div class="warn-box">
          <IconAlertTriangle size={15} stroke={1.8} />
          <span>{$t('themes.compatWarning').replace('{detail}', preview.info.compatWarning)}</span>
        </div>
      {/if}

      {#if showParts}
        <div class="parts">
          <p class="parts-heading">{$t('themes.partsHeading')}</p>
          <label class="choice">
            <input type="checkbox" bind:checked={importTokens} />
            <span>
              {$t('themes.partColors')}
              <span class="sub">{$t('themes.partColorsDetail').replace('{count}', String(preview.tokenCount))}</span>
            </span>
          </label>
          <label class="choice">
            <input type="checkbox" bind:checked={importCSS} />
            <span>
              {$t('themes.partCss')}
              <span class="sub">{$t('themes.partCssDetail').replace('{count}', String(preview.cssFiles.length))}</span>
            </span>
          </label>
          <p class="hint">{$t('themes.partsHint')}</p>
        </div>
      {/if}

      {#if preview.cssFiles?.length}
        <p class="css-heading">{$t('themes.cssHeading')}</p>
        <p class="hint">{$t('themes.cssHint')}</p>
        {#each preview.cssFiles as file (file.path)}
          <details class="css-file">
            <summary>
              <span class="mono">{file.path}</span>
              {#if file.remoteRefs?.length}
                <span class="badge warn">{$t('themes.remoteBadge')}</span>
              {/if}
            </summary>
            <pre class="css-view">{file.content}</pre>
          </details>
        {/each}
      {/if}

      {#if remoteRefs.length}
        <div class="warn-box remote">
          <IconAlertTriangle size={15} stroke={1.8} />
          <div>
            <p class="warn-title">{$t('themes.remoteWarningTitle')}</p>
            <p class="warn-text">{$t('themes.remoteWarningText')}</p>
            <ul class="ref-list">
              {#each remoteRefs as ref (ref)}
                <li class="mono">{ref}</li>
              {/each}
            </ul>
            <label class="choice">
              <input type="radio" bind:group={allowRemote} value={false} />
              {$t('themes.remoteBlock')}
            </label>
            <label class="choice">
              <input type="radio" bind:group={allowRemote} value={true} />
              {$t('themes.remoteAllow')}
            </label>
          </div>
        </div>
      {/if}
    </div>

  <svelte:fragment slot="footer">
    <button type="button" class="ghost-btn" on:click={() => dispatch('close')}>{$t('themes.importCancel')}</button>
    <button type="button" class="primary-btn" disabled={installing || nothingPicked} on:click={install}>
      {preview.updatesExisting ? $t('themes.importUpdate') : $t('themes.importInstall')}
    </button>
  </svelte:fragment>
</Modal>

<style>

  .m-body {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .meta {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .name {
    font-size: var(--fz-body);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .sub {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .desc {
    margin: var(--space-1) 0 0;
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }

  .shot {
    margin-top: var(--space-2);
    max-width: 100%;
    border-radius: var(--radius-control);
    border: var(--hairline) solid var(--border-subtle);
  }

  .note {
    margin: 0;
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }

  .hint {
    margin: 0;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .css-heading {
    margin: var(--space-2) 0 0;
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }

  .css-file {
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-control);
  }

  .css-file summary {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    cursor: var(--cursor-action);
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }

  .css-view {
    margin: 0;
    padding: var(--space-3);
    border-top: var(--hairline) solid var(--border-subtle);
    max-height: 220px;
    overflow: auto;
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
    white-space: pre-wrap;
    word-break: break-word;
    -webkit-user-select: text;
    user-select: text;
  }

  .mono {
    font-family: var(--font-mono);
  }

  .badge.warn {
    display: inline-flex;
    padding: 0 var(--space-2);
    border-radius: var(--radius-control);
    font-size: var(--fz-meta);
    color: var(--warning);
    background: var(--warning-bg);
  }

  .warn-box {
    display: flex;
    gap: var(--space-3);
    padding: var(--space-3);
    border-radius: var(--radius-control);
    background: var(--warning-bg);
    color: var(--warning);
    font-size: var(--fz-label);
    align-items: flex-start;
  }

  .warn-title {
    margin: 0;
    font-weight: var(--fw-semibold);
  }

  .warn-text {
    margin: var(--space-1) 0 0;
    color: var(--text-secondary);
  }

  .ref-list {
    margin: var(--space-2) 0;
    padding-left: var(--space-5);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
    word-break: break-all;
  }

  .choice {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) 0;
    font-size: var(--fz-label);
    color: var(--text-primary);
    cursor: var(--cursor-action);
  }

  .parts {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
  }

  .parts-heading {
    margin: 0 0 var(--space-1);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }

  .choice .sub {
    display: block;
  }

  .ghost-btn {
    padding: var(--space-2) var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: var(--cursor-action);
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
    cursor: var(--cursor-action);
  }

  .primary-btn:disabled {
    opacity: 0.6;
    cursor: default;
  }
</style>
