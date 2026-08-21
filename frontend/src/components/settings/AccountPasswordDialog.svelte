<script lang="ts">
  // asks for a mailbox password at sync time.
  //
  // An account imported from another mail client has no password: the import
  // brings the server settings across but the other client's stored secret is
  // not readable. Without this the account simply never syncs and never says
  // why, so the sync path raises this prompt instead.
  //
  // Cancelling is a real answer. It skips that account for this sync rather
  // than blocking the others, and the same prompt comes back next time.
  // "Don't ask again" is the stronger form: the prompt stops interrupting for
  // that account entirely and a warning marker next to the mailbox takes over
  // saying it cannot sync (#290).
  import { fade, scale } from 'svelte/transition'
  import { IconX, IconLock } from '@tabler/icons-svelte'
  import { setAccountPassword } from '../../lib/api'
  import { errorMessage, toastError } from '../../stores/toast'
  import { t } from '../../lib/i18n'
  import type { PasswordPromptResult } from '../../stores/passwordprompt'
  import type { Account } from '../../lib/types'

  /** The account being asked about, or null when the prompt is closed. */
  export let account: Account | null = null
  /** Called with how the prompt ended. */
  export let onDone: (result: PasswordPromptResult) => void = () => {}

  let password = ''
  let busy = false
  let seededFor: unknown = null

  // the address, not the display name. Several mailboxes named "Work" and
  // "Private" all look the same in this prompt, and the address is the part
  // that says which login is being asked for, so it is the part that must
  // never be the one dropped. The name comes along when there is one and it
  // adds something.
  $: who =
    account && account.displayName && account.displayName !== account.email
      ? `${account.displayName} (${account.email})`
      : (account?.email ?? '')

  $: if (account !== seededFor) {
    seededFor = account
    password = ''
    busy = false
  }

  async function submit(): Promise<void> {
    if (!account || busy || password === '') {
      return
    }
    busy = true
    try {
      await setAccountPassword(account.id, password)
      password = ''
      onDone('saved')
    } catch (err) {
      toastError(errorMessage(err))
      password = ''
    } finally {
      busy = false
    }
  }

  function cancel(): void {
    password = ''
    onDone('skipped')
  }

  function dismiss(): void {
    password = ''
    onDone('dismissed')
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      cancel()
    }
  }
</script>

<svelte:window on:keydown={account ? onKeydown : undefined} />

{#if account}
  <div class="backdrop" transition:fade={{ duration: 120 }}></div>
  <div
    class="dialog"
    role="dialog"
    aria-modal="true"
    aria-label={$t('mailboxes.passwordPrompt.title')}
    transition:scale={{ duration: 150, start: 0.94 }}
  >
    <header>
      <h2>
        <span class="glyph"><IconLock size={16} stroke={1.8} /></span>
        {$t('mailboxes.passwordPrompt.title')}
      </h2>
      <button type="button" class="close" aria-label={$t('mailboxes.cancel')} on:click={cancel}>
        <IconX size={16} stroke={1.8} />
      </button>
    </header>

    <p class="lead">
      {$t('mailboxes.passwordPrompt.body').replace('{email}', who)}
    </p>
    <p class="lead">{$t('mailboxes.passwordPrompt.dismissHint')}</p>

    <form on:submit|preventDefault={submit}>
      <label class="field">
        <span>{$t('wizard.field.password')}</span>
        <!-- svelte-ignore a11y-autofocus -->
        <input type="password" bind:value={password} autofocus autocomplete="off" spellcheck="false" />
      </label>

      <div class="actions">
        <button type="button" class="dismiss" on:click={dismiss}>
          {$t('mailboxes.passwordPrompt.dismiss')}
        </button>
        <button type="button" class="cancel" on:click={cancel}>
          {$t('mailboxes.passwordPrompt.skip')}
        </button>
        <button type="submit" class="go" disabled={busy || password === ''}>
          {busy ? $t('mailboxes.saving') : $t('mailboxes.save')}
        </button>
      </div>
    </form>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 330;
    background: var(--scrim, rgba(0, 0, 0, 0.4));
    backdrop-filter: blur(2px);
  }

  .dialog {
    position: fixed;
    z-index: 331;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: min(420px, calc(100vw - 2 * var(--space-5)));
    padding: var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
  }

  h2 {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin: 0;
    font-size: var(--fz-title);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .glyph {
    display: inline-flex;
    color: var(--text-tertiary);
  }

  .close {
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: var(--space-1);
    border-radius: var(--radius-control);
  }
  .close:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .lead {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  form {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .field span {
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .field input {
    padding: var(--space-2) var(--space-3);
    font-family: var(--font-ui);
    font-size: var(--fz-body);
    color: var(--text-primary);
    background: var(--surface-raised);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
  }
  .field input:focus {
    outline: none;
    border-color: var(--accent);
  }

  .actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: var(--space-2);
  }

  /* the quiet option, kept away from the two that answer the question now. */
  .dismiss {
    margin-right: auto;
    padding: var(--space-2) 0;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    background: transparent;
    border: none;
    cursor: var(--cursor-action);
    text-decoration: underline;
    text-underline-offset: 2px;
  }
  .dismiss:hover {
    color: var(--text-primary);
  }

  .cancel {
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
    color: var(--text-secondary);
    background: transparent;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    cursor: pointer;
  }
  .cancel:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .go {
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--accent-fg);
    background: var(--accent);
    border: none;
    border-radius: var(--radius-control);
    cursor: pointer;
  }
  .go:disabled {
    opacity: 0.5;
    cursor: default;
  }
</style>
