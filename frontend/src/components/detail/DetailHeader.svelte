<script lang="ts">
  // the header block of the reading pane: subject, sender with avatar, recipients,
  // full date and the same technical-info badges the list rows show.
  import Avatar from '../common/Avatar.svelte'
  import TechBadges from '../common/TechBadges.svelte'
  import { prefs } from '../../stores/prefs'
  import { formatFullDate, displayName, type TimeFormat } from '../../lib/format'
  import { t } from '../../lib/i18n'
  import type { MessageDetail } from '../../lib/types'

  export let detail: MessageDetail
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

  .date {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    flex-shrink: 0;
    white-space: nowrap;
  }
</style>
