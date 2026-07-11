<script lang="ts">
  // column 3: the reading pane. it loads the open message, shows the header,
  // action toolbar, sanitized body and attachments, and performs the toolbar
  // actions. empty, loading and error states are explicit.
  import peltonLogo from '../../assets/images/icons/pelton-logo.png'
  import DetailHeader from './DetailHeader.svelte'
  import ActionToolbar from './ActionToolbar.svelte'
  import MailBody from './MailBody.svelte'
  import AttachmentList from './AttachmentList.svelte'
  import InfoModal from './InfoModal.svelte'
  import Spinner from '../common/Spinner.svelte'
  import ErrorState from '../common/ErrorState.svelte'
  import { openMessageId } from '../../stores/selection'
  import { messageDetail, loadMessage, clearMessage } from '../../stores/message'
  import { setFlagged, deleteMessage, archiveMessage } from '../../lib/api'
  import { removeFromList, patchInList } from '../../stores/messages'
  import { recordDeleted } from '../../stores/undodelete'
  import { recordArchived } from '../../stores/undoarchive'
  import { openReply, openForward } from '../../stores/compose'
  import { errorMessage, toastError } from '../../stores/toast'
  import { prefs } from '../../stores/prefs'
  import { t } from '../../lib/i18n'
  import { get } from 'svelte/store'
  import type { MessageDetail, EditorMode } from '../../lib/types'

  // default editor mode for replies and forwards, from settings.
  $: replyMode = $prefs.defaultEditorMode as EditorMode

  let infoOpen = false

  let loadedId = -1
  $: if ($openMessageId !== null && $openMessageId !== loadedId) {
    loadedId = $openMessageId
    void loadMessage($openMessageId)
  }
  $: if ($openMessageId === null && loadedId !== -1) {
    loadedId = -1
    clearMessage()
  }

  async function toggleFlag(detail: MessageDetail): Promise<void> {
    const next = !detail.flagged
    messageDetail.update((s) => (s.data ? { ...s, data: { ...s.data, flagged: next } } : s))
    patchInList(detail.id, { flagged: next })
    try {
      await setFlagged(detail.id, next)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function remove(detail: MessageDetail): Promise<void> {
    try {
      await deleteMessage(detail.id)
      recordDeleted(detail)
      removeFromList(detail.id)
      openMessageId.set(null)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // archive moves the message to the account's Archive folder via the imap move
  // binding, then closes the pane and drops it from the list.
  async function archive(detail: MessageDetail): Promise<void> {
    try {
      const undo = await archiveMessage(detail.id)
      if (undo.messageId) {
        recordArchived(detail, undo.messageId, undo.originalFolderId)
      }
      removeFromList(detail.id)
      openMessageId.set(null)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // print builds a clean, self-contained document from the already-sanitized
  // body and opens the system print dialog via a hidden iframe (WKWebView and
  // WebView2 both route window.print() to the native dialog). using the
  // sanitized body keeps remote/script content out of the printed page.
  function escapeHtml(s: string): string {
    return s.replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c] ?? c)
  }

  function printMessage(detail: MessageDetail): void {
    const noSubject = get(t)('detail.noSubject')
    const toLabel = get(t)('detail.print.to')
    const bodyHtml = detail.isHtml
      ? detail.bodyHtmlSafe
      : `<pre style="white-space:pre-wrap;font:14px/1.5 ui-monospace,monospace">${escapeHtml(detail.bodyPlain)}</pre>`
    const doc = `<!doctype html><html><head><meta charset="utf-8"><title>${escapeHtml(detail.subject || noSubject)}</title>
<style>
  body{font:14px/1.6 system-ui,sans-serif;color:#111;margin:24px}
  .hdr{border-bottom:1px solid #ccc;padding-bottom:12px;margin-bottom:16px}
  .hdr h1{font-size:18px;margin:0 0 8px}
  .hdr div{font-size:12px;color:#555}
  img{max-width:100%}
</style></head><body>
<div class="hdr">
  <h1>${escapeHtml(detail.subject || noSubject)}</h1>
  <div><strong>${escapeHtml(detail.fromName || detail.fromAddress)}</strong> &lt;${escapeHtml(detail.fromAddress)}&gt;</div>
  <div>${escapeHtml(detail.date)}</div>
  <div>${escapeHtml(toLabel)} ${escapeHtml(detail.toAddresses)}</div>
</div>
${bodyHtml}
</body></html>`

    const frame = document.createElement('iframe')
    frame.setAttribute('aria-hidden', 'true')
    frame.style.position = 'fixed'
    frame.style.right = '0'
    frame.style.bottom = '0'
    frame.style.width = '0'
    frame.style.height = '0'
    frame.style.border = '0'
    document.body.appendChild(frame)
    const win = frame.contentWindow
    const idoc = frame.contentDocument || win?.document
    if (!win || !idoc) {
      document.body.removeChild(frame)
      toastError(get(t)('detail.print.dialogFailed'))
      return
    }
    idoc.open()
    idoc.write(doc)
    idoc.close()
    // give the iframe a tick to lay out before printing, then clean up after.
    const run = (): void => {
      win.focus()
      win.print()
      setTimeout(() => frame.remove(), 1000)
    }
    if (idoc.readyState === 'complete') {
      setTimeout(run, 50)
    } else {
      frame.onload = run
    }
  }
</script>

<section class="detail">
  {#if $openMessageId === null}
    <div class="placeholder">
      <img class="placeholder-logo" src={$prefs.emptyStateImage || peltonLogo} alt="Pelton" draggable="false" />
    </div>
  {:else if $messageDetail.status === 'loading' && !$messageDetail.data}
    <Spinner label={$t('detail.loadingMessage')} />
  {:else if $messageDetail.status === 'error'}
    <ErrorState message={$messageDetail.error} onRetry={() => $openMessageId && loadMessage($openMessageId)} />
  {:else if $messageDetail.data}
    {@const detail = $messageDetail.data}
    <div class="toolbar-bar">
      <ActionToolbar
        flagged={detail.flagged}
        on:reply={() => openReply(detail, replyMode, false)}
        on:replyAll={() => openReply(detail, replyMode, true)}
        on:forward={() => openForward(detail, replyMode)}
        on:archive={() => archive(detail)}
        on:delete={() => remove(detail)}
        on:toggleFlag={() => toggleFlag(detail)}
        on:print={() => printMessage(detail)}
        on:info={() => (infoOpen = true)}
      />
    </div>

    {#if infoOpen}
      <InfoModal {detail} on:close={() => (infoOpen = false)} />
    {/if}

    <div class="scroll selectable">
      <DetailHeader {detail} />
      <div class="body-wrap">
        <MailBody {detail} />
      </div>
      <AttachmentList messageId={detail.id} attachments={detail.attachments} />
    </div>
  {/if}
</section>

<style>
  .detail {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--surface-raised);
    min-width: 0;
  }

  /* with no message selected the pane shows the Pelton mark, faded, as a calm
     empty state rather than text. */
  .placeholder {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 0;
  }

  .placeholder-logo {
    width: 128px;
    height: 128px;
    object-fit: contain;
    opacity: 0.12;
    filter: grayscale(1);
    user-select: none;
    -webkit-user-select: none;
  }

  .toolbar-bar {
    display: flex;
    align-items: center;
    padding: var(--space-2) var(--pane-pad);
    border-bottom: var(--hairline) solid var(--border-subtle);
    flex-shrink: 0;
  }

  .scroll {
    flex: 1;
    /* min-height:0 is required for a flex child to scroll; without it the body
       grows the pane instead of scrolling inside it. */
    min-height: 0;
    overflow-y: auto;
    padding: var(--pane-pad);
    display: flex;
    flex-direction: column;
  }

  /* the html body iframe now sizes itself to its content (single scrollbar in the
     outer .scroll), so the wrap hugs it instead of stretching and leaving a gap. */
  .body-wrap {
    flex: 0 0 auto;
    margin-top: var(--space-4);
    display: flex;
    flex-direction: column;
  }
</style>
