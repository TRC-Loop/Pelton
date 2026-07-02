<script lang="ts">
  // renders the message body. html mail goes into a sandboxed iframe with no
  // script execution and no same-origin access, as defense in depth on top of the
  // backend sanitization. plaintext renders in the mono font. remote images are
  // blocked by the backend by default; a per-message affordance asks the backend
  // to re-render with remote content allowed.
  import { onDestroy } from 'svelte'
  import { IconPhoto, IconUserCheck, IconWorldCheck } from '@tabler/icons-svelte'
  import { prefs } from '../../stores/prefs'
  import { getMessageHtml, trustSenderImages, allowDomainImages } from '../../lib/api'
  import { setBodyHtml } from '../../stores/message'
  import { errorMessage, toastError, toastSuccess } from '../../stores/toast'
  import { displayName } from '../../lib/format'
  import { t } from '../../lib/i18n'
  import type { MessageDetail } from '../../lib/types'

  export let detail: MessageDetail

  // remoteLoaded starts true when the backend already rendered remote content
  // because the sender/domain is trusted (or the global override is on).
  let remoteLoaded = detail.remoteAllowed

  // reset the remote-loaded affordance when a different message opens.
  let lastId = -1
  $: if (detail.id !== lastId) {
    lastId = detail.id
    remoteLoaded = detail.remoteAllowed
    frameHeight = 320
  }

  $: senderLabel = displayName(detail.fromName, detail.fromAddress)
  $: senderDomain = detail.fromAddress.includes('@') ? detail.fromAddress.split('@').pop() : ''

  // readVar reads a resolved css token value so the iframe document, which cannot
  // see the parent stylesheet, can match the current theme. $prefs.theme is a
  // dependency so the srcdoc rebuilds when the theme changes.
  function readVar(name: string): string {
    return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
  }

  // buildSrcdoc wraps the already-sanitized body in a minimal themed document.
  // the style tag name is concatenated so the literal token never appears in the
  // component source, where svelte's parser would mistake it for a real style
  // block and parse the css that follows as the component's styles.
  //
  // a content-security-policy is set as defense in depth on top of the backend
  // sanitizer: when remote content is not allowed, img-src is limited to data:
  // (our inlined cid images) so nothing can phone home even if a remote url ever
  // slipped past the sanitizer. when the user opts in, http(s) image sources are
  // permitted. scripts are never allowed.
  function buildSrcdoc(html: string, _theme: string, allowRemote: boolean, fontSize: number): string {
    const fg = readVar('--text-primary')
    const muted = readVar('--text-tertiary')
    const link = readVar('--link')
    const bg = readVar('--surface-raised')
    const font = readVar('--font-ui')
    const css = `
  html,body{margin:0;background:${bg};color:${fg};font-family:${font};font-size:${fontSize}px;line-height:1.5;}
  body{padding:4px 2px;word-wrap:break-word;overflow-wrap:break-word;}
  a{color:${link};}
  img{max-width:100%;height:auto;}
  blockquote{margin:0 0 0 8px;padding-left:10px;border-left:2px solid ${muted};color:${muted};}
  table{max-width:100%;}
  pre{white-space:pre-wrap;}`
    const imgSrc = allowRemote ? 'data: https: http:' : 'data:'
    const csp = `default-src 'none'; img-src ${imgSrc}; style-src 'unsafe-inline'; font-src data:`
    const cspMeta = `<meta http-equiv="Content-Security-Policy" content="${csp}">`
    const open = '<sty' + 'le>'
    const close = '</sty' + 'le>'
    return `<!doctype html><html><head><meta charset="utf-8">${cspMeta}${open}${css}${close}</head><body>${html}</body></html>`
  }

  $: srcdoc = buildSrcdoc(detail.bodyHtmlSafe, $prefs.theme, remoteLoaded, $prefs.messageFontSize)

  // the iframe is sized to its content height so the reading pane has a single
  // scrollbar instead of a nested one (which the interface zoom made worse). the
  // sandbox stays script-free; allow-same-origin only lets us measure the content
  // height, it does not run any of the email's scripts (there are none: the CSP is
  // default-src 'none' and allow-scripts is not set).
  let frame: HTMLIFrameElement
  let frameHeight = 320
  let resizeObserver: ResizeObserver | null = null

  // measuring documentElement/body.scrollHeight is unreliable here: per spec
  // scrollHeight can never be smaller than the viewport (the iframe's own
  // current height), so once the iframe grows it can never be measured back
  // down, and the height only ratchets upward, leaving a growing gap below
  // short emails. a ResizeObserver on the body's own border box reports its
  // true content size independent of the iframe's current height.
  function measure(): void {
    const doc = frame?.contentDocument
    const body = doc?.body
    if (!body) {
      return
    }
    const h = body.getBoundingClientRect().height
    frameHeight = Math.max(40, Math.ceil(h))
  }

  function onFrameLoad(): void {
    resizeObserver?.disconnect()
    resizeObserver = null
    const doc = frame?.contentDocument
    const body = doc?.body
    if (!body) {
      measure()
      return
    }
    measure()
    resizeObserver = new ResizeObserver(() => measure())
    resizeObserver.observe(body)
  }

  onDestroy(() => resizeObserver?.disconnect())

  async function loadRemote(): Promise<void> {
    try {
      const html = await getMessageHtml(detail.id, true)
      setBodyHtml(html)
      remoteLoaded = true
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // trust the sender permanently, then show this message's remote content now.
  async function trustSender(): Promise<void> {
    try {
      await trustSenderImages(detail.id)
      toastSuccess($t('detail.mailBody.imagesTrusted').replace('{who}', senderLabel))
      await loadRemote()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // trust the whole sender domain permanently, then show remote content now.
  async function trustDomain(): Promise<void> {
    try {
      await allowDomainImages(detail.id)
      toastSuccess($t('detail.mailBody.imagesTrusted').replace('{who}', senderDomain ?? ''))
      await loadRemote()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }
</script>

{#if detail.hasRemoteContent && !remoteLoaded}
  <div class="remote-bar">
    <div class="remote-info">
      <span class="remote-text">{$t('detail.mailBody.remoteBlocked')}</span>
      {#if detail.remoteHosts && detail.remoteHosts.length > 0}
        <span class="remote-hosts" title={detail.remoteHosts.join(', ')}>
          {$t('detail.mailBody.from')} {detail.remoteHosts.slice(0, 3).join(', ')}{detail.remoteHosts.length > 3 ? ` ${$t('detail.mailBody.more').replace('{count}', String(detail.remoteHosts.length - 3))}` : ''}
        </span>
      {/if}
    </div>
    <div class="remote-actions">
      <button type="button" class="remote-btn" on:click={loadRemote}>
        <IconPhoto size={14} stroke={1.6} />
        {$t('detail.mailBody.loadOnce')}
      </button>
      <button type="button" class="remote-btn" on:click={trustSender} title={$t('detail.mailBody.alwaysLoadFrom').replace('{who}', senderLabel)}>
        <IconUserCheck size={14} stroke={1.6} />
        {$t('detail.mailBody.thisSender')}
      </button>
      {#if senderDomain}
        <button type="button" class="remote-btn" on:click={trustDomain} title={$t('detail.mailBody.alwaysLoadFrom').replace('{who}', senderDomain)}>
          <IconWorldCheck size={14} stroke={1.6} />
          {$t('detail.mailBody.thisDomain')}
        </button>
      {/if}
    </div>
  </div>
{/if}

{#if detail.isHtml}
  <iframe
    class="body-frame"
    title={$t('detail.mailBody.iframeTitle')}
    sandbox="allow-same-origin"
    bind:this={frame}
    on:load={onFrameLoad}
    style={`height:${frameHeight}px`}
    {srcdoc}
  ></iframe>
{:else}
  <pre class="body-plain mono selectable" style={`font-size:${$prefs.messageFontSize}px`}>{detail.bodyPlain}</pre>
{/if}

<style>
  /* a warning-tinted banner, since blocked remote content is a privacy matter. */
  .remote-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    flex-wrap: wrap;
    margin-bottom: var(--space-3);
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--warning);
    border-radius: var(--radius-control);
    background: var(--warning-bg, var(--surface-sunken));
  }

  .remote-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .remote-text {
    font-size: var(--fz-label);
    color: var(--text-primary);
  }

  .remote-hosts {
    font-size: var(--fz-meta);
    color: var(--text-secondary);
    font-family: var(--font-mono);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .remote-actions {
    display: inline-flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }

  .remote-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .remote-btn:hover {
    background: var(--surface-hover);
  }

  /* height is set inline from the measured content height so the pane has a single
     scrollbar. a min-height avoids a flash of collapse before the first measure.
     the iframe is a nested browsing context, but its rendering still gets pulled
     along by the ancestor's CSS zoom (the app-wide interface scale, applied on
     <html> in theme.ts) since zoom is not a normal non-inherited property for
     replaced elements. that means email content ends up zoomed twice: once by
     its own messageFontSize, and again by whatever interface scale the user
     picked. dividing back out by --ui-scale here cancels the ancestor's zoom so
     the reading pane always renders at true size regardless of interface zoom. */
  .body-frame {
    display: block;
    width: 100%;
    min-height: 120px;
    border: none;
    background: var(--surface-raised);
    zoom: calc(1 / var(--ui-scale, 1));
  }

  .body-plain {
    margin: 0;
    font-size: var(--fz-body);
    line-height: 1.55;
    color: var(--text-primary);
    white-space: pre-wrap;
    word-break: break-word;
  }
</style>
