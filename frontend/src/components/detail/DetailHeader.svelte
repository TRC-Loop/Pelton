<script lang="ts">
  // the header block of the reading pane: subject, sender with avatar, recipients,
  // full date, the technical-info badges the list rows show, and the
  // unsubscribe button when the message advertises a mechanism (or, failing
  // that, contains an unsubscribe link in its body).
  import { IconMailOff, IconCheck } from '@tabler/icons-svelte'
  import Avatar from '../common/Avatar.svelte'
  import TechBadges from '../common/TechBadges.svelte'
  import { prefs } from '../../stores/prefs'
  import { formatFullDate, displayName, type TimeFormat } from '../../lib/format'
  import { unsubscribeMessage } from '../../lib/api'
  import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'
  import { errorMessage, toastError, toastSuccess } from '../../stores/toast'
  import { t } from '../../lib/i18n'
  import type { MessageDetail, UnsubscribeInfo } from '../../lib/types'

  export let detail: MessageDetail

  // unsub resolves the mechanism: the stored List-Unsubscribe headers first,
  // otherwise an unsubscribe-looking link scraped from the (already
  // sanitized) body as a browser-link fallback.
  $: unsub = detail.unsubscribe ?? bodyUnsubscribeLink(detail.bodyHtmlSafe)
  $: done = detail.unsubscribe?.done ?? false

  // a mis-click must not silently POST anywhere: the first click arms the
  // button, the second within the window carries it out.
  let confirming = false
  let confirmTimer = 0
  let working = false

  // reset the confirm state when another message opens.
  $: if (detail.id) {
    confirming = false
    working = false
  }

  const unsubWords = /unsubscrib|abmelden|abbestellen|d[eé]sabonn|desuscrib|afmeld/i

  function bodyUnsubscribeLink(html: string): UnsubscribeInfo | null {
    if (!html) {
      return null
    }
    const doc = new DOMParser().parseFromString(html, 'text/html')
    for (const a of Array.from(doc.querySelectorAll('a[href]'))) {
      const href = a.getAttribute('href') ?? ''
      if (!/^https?:/i.test(href)) {
        continue
      }
      if (unsubWords.test(href) || unsubWords.test(a.textContent ?? '')) {
        return { kind: 'link', target: href, done: false }
      }
    }
    return null
  }

  async function onUnsubscribe(): Promise<void> {
    if (!unsub || working || done) {
      return
    }
    if (!confirming) {
      confirming = true
      clearTimeout(confirmTimer)
      confirmTimer = window.setTimeout(() => (confirming = false), 5000)
      return
    }
    clearTimeout(confirmTimer)
    confirming = false
    if (unsub.kind === 'link') {
      BrowserOpenURL(unsub.target)
      return
    }
    working = true
    try {
      await unsubscribeMessage(detail.id)
      done = true
      toastSuccess($t('detail.unsubscribe.done'))
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      working = false
    }
  }
</script>

<header class="head">
  <h1 class="subject">{detail.subject || $t('detail.noSubject')}</h1>

  <div class="from-row">
    <Avatar name={detail.fromName} email={detail.fromAddress} size={36} />
    <div class="from-info">
      <div class="from-line">
        <span class="from-name">{displayName(detail.fromName, detail.fromAddress)}</span>
        {#if detail.fromName}
          <span class="from-addr">&lt;{detail.fromAddress}&gt;</span>
        {/if}
      </div>
      <div class="recipients">
        {#if detail.toAddresses}<span>{$t('detail.header.to')} {detail.toAddresses}</span>{/if}
        {#if detail.ccAddresses}<span class="cc">{$t('detail.header.cc')} {detail.ccAddresses}</span>{/if}
      </div>
    </div>
    {#if $prefs.showDateTime}
      <time class="date">{formatFullDate(detail.date, $prefs.timeFormat as TimeFormat)}</time>
    {/if}
  </div>

  <div class="badges-row">
    <TechBadges
      accountEmail={detail.accountEmail}
      folderName={detail.folderName}
      pgp={detail.pgp}
      auth={detail.auth}
    />
    {#if unsub}
      <button type="button" class="unsub" class:confirming disabled={done || working} on:click={onUnsubscribe}>
        {#if done}
          <IconCheck size={13} stroke={2} />
          {$t('detail.unsubscribe.doneLabel')}
        {:else if confirming}
          {unsub.kind === 'link' ? $t('detail.unsubscribe.confirmOpen') : $t('detail.unsubscribe.confirm')}
        {:else}
          <IconMailOff size={13} stroke={1.8} />
          {$t('detail.unsubscribe.button')}
        {/if}
      </button>
    {/if}
  </div>
</header>

<style>
  .head {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding-bottom: var(--space-4);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .subject {
    margin: 0;
    font-size: var(--fz-title);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
    line-height: 1.3;
    user-select: text;
  }

  .from-row {
    display: flex;
    align-items: flex-start;
    gap: var(--space-3);
  }

  .from-info {
    flex: 1;
    min-width: 0;
  }

  .from-line {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    flex-wrap: wrap;
  }

  .from-name {
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .from-addr {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }

  .recipients {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
    margin-top: var(--space-1);
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }

  .recipients span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 48ch;
  }

  .badges-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex-wrap: wrap;
  }

  .unsub {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: 2px var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--text-secondary);
    font-size: var(--fz-meta);
    cursor: pointer;
  }

  .unsub:hover:not(:disabled) {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .unsub.confirming {
    border-color: var(--warning);
    color: var(--warning);
  }

  .unsub:disabled {
    color: var(--text-tertiary);
    cursor: default;
  }

  .date {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    flex-shrink: 0;
    white-space: nowrap;
  }
</style>
