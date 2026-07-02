<script lang="ts">
  // the downloadable attachment list for the open message, shown as outlook-style
  // cards at the bottom of the reading pane. inline parts (cid images already
  // shown in the body, and only those actually referenced) are hidden. each card
  // shows a file-type icon, the name and size, and downloads via a native save
  // dialog on click.
  import { IconPaperclip, IconDownload, IconEye, IconDownloadOff } from '@tabler/icons-svelte'
  import { saveAttachment, saveAllAttachments } from '../../lib/api'
  import { formatBytes } from '../../lib/format'
  import { fileIcon, previewKind } from '../../lib/filetype'
  import { openPreview } from '../../stores/preview'
  import { errorMessage, toastError, toastSuccess } from '../../stores/toast'
  import { t } from '../../lib/i18n'
  import type { Attachment } from '../../lib/types'

  export let messageId: number
  export let attachments: Attachment[]

  let busy = new Set<number>()
  let savingAll = false

  $: visible = attachments.filter((a) => !a.inline)
  $: attachmentsLabel = `${visible.length} ${visible.length === 1 ? $t('detail.attachments.one') : $t('detail.attachments.many')}`

  // whether the card click should preview (previewable types) or download.
  function canPreview(att: Attachment): boolean {
    return previewKind(att.contentType, att.filename) !== 'none'
  }

  function activate(att: Attachment): void {
    if (canPreview(att)) {
      openPreview(messageId, att)
    } else {
      void download(att)
    }
  }

  async function download(att: Attachment): Promise<void> {
    if (busy.has(att.id)) {
      return
    }
    busy = new Set(busy).add(att.id)
    try {
      const path = await saveAttachment(messageId, att.id)
      if (path) {
        toastSuccess($t('detail.attachments.saved').replace('{name}', att.filename))
      }
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      const next = new Set(busy)
      next.delete(att.id)
      busy = next
    }
  }

  async function saveAll(): Promise<void> {
    if (savingAll) {
      return
    }
    savingAll = true
    try {
      const dir = await saveAllAttachments(messageId)
      if (dir) {
        toastSuccess($t('detail.attachments.savedAll').replace('{count}', String(visible.length)))
      }
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      savingAll = false
    }
  }
</script>

{#if visible.length > 0}
  <div class="attachments">
    <div class="head">
      <IconPaperclip size={13} stroke={1.6} />
      <span>{attachmentsLabel}</span>
      {#if visible.length > 1}
        <button type="button" class="save-all" class:busy={savingAll} on:click={saveAll}>
          <IconDownload size={13} stroke={1.7} />
          {$t('detail.attachments.saveAll')}
        </button>
      {/if}
    </div>
    <div class="cards">
      {#each visible as att (att.id)}
        <div class="card" class:busy={busy.has(att.id)}>
          <button
            type="button"
            class="card-main"
            on:click={() => activate(att)}
            title={canPreview(att) ? $t('detail.attachments.preview').replace('{name}', att.filename) : $t('detail.attachments.download').replace('{name}', att.filename)}
          >
            <span class="glyph">
              <svelte:component this={fileIcon(att.contentType, att.filename)} size={22} stroke={1.5} />
            </span>
            <span class="meta">
              <span class="name">{att.filename}</span>
              <span class="size">{formatBytes(att.sizeBytes)}</span>
            </span>
            {#if canPreview(att)}
              <span class="dl" aria-hidden="true"><IconEye size={15} stroke={1.7} /></span>
            {/if}
          </button>
          <button type="button" class="dl-btn" title={$t('detail.attachments.download').replace('{name}', att.filename)} aria-label={$t('detail.attachments.download').replace('{name}', att.filename)} on:click={() => download(att)}>
            {#if busy.has(att.id)}
              <IconDownloadOff size={15} stroke={1.7} />
            {:else}
              <IconDownload size={15} stroke={1.7} />
            {/if}
          </button>
        </div>
      {/each}
    </div>
  </div>
{/if}

<style>
  .attachments {
    margin-top: var(--space-4);
    padding-top: var(--space-3);
    border-top: var(--hairline) solid var(--border-subtle);
  }

  .head {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    margin-bottom: var(--space-3);
  }

  .cards {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
  }

  .save-all {
    margin-left: auto;
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: 2px var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-secondary);
    cursor: pointer;
    font-size: var(--fz-meta);
  }
  .save-all:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }
  .save-all.busy {
    opacity: 0.6;
    cursor: default;
  }

  .card {
    display: flex;
    align-items: stretch;
    width: 248px;
    max-width: 100%;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-raised);
    overflow: hidden;
  }

  .card:hover {
    border-color: var(--border-strong, var(--border-default));
  }

  .card.busy {
    opacity: 0.6;
  }

  .card-main {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex: 1;
    min-width: 0;
    padding: var(--space-3);
    border: none;
    background: transparent;
    text-align: left;
    cursor: pointer;
  }
  .card-main:hover {
    background: var(--surface-hover);
  }

  .dl-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 40px;
    border: none;
    border-left: var(--hairline) solid var(--border-subtle);
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
  }
  .dl-btn:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .glyph {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
    flex-shrink: 0;
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--accent);
  }

  .meta {
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

  .size {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .dl {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: var(--text-tertiary);
  }

  .card:hover .dl {
    color: var(--text-primary);
  }
</style>
