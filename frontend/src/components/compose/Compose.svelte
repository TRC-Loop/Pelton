<script lang="ts">
  // a single compose pane. it edits one session: addresses, subject, the body in
  // the chosen editor mode, and attachments. the floating pane can be minimized
  // to its title bar or expanded to fill the window. markdown mode gets a
  // formatting toolbar and a github-style live preview. send enqueues to the
  // outbox (with the undo-send window when enabled); save stores a local draft.
  import { tick, onMount } from 'svelte'
  import { marked } from 'marked'
  import {
    IconX,
    IconSend,
    IconDeviceFloppy,
    IconMinus,
    IconArrowsDiagonal,
    IconArrowsDiagonalMinimize2,
  } from '@tabler/icons-svelte'
  import AddressFields from './AddressFields.svelte'
  import EditorModeSwitch from './EditorModeSwitch.svelte'
  import EditorToolbar from './EditorToolbar.svelte'
  import AttachmentPicker from './AttachmentPicker.svelte'
  import { updateCompose, closeCompose, setComposeFullscreenDefault, type ComposeSession } from '../../stores/compose'
  import { signatures, signatureById, getAccountSignatures } from '../../stores/signatures'
  import type { Signature } from '../../lib/types'
  import { sidebar } from '../../stores/accounts'
  import { sendMessage, saveDraft, deleteDraft } from '../../lib/api'
  import { loadOutbox } from '../../stores/outbox'
  import { scheduleUndo } from '../../stores/undosend'
  import { prefs } from '../../stores/prefs'
  import { buildRequest, hasRecipients } from '../../lib/mailcompose'
  import { errorMessage, toastError, toastSuccess } from '../../stores/toast'
  import type { EditorMode } from '../../lib/types'

  export let session: ComposeSession

  let sending = false
  let preview = false
  let bodyEl: HTMLTextAreaElement

  // the rich editor (tiptap) is code-split: it only loads when the session is in
  // wysiwyg mode, keeping prosemirror out of the main bundle.
  let RichEditor: typeof import('./RichEditor.svelte').default | null = null
  $: if (session.mode === 'wysiwyg' && !RichEditor) {
    void import('./RichEditor.svelte').then((m) => (RichEditor = m.default))
  }

  $: accounts = $sidebar.data?.accounts ?? []
  $: previewHtml = marked.parse(session.body || '', { async: false }) as string

  // withBlock inserts a signature block into a body: headers go to the top,
  // footers to the bottom after the standard "-- " delimiter.
  function withBlock(body: string, sig: Signature): string {
    if (sig.kind === 'header') {
      return `${sig.content}\n\n${body}`
    }
    return `${body.replace(/\s*$/, '')}\n\n-- \n${sig.content}`
  }

  // on first open, insert the account's default header/footer once. drafts and
  // reopened messages have signaturesApplied already set, so they are skipped.
  onMount(async () => {
    if (session.signaturesApplied) {
      return
    }
    try {
      const assigned = await getAccountSignatures(session.accountId)
      let body = session.body
      const header = signatureById(assigned.headerId)
      const footer = signatureById(assigned.footerId)
      if (header) {
        body = withBlock(body, header)
      }
      if (footer) {
        body = withBlock(body, footer)
      }
      updateCompose(session.id, { body, signaturesApplied: true })
    } catch {
      updateCompose(session.id, { signaturesApplied: true })
    }
  })

  // insertSignature lets the user drop a different block into the body from the
  // compose footer menu (the "change in compose" path).
  function insertSignature(event: Event): void {
    const select = event.currentTarget as HTMLSelectElement
    const id = Number(select.value)
    select.value = ''
    const sig = signatureById(id)
    if (!sig) {
      return
    }
    updateCompose(session.id, { body: withBlock(session.body, sig) })
  }

  function setMode(event: CustomEvent<EditorMode>): void {
    if (event.detail !== 'markdown') {
      preview = false
    }
    updateCompose(session.id, { mode: event.detail })
  }

  function toggleFullscreen(): void {
    const fullscreen = !session.fullscreen
    updateCompose(session.id, { fullscreen, minimized: false })
    // remember the size so the next compose opens the same way.
    setComposeFullscreenDefault(fullscreen)
  }

  function toggleMinimize(): void {
    updateCompose(session.id, { minimized: !session.minimized })
  }

  // applyFormat wraps the selection (or inserts a placeholder) with the markdown
  // for the requested action, then restores focus and the caret.
  async function applyFormat(action: string): Promise<void> {
    if (!bodyEl) {
      return
    }
    const start = bodyEl.selectionStart
    const end = bodyEl.selectionEnd
    const val = session.body
    const sel = val.slice(start, end)
    let next = val
    let caret = end

    const wrap = (token: string, placeholder: string): void => {
      const inner = sel || placeholder
      next = val.slice(0, start) + token + inner + token + val.slice(end)
      caret = start + token.length + inner.length + token.length
    }
    const linePrefix = (prefix: string): void => {
      const lineStart = val.lastIndexOf('\n', start - 1) + 1
      const block = val.slice(lineStart, end) || prefix
      const prefixed = block.split('\n').map((l) => prefix + l).join('\n')
      next = val.slice(0, lineStart) + prefixed + val.slice(end)
      caret = lineStart + prefixed.length
    }

    switch (action) {
      case 'bold':
        wrap('**', 'bold text')
        break
      case 'italic':
        wrap('*', 'italic text')
        break
      case 'code':
        wrap('`', 'code')
        break
      case 'link': {
        const text = sel || 'link'
        next = val.slice(0, start) + `[${text}](https://)` + val.slice(end)
        caret = start + text.length + 3
        break
      }
      case 'heading':
        linePrefix('## ')
        break
      case 'list':
        linePrefix('- ')
        break
      case 'quote':
        linePrefix('> ')
        break
    }

    updateCompose(session.id, { body: next })
    await tick()
    bodyEl.focus()
    bodyEl.setSelectionRange(caret, caret)
  }

  async function send(): Promise<void> {
    if (!hasRecipients(session)) {
      toastError('Add at least one recipient before sending.')
      return
    }
    sending = true
    try {
      const snapshot: ComposeSession = { ...session }
      const id = await sendMessage(buildRequest(session))
      if (session.draftId) {
        await deleteDraft(session.draftId)
      }
      await loadOutbox()
      closeCompose(session.id)

      const delay = $prefs.sendDelaySeconds
      if (delay > 0) {
        scheduleUndo(id, snapshot, delay)
      } else {
        toastSuccess('Message queued for sending.')
      }
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      sending = false
    }
  }

  async function save(): Promise<void> {
    try {
      const id = await saveDraft(session.draftId, buildRequest(session))
      updateCompose(session.id, { draftId: id })
      toastSuccess('Draft saved.')
    } catch (err) {
      toastError(errorMessage(err))
    }
  }
