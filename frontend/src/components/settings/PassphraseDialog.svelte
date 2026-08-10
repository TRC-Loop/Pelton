<script lang="ts">
  // the prompt for a locked PGP private key's passphrase (#192).
  //
  // What is typed here goes straight to the backend, which checks it against
  // the key before holding it. "Remember" puts it in the OS keyring, the same
  // place account passwords live; unticked, it is held in memory and forgotten
  // when Pelton quits. Either way it never reaches the settings database.
  import { fade, scale } from 'svelte/transition'
  import { IconX, IconLock } from '@tabler/icons-svelte'
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

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      closePassphrase()
    }
  }
</script>

<svelte:window on:keydown={request ? onKeydown : undefined} />

{#if request}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="backdrop" transition:fade={{ duration: 120 }} on:click={closePassphrase}></div>
  <div
    class="dialog"
    role="dialog"
    aria-modal="true"
    aria-label={$t('encryption.passphrase.title')}
    transition:scale={{ duration: 150, start: 0.94 }}
  >
    <header>
      <h2>
        <span class="glyph"><IconLock size={16} stroke={1.8} /></span>
        {$t('encryption.passphrase.title')}
      </h2>
      <button type="button" class="close" aria-label={$t('folders.cancel')} on:click={closePassphrase}>
        <IconX size={16} stroke={1.8} />
      </button>
    </header>

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

      <div class="actions">
        <button type="button" class="cancel" on:click={closePassphrase}>
          {$t('folders.cancel')}
        </button>
        <button type="submit" class="go" disabled={busy || passphrase === ''}>
          {$t('encryption.unlock')}
        </button>
      </div>
    </form>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 320;
    background: var(--scrim, rgba(0, 0, 0, 0.4));
    backdrop-filter: blur(2px);
  }

  .dialog {
    position: fixed;
    z-index: 321;
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
    cursor: pointer;
  }

  .remember small {
    display: block;
    color: var(--text-tertiary);
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
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
