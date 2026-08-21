<script lang="ts">
  // the per-message technical-info badges shared by the list row and the detail
  // header: mailbox/account badge, pgp status and auth status. each badge is
  // shown only when its preference toggle is on. the auth badge reports what
  // the receiving server said about spf, dkim and dmarc; mail whose server said
  // nothing reads as "not available" and never as a failure.
  import {
    IconMailbox,
    IconLock,
    IconShieldCheck,
    IconShieldQuestion,
    IconShieldCheckFilled,
    IconShieldExclamation,
    IconCertificate,
    IconCertificateOff,
    IconShieldX,
  } from '@tabler/icons-svelte'
  import { prefs } from '../../stores/prefs'
  import { t } from '../../lib/i18n'
  import type { PGPStatus, SMIMESignature, SMIMERevocation } from '../../lib/types'

  export let accountEmail: string = ''
  export let folderName: string = ''
  export let pgp: string = 'none'
  export let auth: string = 'unavailable'
  // the s/mime signature verdict; an empty status renders nothing.
  export let smime: SMIMESignature | undefined = undefined
  // what the issuing authority says about that certificate now, when the reader
  // has turned revocation checking on. An empty status means no check was made
  // and the badge reads exactly as it did before.
  export let revocation: SMIMERevocation | undefined = undefined

  // pgp label and icon per status. "none" renders nothing.
  function pgpLabel(status: string, tFn: (key: string) => string): string {
    if (status === 'encrypted') return tFn('common.techBadges.encrypted')
    if (status === 'signed') return tFn('common.techBadges.signed')
    return ''
  }

  // the s/mime badge rides the same preference as pgp: both answer "is this
  // message cryptographically protected", and splitting them into two toggles
  // would make the reader think about a distinction they do not care about.
  // a withdrawn certificate outranks the signature verdict. The bytes are still
  // intact, but the authority is positively saying the key must not be relied
  // on, and "Signature verified" would be the one badge that actively misleads.
  $: revoked = revocation?.status === 'revoked'
  $: smimeStatus = revoked ? 'revoked' : (smime?.status ?? '')
  $: showSmime = $prefs.showPgp && smimeStatus !== ''
  // the signer is what the badge is actually asserting, so it leads the tooltip
  // and falls back to the address when the certificate carries no name.
  $: smimeTitle = smime
    ? [
        $t(`common.techBadges.smime.${smimeStatus}`),
        smime.signer || smime.email,
        revoked ? revocation?.detail : smime.detail,
        revocationNote,
      ]
        .filter((part) => part !== '' && part !== undefined)
        .join(' · ')
    : ''

  // the soft-fail line. Being offline says nothing about a certificate, so a
  // check that could not be made leaves the verdict alone and says so in words
  // rather than by changing what the badge claims.
  $: revocationNote =
    revocation?.status === 'unknown'
      ? $t('common.techBadges.smime.uncheckedNote')
      : revocation?.status === 'good'
        ? $t('common.techBadges.smime.checkedNote')
        : ''

  $: showBadge = $prefs.showMailboxBadge && (accountEmail !== '' || folderName !== '')
  $: showPgp = $prefs.showPgp && pgp !== 'none'
  $: showAuth = $prefs.showAuth
  $: pgpStatus = pgp as PGPStatus
  // pass, partial, fail, or unavailable when the server reported nothing, which
  // is most older cached mail. Anything unrecognised is shown as unavailable
  // rather than echoed, so a future value cannot leak into the badge raw.
  $: authKnown = auth === 'pass' || auth === 'partial' || auth === 'fail'
  $: authText = authKnown ? $t(`common.techBadges.auth.${auth}`) : $t('common.techBadges.authNA')
  $: authTitle = authKnown ? $t(`common.techBadges.authTitle.${auth}`) : $t('common.techBadges.authTitle')
  $: AuthIcon = auth === 'pass' ? IconShieldCheckFilled : auth === 'fail' ? IconShieldExclamation : IconShieldQuestion
</script>

{#if showBadge || showPgp || showSmime || showAuth}
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

    {#if showSmime}
      <span class="badge smime {smimeStatus}" title={smimeTitle} aria-label={smimeTitle}>
        {#if smimeStatus === 'valid'}
          <IconCertificate size={12} stroke={1.6} />
        {:else if smimeStatus === 'invalid' || smimeStatus === 'revoked'}
          <IconShieldX size={12} stroke={1.6} />
        {:else}
          <IconCertificateOff size={12} stroke={1.6} />
        {/if}
        <span class="badge-text">{$t(`common.techBadges.smime.${smimeStatus}`)}</span>
      </span>
    {/if}

    {#if showAuth}
      <span
        class="badge auth"
        class:auth-fail={auth === 'fail'}
        title={authTitle}
        aria-label={authTitle}
      >
        <svelte:component this={AuthIcon} size={12} stroke={1.6} />
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

  /* the s/mime verdict is the one badge that carries a judgement, so it is the
     one that takes colour. a signature that fails reads as a warning; one that
     merely cannot be vouched for stays neutral rather than alarming, since
     unverifiable is not the same as forged. */
  .smime.valid {
    color: var(--success);
    border-color: var(--success);
  }
  .smime.revoked,
  .smime.invalid {
    color: var(--danger);
    border-color: var(--danger);
  }

  /* auth is deliberately the dimmest: it carries no real data yet. */
  /* the muted, italic treatment was for a badge that never said anything. It
     stays for the unknown case and comes off once there is a real result. */
  .auth {
    opacity: 0.7;
  }

  .auth-fail {
    opacity: 1;
    color: var(--danger);
  }
</style>
