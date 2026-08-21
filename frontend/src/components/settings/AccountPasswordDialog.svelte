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
  import { IconLock } from '@tabler/icons-svelte'
  import Modal from '../common/Modal.svelte'
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

</script>

{#if account}
  <Modal title={$t('mailboxes.passwordPrompt.title')} size="small" on:close={cancel}>
    <span slot="icon"><IconLock size={16} stroke={1.8} /></span>

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

    </form>

    <svelte:fragment slot="footer">
      <button type="button" class="dismiss" on:click={dismiss}>
        {$t('mailboxes.passwordPrompt.dismiss')}
      </button>
      <button type="button" class="cancel" on:click={cancel}>
        {$t('mailboxes.passwordPrompt.skip')}
      </button>
      <button type="button" class="go" disabled={busy || password === ''} on:click={submit}>
        {busy ? $t('mailboxes.saving') : $t('mailboxes.save')}
      </button>
    </svelte:fragment>
  </Modal>
{/if}

<style>

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