</script>

<div class="compose" class:fullscreen={session.fullscreen} class:minimized={session.minimized} role="dialog" aria-label="Compose message">
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <header class="head" on:dblclick={toggleMinimize}>
    <span class="title">{session.subject || 'Compose'}</span>
    <div class="win">
      <button type="button" class="win-btn" aria-label="Minimize" title="Minimize" on:click={toggleMinimize}>
        <IconMinus size={15} stroke={1.8} />
      </button>
      <button type="button" class="win-btn" aria-label={session.fullscreen ? 'Restore' : 'Fullscreen'} title={session.fullscreen ? 'Restore' : 'Fullscreen'} on:click={toggleFullscreen}>
        {#if session.fullscreen}
          <IconArrowsDiagonalMinimize2 size={15} stroke={1.8} />
        {:else}
          <IconArrowsDiagonal size={15} stroke={1.8} />
        {/if}
      </button>
      <button type="button" class="win-btn" aria-label="Close compose" on:click={() => closeCompose(session.id)}>
        <IconX size={16} stroke={1.8} />
      </button>
    </div>
  </header>

  {#if !session.minimized}
    {#if accounts.length > 1}
      <div class="from">
        <label for={`from-${session.id}`}>From</label>
        <select
          id={`from-${session.id}`}
          value={session.accountId}
          on:change={(e) => updateCompose(session.id, { accountId: Number(e.currentTarget.value) })}
        >
          {#each accounts as acc (acc.id)}
            <option value={acc.id}>{acc.email}</option>
          {/each}
        </select>
      </div>
    {/if}

    <AddressFields {session} />

    <div class="editor">
      {#if session.mode === 'wysiwyg'}
        <!-- the rich editor loads lazily; show a hint while the chunk arrives. -->
        {#if RichEditor}
          <svelte:component
            this={RichEditor}
            content={session.body}
            on:change={(e) => updateCompose(session.id, { body: e.detail })}
          />
        {:else}
          <div class="editor-loading">Loading editor…</div>
        {/if}
      {:else if session.mode === 'markdown'}
        <EditorToolbar {preview} on:format={(e) => applyFormat(e.detail)} on:togglePreview={() => (preview = !preview)} />
        {#if preview}
          <div class="preview selectable">{@html previewHtml}</div>
        {:else}
          <textarea
            class="body selectable"
            bind:this={bodyEl}
            aria-label="Message body"
            placeholder="Write markdown…"
            value={session.body}
            on:input={(e) => updateCompose(session.id, { body: e.currentTarget.value })}
          ></textarea>
        {/if}
      {:else}
        <textarea
          class="body mono selectable"
          aria-label="Message body"
          placeholder="Write your message…"
          value={session.body}
          on:input={(e) => updateCompose(session.id, { body: e.currentTarget.value })}
        ></textarea>
      {/if}
    </div>

    <div class="attach-row">
      <AttachmentPicker {session} />
    </div>

    <footer class="foot">
      <button type="button" class="send" disabled={sending} on:click={send}>
        <IconSend size={15} stroke={1.7} />
        {sending ? 'Sending…' : 'Send'}
      </button>
      <button type="button" class="save" on:click={save}>
        <IconDeviceFloppy size={15} stroke={1.6} />
        Save draft
      </button>
      <span class="spacer"></span>
      {#if $signatures.length > 0}
        <select class="sig-select" aria-label="Insert signature" on:change={insertSignature}>
          <option value="" disabled selected>Signature…</option>
          {#each $signatures as sig (sig.id)}
            <option value={sig.id}>{sig.name}</option>
          {/each}
        </select>
      {/if}
      <EditorModeSwitch mode={session.mode} on:change={setMode} />
    </footer>
  {/if}
</div>

<style>
  .compose {
    display: flex;
    flex-direction: column;
    width: 460px;
    height: 540px;
    max-height: calc(100vh - 40px);
    background: var(--surface-overlay);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    box-shadow: var(--shadow-overlay);
    overflow: hidden;
  }

  /* fullscreen breaks the pane out of the corner stack to fill the window. */
  .compose.fullscreen {
    position: fixed;
    inset: 24px;
    width: auto;
    height: auto;
    max-height: none;
    z-index: 130;
  }

  /* minimized collapses to just the title bar. */
  .compose.minimized {
    height: auto;
  }

  .head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-2) var(--space-2) var(--space-4);
    background: var(--surface-sunken);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .compose.minimized .head {
    border-bottom: none;
  }

  .title {
    font-size: var(--fz-label);
    font-weight: var(--fw-semibold);
    color: var(--text-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  .win {
    display: inline-flex;
    gap: 2px;
    flex-shrink: 0;
  }

  .win-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    border-radius: var(--radius-control);
  }

  .win-btn:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .from,
  .attach-row {
    padding: var(--space-2) var(--space-4);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .from {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }

  .from label {
    width: 52px;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }

  .from select {
    flex: 1;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    padding: var(--space-1) var(--space-2);
  }

  :global(.compose .fields) {
    padding: 0 var(--space-4);
  }

  .editor {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .body,
  .preview {
    flex: 1;
    width: 100%;
    min-height: 0;
    border: none;
    outline: none;
    resize: none;
    background: transparent;
    color: var(--text-primary);
    font-size: var(--fz-body);
    line-height: 1.55;
    overflow-y: auto;
    padding: var(--space-3) var(--space-4);
  }

  /* shown briefly while the rich editor chunk loads. */
  .editor-loading {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-tertiary);
    font-size: var(--fz-label);
  }

  .body.mono {
    font-family: var(--font-mono);
  }

  /* a light github-style rendered preview. */
  .preview :global(h1),
  .preview :global(h2),
  .preview :global(h3) {
    margin: var(--space-3) 0 var(--space-2);
  }

  .preview :global(a) {
    color: var(--link);
  }

  .preview :global(pre),
  .preview :global(code) {
    font-family: var(--font-mono);
    background: var(--surface-sunken);
    border-radius: var(--radius-control);
  }

  .preview :global(pre) {
    padding: var(--space-3);
    overflow-x: auto;
  }

  .preview :global(blockquote) {
    margin: 0 0 var(--space-2);
    padding-left: var(--space-3);
    border-left: 2px solid var(--border-strong);
    color: var(--text-secondary);
  }

  .foot {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-3) var(--space-4);
    border-top: var(--hairline) solid var(--border-subtle);
    background: var(--surface-sunken);
  }

  .spacer {
    flex: 1;
  }

  .sig-select {
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-secondary);
    font-size: var(--fz-label);
    padding: var(--space-1) var(--space-2);
    cursor: pointer;
  }

  .send,
  .save {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    cursor: pointer;
    font-size: var(--fz-label);
  }

  .send:hover:not(:disabled),
  .save:hover {
    background: var(--surface-hover);
  }

  .send:disabled {
    opacity: 0.6;
    cursor: default;
  }
</style>
