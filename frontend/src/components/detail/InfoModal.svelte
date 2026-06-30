<script lang="ts">
  // the message info modal opened from the "i" button. it lays out everything we
  // know about the open message in copyable rows and offers a one-click "copy as
  // Markdown" of the whole thing. raw headers and parsed authentication results
  // (spf/dkim/dmarc) are a documented backend follow-up: they need an on-demand
  // header fetch, so we show an honest placeholder rather than faking a result.
  import { createEventDispatcher } from 'svelte'
  import { IconX, IconCopy, IconCheck } from '@tabler/icons-svelte'
  import { formatFullDate, formatBytes } from '../../lib/format'
  import { toastSuccess, toastError } from '../../stores/toast'
  import type { MessageDetail } from '../../lib/types'

  export let detail: MessageDetail

  const dispatch = createEventDispatcher<{ close: void }>()

  const pgpLabel: Record<string, string> = {
    none: 'No PGP markers',
    signed: 'PGP signed',
    encrypted: 'PGP encrypted',
  }

  interface Field {
    label: string
    value: string
  }

  $: fields = buildFields(detail)
  function buildFields(d: MessageDetail): Field[] {
    const list: Field[] = [
      { label: 'Subject', value: d.subject || '(no subject)' },
      { label: 'From', value: d.fromName ? `${d.fromName} <${d.fromAddress}>` : d.fromAddress },
      { label: 'To', value: d.toAddresses || '—' },
    ]
    if (d.ccAddresses) {
      list.push({ label: 'Cc', value: d.ccAddresses })
    }
    list.push(
      { label: 'Date', value: formatFullDate(d.date) || d.date },
      { label: 'Account', value: d.accountEmail },
      { label: 'Folder', value: d.folderName },
      { label: 'Format', value: d.isHtml ? 'HTML' : 'Plain text' },
      { label: 'Remote content', value: d.hasRemoteContent ? 'Present (blocked by default)' : 'None' },
      { label: 'Encryption', value: pgpLabel[d.pgp] ?? d.pgp },
    )
    if (d.attachments.length > 0) {
      const names = d.attachments.map((a) => `${a.filename} (${formatBytes(a.sizeBytes)})`).join(', ')
      list.push({ label: 'Attachments', value: names })
    }
    return list
  }

  function buildMarkdown(): string {
    const lines = ['## Message details', '']
    for (const f of fields) {
      lines.push(`- **${f.label}:** ${f.value}`)
    }
    return lines.join('\n')
  }

  async function copyAll(): Promise<void> {
    try {
      await navigator.clipboard.writeText(buildMarkdown())
      toastSuccess('Copied message details as Markdown.')
    } catch {
      toastError('Could not copy to the clipboard.')
    }
  }

  async function copyOne(field: Field): Promise<void> {
    try {
      await navigator.clipboard.writeText(field.value)
      toastSuccess(`Copied ${field.label.toLowerCase()}.`)
    } catch {
      toastError('Could not copy to the clipboard.')
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
  <div class="modal" role="dialog" aria-modal="true" aria-label="Message information" on:click|stopPropagation>
    <header>
      <span class="title">Message info</span>
      <button type="button" class="close" aria-label="Close" on:click={() => dispatch('close')}>
        <IconX size={18} stroke={1.8} />
      </button>
    </header>

    <div class="fields selectable">
      {#each fields as field}
        <div class="field">
          <span class="key">{field.label}</span>
          <span class="val">{field.value}</span>
          <button type="button" class="copy" aria-label={`Copy ${field.label}`} on:click={() => copyOne(field)}>
            <IconCopy size={14} stroke={1.6} />
          </button>
        </div>
      {/each}

      <div class="field note">
        <span class="key">Authentication</span>
        <span class="val muted">
          Raw headers and SPF / DKIM / DMARC results are not available yet. They are a planned
          on-demand fetch, so nothing here is guessed.
        </span>
      </div>
    </div>

    <footer>
      <button type="button" class="primary" on:click={copyAll}>
        <IconCheck size={15} stroke={1.8} />
        Copy as Markdown
      </button>
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
    max-width: 520px;
    max-height: 80vh;
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

  .fields {
    padding: var(--space-3) var(--space-5);
    overflow-y: auto;
  }

  .field {
    display: grid;
    grid-template-columns: 120px 1fr auto;
    align-items: start;
    gap: var(--space-3);
    padding: var(--space-3) 0;
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .field:last-child {
    border-bottom: none;
  }

  .key {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }

  .val {
    font-size: var(--fz-label);
    color: var(--text-primary);
    word-break: break-word;
  }

  .val.muted {
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .copy {
    display: inline-flex;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: var(--space-1);
    border-radius: var(--radius-control);
  }

  .copy:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .note .val {
    grid-column: 2 / span 2;
  }

  footer {
    display: flex;
    justify-content: flex-end;
    padding: var(--space-4) var(--space-5);
    border-top: var(--hairline) solid var(--border-subtle);
  }

  .primary {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-3) var(--space-5);
    border: none;
    border-radius: var(--radius-control);
    background: var(--accent);
    color: var(--accent-fg);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    cursor: pointer;
  }

  .primary:hover {
    filter: brightness(1.05);
  }
</style>
