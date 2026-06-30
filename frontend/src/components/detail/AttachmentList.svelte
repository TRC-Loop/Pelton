<script lang="ts">
  // the downloadable attachment list for the open message, shown as outlook-style
  // cards at the bottom of the reading pane. inline parts (cid images already
  // shown in the body, and only those actually referenced) are hidden. each card
  // shows a file-type icon, the name and size, and downloads via a native save
  // dialog on click.
  import { IconPaperclip, IconDownload } from '@tabler/icons-svelte'
  import { saveAttachment } from '../../lib/api'
  import { formatBytes } from '../../lib/format'
  import { fileIcon } from '../../lib/filetype'
  import { errorMessage, toastError, toastSuccess } from '../../stores/toast'
  import type { Attachment } from '../../lib/types'

  export let messageId: number
  export let attachments: Attachment[]

  let busy = new Set<number>()

  $: visible = attachments.filter((a) => !a.inline)

  async function download(att: Attachment): Promise<void> {
    if (busy.has(att.id)) {
      return
    }
    busy = new Set(busy).add(att.id)
    try {
      const path = await saveAttachment(messageId, att.id)
      if (path) {
        toastSuccess(`Saved ${att.filename}`)
      }
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      const next = new Set(busy)
      next.delete(att.id)
      busy = next
    }
  }
</script>

{#if visible.length > 0}
  <div class="attachments">
    <div class="head">
      <IconPaperclip size={13} stroke={1.6} />
      <span>{visible.length} attachment{visible.length === 1 ? '' : 's'}</span>
    </div>
    <div class="cards">
      {#each visible as att (att.id)}
        <button
          type="button"
          class="card"
          class:busy={busy.has(att.id)}
          on:click={() => download(att)}
          title={`Download ${att.filename}`}
        >
          <span class="glyph">
            <svelte:component this={fileIcon(att.contentType, att.filename)} size={22} stroke={1.5} />
          </span>
          <span class="meta">
            <span class="name">{att.filename}</span>
            <span class="size">{formatBytes(att.sizeBytes)}</span>
          </span>
          <span class="dl" aria-hidden="true">
            <IconDownload size={15} stroke={1.7} />
          </span>
        </button>
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

  .card {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 248px;
    max-width: 100%;
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-raised);
    text-align: left;
    cursor: pointer;
  }

  .card:hover {
    background: var(--surface-hover);
    border-color: var(--border-strong, var(--border-default));
  }

  .card.busy {
    opacity: 0.6;
    cursor: default;
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
