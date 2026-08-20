<script lang="ts">
  // the pgp control in the compose footer: sign, encrypt, both or neither, for
  // this one message. It only ever offers what the keys actually allow, so a
  // send cannot be set up here and then refused at the last moment, and it says
  // which recipient is holding encryption back rather than "someone".
  import { createEventDispatcher } from 'svelte'
  import { IconLock, IconLockOpen, IconSignature, IconAlertTriangle } from '@tabler/icons-svelte'
  import { t } from '../../lib/i18n'
  import type { ProtectionStatus } from '../../lib/types'

  // the chosen protection: 'none' | 'sign' | 'encrypt' | 'signencrypt'.
  export let protection: string = 'none'
  // what is possible for this account and these recipients; null while the
  // first status call is still in flight.
  export let status: ProtectionStatus | null = null

  const dispatch = createEventDispatcher<{ change: string }>()

  $: canSign = status?.canSign ?? false
  $: canEncrypt = status?.canEncrypt ?? false
  $: signing = protection === 'sign' || protection === 'signencrypt'
  $: encrypting = protection === 'encrypt' || protection === 'signencrypt'
  // the recipients with no key are the reason encryption is unavailable, and
  // naming them is the difference between a useful control and a dead one.
  $: missing = (status?.recipients ?? []).filter((r) => !r.hasKey).map((r) => r.email)

  function combine(sign: boolean, encrypt: boolean): string {
    if (sign && encrypt) return 'signencrypt'
    if (sign) return 'sign'
    if (encrypt) return 'encrypt'
    return 'none'
  }

  function toggleSign(): void {
    if (!canSign) {
      return
    }
    dispatch('change', combine(!signing, encrypting))
  }

  function toggleEncrypt(): void {
    if (!canEncrypt) {
      return
    }
    dispatch('change', combine(signing, !encrypting))
  }

  $: encryptTitle = canEncrypt
    ? $t('compose.protection.encryptTitle')
    : missing.length > 0
      ? $t('compose.protection.noKeyFor').replace('{emails}', missing.join(', '))
      : $t('compose.protection.noRecipients')

  $: signTitle = canSign
    ? status?.signerLocked
      ? $t('compose.protection.signLockedTitle')
      : $t('compose.protection.signTitle')
    : $t('compose.protection.noSigningKey')
</script>

<div class="protection" role="group" aria-label={$t('compose.protection.label')}>
  <button
    type="button"
    class="chip"
    class:on={signing}
    disabled={!canSign}
    title={signTitle}
    aria-pressed={signing}
    on:click={toggleSign}
  >
    <IconSignature size={14} stroke={1.7} />
    {$t('compose.protection.sign')}
  </button>
  <button
    type="button"
    class="chip"
    class:on={encrypting}
    disabled={!canEncrypt}
    title={encryptTitle}
    aria-pressed={encrypting}
    on:click={toggleEncrypt}
  >
    {#if encrypting}
      <IconLock size={14} stroke={1.7} />
    {:else}
      <IconLockOpen size={14} stroke={1.7} />
    {/if}
    {$t('compose.protection.encrypt')}
  </button>
  {#if encrypting && status?.signerLocked && signing}
    <span class="hint" title={$t('compose.protection.signLockedTitle')}>
      <IconAlertTriangle size={13} stroke={1.7} />
    </span>
  {/if}
</div>

<style>
  .protection {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
  }

  .chip {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-1) var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--text-secondary);
    font-size: var(--fz-meta);
    cursor: var(--cursor-action);
  }

  .chip:hover:not(:disabled) {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  /* an active chip is the one state worth colouring: it is saying that this
     message is going to leave the machine protected. */
  .chip.on {
    border-color: var(--accent);
    background: var(--accent);
    color: var(--accent-fg);
  }

  .chip:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .hint {
    display: inline-flex;
    color: var(--warning);
  }
</style>
