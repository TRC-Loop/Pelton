<script lang="ts">
  // renders the message body. html mail goes into a sandboxed iframe with no
  // script execution and no same-origin access, as defense in depth on top of the
  // backend sanitization. plaintext renders in the mono font. remote images are
  // blocked by the backend by default; a per-message affordance asks the backend
  // to re-render with remote content allowed.
  import { IconPhoto, IconUserCheck, IconWorldCheck } from '@tabler/icons-svelte'
  import { prefs } from '../../stores/prefs'
  import { getMessageHtml, trustSenderImages, allowDomainImages } from '../../lib/api'
  import { setBodyHtml } from '../../stores/message'
  import { errorMessage, toastError, toastSuccess } from '../../stores/toast'
  import { displayName } from '../../lib/format'
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
      toastSuccess(`Images from ${senderLabel} will load from now on.`)
      await loadRemote()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // trust the whole sender domain permanently, then show remote content now.
  async function trustDomain(): Promise<void> {
    try {
      await allowDomainImages(detail.id)
      toastSuccess(`Images from ${senderDomain} will load from now on.`)
      await loadRemote()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }
</script>

{#if detail.hasRemoteContent && !remoteLoaded}
  <div class="remote-bar">
    <div class="remote-info">
      <span class="remote-text">Remote images are blocked to protect your privacy.</span>
      {#if detail.remoteHosts && detail.remoteHosts.length > 0}
        <span class="remote-hosts" title={detail.remoteHosts.join(', ')}>
          from {detail.remoteHosts.slice(0, 3).join(', ')}{detail.remoteHosts.length > 3 ? ` +${detail.remoteHosts.length - 3} more` : ''}
        </span>
      {/if}
    </div>
    <div class="remote-actions">
      <button type="button" class="remote-btn" on:click={loadRemote}>
        <IconPhoto size={14} stroke={1.6} />
        Load once
      </button>
      <button type="button" class="remote-btn" on:click={trustSender} title={`Always load images from ${senderLabel}`}>
        <IconUserCheck size={14} stroke={1.6} />
        This sender
      </button>
      {#if senderDomain}
        <button type="button" class="remote-btn" on:click={trustDomain} title={`Always load images from ${senderDomain}`}>
          <IconWorldCheck size={14} stroke={1.6} />
          This domain
        </button>
      {/if}
    </div>
  </div>
{/if}

{#if detail.isHtml}
  <iframe class="body-frame" title="Message content" sandbox="" {srcdoc}></iframe>
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

  .body-frame {
    width: 100%;
    height: 100%;
    border: none;
    background: var(--surface-raised);
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
