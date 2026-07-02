<script lang="ts">
  // the formatting toolbar shown above the body in markdown mode. it only
  // dispatches intents; the compose pane applies them to its textarea so the
  // selection and caret are handled in one place. a preview toggle flips the body
  // between the markdown source and a github-style rendered preview.
  import { createEventDispatcher } from 'svelte'
  import {
    IconBold,
    IconItalic,
    IconCode,
    IconLink,
    IconList,
    IconQuote,
    IconHeading,
    IconEye,
    IconPencil,
  } from '@tabler/icons-svelte'
  import { t } from '../../lib/i18n'

  export let preview = false

  const dispatch = createEventDispatcher<{ format: string; togglePreview: void }>()

  $: actions = [
    { name: 'bold', icon: IconBold, label: $t('compose.toolbar.bold') },
    { name: 'italic', icon: IconItalic, label: $t('compose.toolbar.italic') },
    { name: 'code', icon: IconCode, label: $t('compose.toolbar.code') },
    { name: 'link', icon: IconLink, label: $t('compose.toolbar.link') },
    { name: 'heading', icon: IconHeading, label: $t('compose.toolbar.heading') },
    { name: 'list', icon: IconList, label: $t('compose.toolbar.list') },
    { name: 'quote', icon: IconQuote, label: $t('compose.toolbar.quote') },
  ]
</script>

<div class="toolbar" role="toolbar" aria-label={$t('compose.toolbar.ariaLabel')}>
  {#each actions as a (a.name)}
    <button type="button" title={a.label} aria-label={a.label} disabled={preview} on:click={() => dispatch('format', a.name)}>
      <svelte:component this={a.icon} size={16} stroke={1.7} />
    </button>
  {/each}
  <span class="spacer"></span>
  <button type="button" class="preview-btn" class:active={preview} title={$t('compose.toolbar.togglePreview')} on:click={() => dispatch('togglePreview')}>
    {#if preview}
      <IconPencil size={15} stroke={1.7} /> {$t('compose.toolbar.edit')}
    {:else}
      <IconEye size={15} stroke={1.7} /> {$t('compose.toolbar.preview')}
    {/if}
  </button>
</div>

<style>
  .toolbar {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1) var(--space-2);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  button {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    border: none;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    padding: var(--space-2);
    border-radius: var(--radius-control);
  }

  button:hover:not(:disabled) {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  button:disabled {
    opacity: 0.4;
    cursor: default;
  }

  .spacer {
    flex: 1;
  }

  .preview-btn {
    font-size: var(--fz-meta);
  }

  .preview-btn.active {
    background: var(--selection-bg);
    color: var(--text-primary);
  }
</style>
