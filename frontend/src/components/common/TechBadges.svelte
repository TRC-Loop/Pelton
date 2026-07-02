<script lang="ts">
  // the per-message technical-info badges shared by the list row and the detail
  // header: mailbox/account badge, pgp status and auth status. each badge is
  // shown only when its preference toggle is on. the auth badge is always a
  // neutral "not available" state because the backend does not parse
  // Authentication-Results yet (documented follow-up); we never invent a result.
  import { IconMailbox, IconLock, IconShieldCheck, IconShieldQuestion } from '@tabler/icons-svelte'
  import { prefs } from '../../stores/prefs'
  import { t } from '../../lib/i18n'
  import type { PGPStatus } from '../../lib/types'

  export let accountEmail: string = ''
  export let folderName: string = ''
  export let pgp: string = 'none'
  export let auth: string = 'unavailable'

  // pgp label and icon per status. "none" renders nothing.
  function pgpLabel(status: string, tFn: (key: string) => string): string {
    if (status === 'encrypted') return tFn('common.techBadges.encrypted')
    if (status === 'signed') return tFn('common.techBadges.signed')
    return ''
  }

  $: showBadge = $prefs.showMailboxBadge && (accountEmail !== '' || folderName !== '')
  $: showPgp = $prefs.showPgp && pgp !== 'none'
  $: showAuth = $prefs.showAuth
  $: pgpStatus = pgp as PGPStatus
  // auth has only the "unavailable" state today; show n/a until the backend
  // parses Authentication-Results, otherwise echo whatever it reports.
  $: authText = auth === 'unavailable' ? $t('common.techBadges.authNA') : auth
</script>

{#if showBadge || showPgp || showAuth}
  <span class="badges">
    {#if showBadge}
      <span class="badge" title={`${accountEmail} · ${folderName}`}>
        <IconMailbox size={12} stroke={1.6} />
        <span class="badge-text">{folderName || accountEmail}</span>
      </span>
    {/if}

    {#if showPgp}
      <span class="badge pgp" title={`PGP: ${pgpLabel(pgpStatus, $t)}`} aria-label={`PGP ${pgpLabel(pgpStatus, $t)}`}>
        {#if pgpStatus === 'encrypted'}
          <IconLock size={12} stroke={1.6} />
        {:else}
          <IconShieldCheck size={12} stroke={1.6} />
        {/if}
        <span class="badge-text">{pgpLabel(pgpStatus, $t)}</span>
      </span>
    {/if}

    {#if showAuth}
      <span
        class="badge auth"
        title={$t('common.techBadges.authTitle')}
        aria-label={$t('common.techBadges.authAriaLabel')}
      >
        <IconShieldQuestion size={12} stroke={1.6} />
        <span class="badge-text">{authText}</span>
      </span>
    {/if}
  </span>
{/if}

<style>
  .badges {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
  }

  .badge {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: 1px var(--space-2);
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-tertiary);
    font-size: var(--fz-meta);
    line-height: 1.4;
    max-width: 16ch;
  }

  .badge-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* auth is deliberately the dimmest: it carries no real data yet. */
  .auth {
    opacity: 0.7;
    font-style: italic;
  }
</style>
