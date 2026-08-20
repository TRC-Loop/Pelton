<script lang="ts">
  // the banner above a message the local checks found something wrong with
  // (#206). It states what was observed and never accuses the sender outright:
  // "this message was not sent by X" and "this link's text does not match its
  // target" are very different claims, and the wording keeps them apart.
  //
  // Nothing is hidden or binned on the strength of it. The reader decides.
  import { IconAlertTriangle, IconInfoCircle, IconChevronDown } from '@tabler/icons-svelte'
  import { t } from '../../lib/i18n'
  import type { PhishingReport, PhishingSignal } from '../../lib/types'

  export let report: PhishingReport
  // the message this report belongs to, so opening another one collapses the
  // details again rather than carrying the previous message's state over.
  export let messageId: number

  let expanded = false
  let shownFor = -1
  $: if (messageId !== shownFor) {
    shownFor = messageId
    expanded = false
  }

  $: signals = report.signals ?? []
  $: warning = report.level === 'warning'

  // headline says how strongly this reads, and each signal below says what was
  // actually seen. A signal with no explanation of its own falls back to its
  // kind so an unknown one from a newer backend still shows something.
  function explain(signal: PhishingSignal, tFn: (key: string) => string): string {
    const text = tFn(`phishing.signal.${signal.kind}`)
    if (text === `phishing.signal.${signal.kind}`) {
      return signal.detail ?? signal.kind
    }
    return signal.detail ? text.replace('{detail}', signal.detail) : text
  }
</script>

<div class="notice" class:warning role={warning ? 'alert' : 'status'}>
  <div class="head">
    <span class="icon">
      {#if warning}
        <IconAlertTriangle size={16} stroke={1.8} />
      {:else}
        <IconInfoCircle size={16} stroke={1.8} />
      {/if}
    </span>
    <span class="text">
      {warning ? $t('phishing.warning') : $t('phishing.caution')}
    </span>
    {#if signals.length > 0}
      <button type="button" class="toggle" class:open={expanded} on:click={() => (expanded = !expanded)}>
        {expanded ? $t('phishing.hideDetails') : $t('phishing.showDetails')}
        <IconChevronDown size={13} stroke={1.8} />
      </button>
    {/if}
  </div>

  {#if expanded}
    <ul class="signals">
      {#each signals as signal (signal.kind + (signal.detail ?? '') + (signal.target ?? ''))}
        <li>{explain(signal, $t)}</li>
      {/each}
    </ul>
    <p class="footnote">{$t('phishing.footnote')}</p>
  {/if}
</div>

<style>
  .notice {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin: 0 0 var(--space-3);
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-sunken);
    font-size: var(--fz-label);
    color: var(--text-primary);
  }

  /* the stronger reading gets the danger tone; caution stays neutral, because a
     banner that shouts at every newsletter is one the reader learns to skip. */
  .notice.warning {
    border-color: var(--danger);
    color: var(--danger);
  }

  .head {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .icon {
    display: inline-flex;
    flex-shrink: 0;
  }

  .text {
    flex: 1;
    min-width: 0;
  }

  .toggle {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    flex-shrink: 0;
    padding: var(--space-1) var(--space-2);
    border: none;
    border-radius: var(--radius-control);
    background: transparent;
    color: inherit;
    font-size: var(--fz-meta);
    cursor: var(--cursor-action);
  }

  .toggle:hover {
    background: var(--surface-hover);
  }

  .toggle.open :global(svg) {
    transform: rotate(180deg);
  }

  .signals {
    margin: 0;
    padding-left: var(--space-5);
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
    overflow-wrap: anywhere;
  }

  .footnote {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }
</style>
