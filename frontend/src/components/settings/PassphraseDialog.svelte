<script lang="ts">
  // the prompt for a locked PGP private key's passphrase (#192).
  //
  // What is typed here goes straight to the backend, which checks it against
  // the key before holding it. "Remember" puts it in the OS keyring, the same
  // place account passwords live; unticked, it is held in memory and forgotten
  // when Pelton quits. Either way it never reaches the settings database.
  import { IconLock } from '@tabler/icons-svelte'
  import Modal from '../common/Modal.svelte'
  import { passphraseRequest, closePassphrase } from '../../stores/passphrase'
  import { unlockPGPKey } from '../../lib/api'
  import { errorMessage, toastError } from '../../stores/toast'
  import { t } from '../../lib/i18n'

  let passphrase = ''
  let remember = false
  let busy = false
  let failed = false
  // the request the fields were reset for, so reopening for a different key
  // clears them without fighting the user's typing.
  let seededFor: unknown = null

  $: request = $passphraseRequest
  $: if (request !== seededFor) {
    seededFor = request
    passphrase = ''
    remember = false
    busy = false
    failed = false
  }

  $: keyLabel = request
    ? request.key.name && request.key.email
      ? `${request.key.name} <${request.key.email}>`
      : request.key.email || request.key.fingerprint.slice(-16)
    : ''

  async function submit(): Promise<void> {
    if (!request || busy || passphrase === '') {
      return
    }
    busy = true
    failed = false
    try {
      await unlockPGPKey(request.key.fingerprint, passphrase, remember)
      const done = request.onUnlocked
      // drop the typed value as soon as it has been handed over.
      passphrase = ''
      closePassphrase()
      done()
    } catch (err) {
      // a wrong passphrase is the expected case, so it corrects in place
      // rather than closing the dialog and making the user start again.
      failed = true
      passphrase = ''
      toastError(errorMessage(err))
    } finally {
      busy = false
    }
  }

</script>

{#if request}
  <Modal title={$t('encryption.passphrase.title')} size="small" on:close={closePassphrase}>
    <span slot="icon"><IconLock size={16} stroke={1.8} /></span>

    <p class="key">{keyLabel}</p>

    <form on:submit|preventDefault={submit}>
      <label class="field">
        <span>{$t('encryption.passphrase.label')}</span>
        <!-- svelte-ignore a11y-autofocus -->
        <input
          type="password"
          bind:value={passphrase}
          autofocus
          spellcheck="false"
          autocomplete="off"
          class:failed
        />
      </label>

      <label class="remember">
        <input type="checkbox" bind:checked={remember} />
        <span>
          {$t('encryption.passphrase.remember')}
          <small>{$t('encryption.passphrase.rememberHint')}</small>
        </span>
      </label>

    </form>

    <svelte:fragment slot="footer">
      <button type="button" class="cancel" on:click={closePassphrase}>
        {$t('folders.cancel')}
      </button>
      <button type="button" class="go" disabled={busy || passphrase === ''} on:click={submit}>
        {$t('encryption.unlock')}
      </button>
    </svelte:fragment>
  </Modal>
{/if}

<style>

  .key {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-secondary);
    overflow-wrap: anywhere;
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
  .field input.failed {
    border-color: var(--danger);
  }

  .remember {
    display: flex;
    align-items: flex-start;
    gap: var(--space-2);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
    cursor: var(--cursor-action);
  }

  .remember small {
    display: block;
    color: var(--text-tertiary);
  }

  .cancel {
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
    color: var(--text-secondary);
    background: transparent;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    cursor: var(--cursor-action);
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
    cursor: var(--cursor-action);
  }
  .go:disabled {
    opacity: 0.5;
    cursor: default;
  }
</style>
