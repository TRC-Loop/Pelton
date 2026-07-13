<script lang="ts">
  // renders the message body. html mail goes into a sandboxed iframe, as
  // defense in depth on top of the backend sanitization: a strict CSP allows
  // exactly one nonce-scoped inline script (Pelton's own click handler, built
  // fresh per render - see buildSrcdoc), so nothing from the email itself can
  // ever execute even though the sandbox now permits scripts to run at all.
  // plaintext renders in the mono font, with bare urls linkified. remote
  // images are blocked by the backend by default; a per-message affordance
  // asks the backend to re-render with remote content allowed.
  import { onDestroy } from 'svelte'
  import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'
  import { IconPhoto, IconUserCheck, IconWorldCheck } from '@tabler/icons-svelte'
  import { prefs } from '../../stores/prefs'
  import { getMessageHtml, trustSenderImages, allowDomainImages } from '../../lib/api'
  import { setBodyHtml } from '../../stores/message'
  import { errorMessage, toastError, toastSuccess } from '../../stores/toast'
  import { displayName, linkifySegments } from '../../lib/format'
  import { bodyFontStack } from '../../lib/fonts'
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

  // buildSrcdoc wraps the already-sanitized body in a minimal document. the
  // style tag name is concatenated so the literal token never appears in the
  // component source, where svelte's parser would mistake it for a real style
  // block and parse the css that follows as the component's styles.
  //
  // the background/text colors here are fixed light values, not theme
  // tokens, and deliberately do not follow dark mode. html mail almost always
  // assumes it is being rendered on a white page and sets its own colors (or
  // none at all) on that assumption - senders that hardcode dark/black text
  // with no background of their own used to inherit our dark-theme background,
  // producing unreadable black-on-near-black text. every other mail client
  // (Thunderbird, Apple Mail, Gmail) renders html mail on a fixed light
  // background for the same reason; a message's own styles still override
  // this when it sets them explicitly.
  //
  // a content-security-policy is set as defense in depth on top of the backend
  // sanitizer: when remote content is not allowed, img-src is limited to data:
  // (our inlined cid images) so nothing can phone home even if a remote url ever
  // slipped past the sanitizer. when the user opts in, http(s) image sources are
  // permitted.
  //
  // script-src is scoped to a single nonce, generated fresh per render, and
  // that nonce is used on exactly one inline <script> block below (Pelton's
  // own, never sender content). Any <script> tag the sanitizer might have let
  // through from the email itself has no nonce and so still can't execute:
  // the sandbox no longer omits allow-scripts, but the CSP keeps the "sender
  // html can never run code" guarantee. allow-scripts had to come back
  // because clicks inside a script-less sandboxed iframe never reached a
  // cross-frame listener registered from the parent on some webview engines
  // (contentDocument access itself worked - measuring the body's height was
  // fine - but no click ever arrived at a listener attached that way). A
  // click handler that runs natively inside the iframe's own document, using
  // postMessage to hand the url back to the parent, has no such dependency on
  // cross-frame event delivery.
  function buildSrcdoc(html: string, allowRemote: boolean, fontSize: number, bodyFont: string, nonce: string): string {
    // the reader font preference only sets the fallback; mail that declares
    // its own fonts keeps them since this is just the base font-family.
    const font = bodyFontStack(bodyFont) ?? readVar('--font-ui')
    const css = `
  html,body{margin:0;background:#ffffff;color:#1a1a1a;font-family:${font};font-size:${fontSize}px;line-height:1.5;}
  body{padding:4px 2px;word-wrap:break-word;overflow-wrap:break-word;}
  a{color:#1a56db;}
  img{max-width:100%;height:auto;}
  blockquote{margin:0 0 0 8px;padding-left:10px;border-left:2px solid #94a3b8;color:#55606c;}
  table{max-width:100%;}
  pre{white-space:pre-wrap;}`
    const imgSrc = allowRemote ? 'data: https: http:' : 'data:'
    const csp = `default-src 'none'; img-src ${imgSrc}; style-src 'unsafe-inline'; font-src data:; script-src 'nonce-${nonce}'`
    const cspMeta = `<meta http-equiv="Content-Security-Policy" content="${csp}">`
    const open = '<sty' + 'le>'
    const close = '</sty' + 'le>'
    // relays clicked links to the parent instead of trying to navigate this
    // sandboxed iframe (which has no allow-top-navigation and would silently
    // do nothing). runs inside the iframe's own document so it's a normal,
    // same-document click listener - no cross-frame event delivery involved.
    const script =
      '<scr' +
      'ipt nonce="' +
      nonce +
      '">document.addEventListener("click",function(e){var a=e.target&&e.target.closest("a");if(!a)return;var href=(a.getAttribute("href")||"").trim();if(!href)return;e.preventDefault();if(/^(https?:|mailto:)/i.test(href)){window.parent.postMessage({peltonOpenUrl:href},"*")}});</scr' +
      'ipt>'
    // data-pelton-ready marks the body as belonging to our own srcdoc, not
    // the iframe's initial blank placeholder document: a fresh iframe already
    // has an empty <body> before any srcdoc has loaded, so a readiness check
    // that only looks for "a body" would resolve instantly against that
    // placeholder instead of waiting for the real content.
    return `<!doctype html><html><head><meta charset="utf-8">${cspMeta}${open}${css}${close}</head><body data-pelton-ready="1">${html}${script}</body></html>`
  }

  // nonce is regenerated per message so a stale nonce from a previous render
  // can never be replayed against a new one.
  function makeNonce(): string {
    return crypto.randomUUID().replace(/-/g, '')
  }

  $: srcdoc = buildSrcdoc(detail.bodyHtmlSafe, remoteLoaded, $prefs.messageFontSize, $prefs.bodyFont, makeNonce())

  // plain-text bodies render in a <pre>, not the sandboxed iframe, so bare
  // urls need their own linkification: nothing upstream turns them into real
  // links the way html mail's own <a> tags already are.
  $: plainSegments = detail.isHtml ? [] : linkifySegments(detail.bodyPlain)

  // the iframe is sized to its content height so the reading pane has a single
  // scrollbar instead of a nested one (which the interface zoom made worse).
  // allow-same-origin lets us measure the content height from the parent.
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
    const body = readyBody()
    if (!body) {
      return
    }
    const h = body.getBoundingClientRect().height
    frameHeight = Math.max(40, Math.ceil(h))
  }

  // readyBody returns the iframe's body only once it's our own rendered
  // srcdoc, not the iframe's transient initial blank document.
  function readyBody(): HTMLElement | null {
    const body = frame?.contentDocument?.body
    return body?.dataset.peltonReady ? body : null
  }

  // readyPollHandle cancels the fallback readiness poll started by
  // scheduleAttach, so a later reload (a new message) doesn't leave a stale
  // poll running against a document that's already been replaced.
  let readyPollHandle = 0

  // attachInteractivity wires up sizing for the iframe's current document.
  // Link clicks are handled separately, by the nonce-scoped script inside the
  // srcdoc itself (see buildSrcdoc) relaying to onWindowMessage below - not by
  // a listener attached here from the parent, which proved unreliable. This
  // is idempotent per document: calling it twice for the same load is
  // harmless since a fresh resizeObserver just replaces the previous one.
  function attachInteractivity(): void {
    resizeObserver?.disconnect()
    resizeObserver = null
    const body = readyBody()
    if (!body) {
      return
    }
    measure()
    resizeObserver = new ResizeObserver(() => measure())
    resizeObserver.observe(body)
  }

  // onFrameLoad fires when the iframe's srcdoc finishes loading. It is the
  // fast path; scheduleAttach below also runs unconditionally as a fallback
  // in case 'load' fires before contentDocument.body is actually populated
  // (observed to be unreliable timing on some webview engines).
  function onFrameLoad(): void {
    attachInteractivity()
  }

  // scheduleAttach polls (briefly, via rAF) until the iframe's document has a
  // body, then attaches interactivity. This exists because relying solely on
  // the 'load' event has proven unreliable: a listener registered a beat too
  // late leaves clicks silently doing nothing, which is worse than the small
  // cost of polling for under a second.
  function scheduleAttach(): void {
    cancelAnimationFrame(readyPollHandle)
    let attempts = 0
    const tick = (): void => {
      if (readyBody()) {
        attachInteractivity()
        return
      }
      attempts += 1
      if (attempts < 120) {
        readyPollHandle = requestAnimationFrame(tick)
      }
    }
    readyPollHandle = requestAnimationFrame(tick)
  }

  $: if (frame && srcdoc) {
    scheduleAttach()
  }

  // onWindowMessage receives the url from the click handler running inside
  // the iframe's own document (see the injected script in buildSrcdoc) and
  // opens it in the external browser. The source check makes sure this only
  // ever acts on messages from our own mail iframe, not anything else that
  // might postMessage into this window.
  function onWindowMessage(event: MessageEvent): void {
    if (event.source !== frame?.contentWindow) {
      return
    }
    const href = (event.data as { peltonOpenUrl?: unknown } | null)?.peltonOpenUrl
    if (typeof href === 'string' && /^(https?:|mailto:)/i.test(href)) {
      BrowserOpenURL(href)
    }
  }

  window.addEventListener('message', onWindowMessage)

  onDestroy(() => {
    resizeObserver?.disconnect()
    cancelAnimationFrame(readyPollHandle)
    window.removeEventListener('message', onWindowMessage)
  })

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
    sandbox="allow-same-origin allow-scripts allow-popups allow-top-navigation-by-user-activation"
    bind:this={frame}
    on:load={onFrameLoad}
    style={`height:${frameHeight}px`}
    {srcdoc}
  ></iframe>
{:else}
  <pre class="body-plain mono selectable" style={`font-size:${$prefs.messageFontSize}px`}>{#each plainSegments as segment}{#if segment.href}<a
        class="plain-link"
        href={segment.href}
        on:click|preventDefault={() => BrowserOpenURL(segment.href ?? '')}
      >{segment.text}</a>{:else}{segment.text}{/if}{/each}</pre>
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
    /* fixed, not theme-derived: matches the fixed light background the
       srcdoc itself renders on (see buildSrcdoc), so there is no dark flash
       around/before the html mail content in dark mode. */
    background: #ffffff;
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

  .plain-link {
    color: var(--accent);
    text-decoration: underline;
    cursor: pointer;
  }
</style>
