<script lang="ts">
  // the fullscreen add-mailbox wizard. it walks: provider -> (oauth client id |
  // password + servers) -> test/sign-in -> done. autodiscovery prefills servers
  // for custom providers; oauth uses the per-user PKCE flow (the user supplies
  // their own client id). this component is code-split and only loaded when the
  // user opens it, so its cost is not paid at startup.
  import { createEventDispatcher, onMount } from 'svelte'
  import { IconX, IconArrowLeft, IconCheck } from '@tabler/icons-svelte'
  import WizardProviders from './WizardProviders.svelte'
  import Spinner from '../common/Spinner.svelte'
  import { discoverConfig, testConnection, addPasswordAccount, addOAuthAccount } from '../../lib/api'
  import { errorMessage } from '../../stores/toast'
  import { providerPresets, type ProviderPreset } from '../../lib/providers'
  import type { AddAccountRequest, Account } from '../../lib/types'

  const dispatch = createEventDispatcher<{ close: void; added: Account }>()

  // when set, the wizard skips the provider grid and opens straight into that
  // provider's setup. used by the onboarding provider cards.
  export let initialProviderId: string | null = null

  type Step = 'provider' | 'config' | 'oauth' | 'working' | 'done' | 'error'
  let step: Step = 'provider'
  let preset: ProviderPreset | null = null
  let error = ''
  let workingMessage = ''
  let testing = false
  let testOk: boolean | null = null

  // the account draft being assembled across steps.
  let draft: AddAccountRequest = blankDraft()

  function blankDraft(): AddAccountRequest {
    return {
      email: '',
      displayName: '',
      imapHost: '',
      imapPort: 993,
      smtpHost: '',
      smtpPort: 465,
      password: '',
      provider: '',
      clientId: '',
      clientSecret: '',
    }
  }

  // whether the advanced section (tls mode, oauth secret) is expanded.
  let showAdvanced = false

  // the imap transport, derived from the port: 143 is STARTTLS, anything else is
  // implicit TLS. selecting a mode sets the conventional port, which the backend
  // reads to choose the transport.
  $: imapTLS = draft.imapPort === 143 ? 'starttls' : 'ssl'

  function setTLS(mode: string): void {
    draft.imapPort = mode === 'starttls' ? 143 : 993
  }

  function selectPreset(p: ProviderPreset): void {
    preset = p
    draft = blankDraft()
    if (p.imapHost) draft.imapHost = p.imapHost
    if (p.imapPort) draft.imapPort = p.imapPort
    if (p.smtpHost) draft.smtpHost = p.smtpHost
    if (p.smtpPort) draft.smtpPort = p.smtpPort
    if (p.oauthProvider) draft.provider = p.oauthProvider
    testOk = null
    error = ''
    showAdvanced = false
    step = p.kind === 'oauth' ? 'oauth' : 'config'
  }

  function pick(event: CustomEvent<ProviderPreset>): void {
    selectPreset(event.detail)
  }

  // jump straight to a provider when the caller requested one.
  onMount(() => {
    if (initialProviderId) {
      const p = providerPresets.find((x) => x.id === initialProviderId)
      if (p) {
        selectPreset(p)
      }
    }
  })

  // for custom providers, try autodiscovery once a full address is present.
  async function maybeDiscover(): Promise<void> {
    if (!preset?.custom || !draft.email.includes('@')) {
      return
    }
    try {
      const d = await discoverConfig(draft.email)
      draft.imapHost = d.imapHost
      draft.imapPort = d.imapPort
      draft.smtpHost = d.smtpHost
      draft.smtpPort = d.smtpPort
    } catch {
      // leave fields for manual entry; discovery is best effort.
    }
  }

  async function test(): Promise<void> {
    testing = true
    testOk = null
    error = ''
    try {
      await testConnection({
        email: draft.email,
        imapHost: draft.imapHost,
        imapPort: draft.imapPort,
        password: draft.password,
      })
      testOk = true
    } catch (err) {
      testOk = false
      error = errorMessage(err)
    } finally {
      testing = false
    }
  }

  async function addPassword(): Promise<void> {
    step = 'working'
    workingMessage = 'Connecting and discovering folders…'
    try {
      const account = await addPasswordAccount(draft)
      finish(account)
    } catch (err) {
      fail(err)
    }
  }

  async function signIn(): Promise<void> {
    step = 'working'
    workingMessage = 'Waiting for sign-in in your browser…'
    try {
      const account = await addOAuthAccount(draft)
      finish(account)
    } catch (err) {
      fail(err)
    }
  }

  function finish(account: Account): void {
    step = 'done'
    dispatch('added', account)
  }

  function fail(err: unknown): void {
    error = errorMessage(err)
    step = 'error'
  }

  function back(): void {
    error = ''
    step = 'provider'
    preset = null
  }

  $: canSubmitPassword = draft.email.includes('@') && draft.password !== '' && draft.imapHost !== ''
  $: canSignIn = draft.email.includes('@') && draft.clientId !== ''
</script>

<div class="screen" role="dialog" aria-modal="true" aria-label="Add mailbox">
  <header class="head">
    {#if step === 'config' || step === 'oauth'}
      <button type="button" class="icon" aria-label="Back" on:click={back}>
        <IconArrowLeft size={18} stroke={1.8} />
      </button>
    {:else}
      <span class="icon-spacer"></span>
    {/if}
    <span class="title">Add mailbox</span>
    <button type="button" class="icon" aria-label="Close" on:click={() => dispatch('close')}>
      <IconX size={18} stroke={1.8} />
    </button>
  </header>

  <div class="body">
    <div class="content">
      {#if step === 'provider'}
        <WizardProviders on:pick={pick} />
      {:else if step === 'config'}
        <h3>{preset?.label}</h3>
        {#if preset?.note}<p class="note">{preset.note}</p>{/if}

        <label class="field">
          <span>Email</span>
          <input type="email" bind:value={draft.email} on:blur={maybeDiscover} placeholder="you@example.com" />
        </label>
        <label class="field">
          <span>Display name</span>
          <input type="text" bind:value={draft.displayName} placeholder="Your name" />
        </label>
        <label class="field">
          <span>Password</span>
          <input type="password" bind:value={draft.password} placeholder="Password or app password" />
        </label>

        <div class="servers">
          <label class="field"><span>IMAP host</span><input type="text" bind:value={draft.imapHost} /></label>
          <label class="field narrow"><span>Port</span><input type="number" bind:value={draft.imapPort} /></label>
        </div>
        <div class="servers">
          <label class="field"><span>SMTP host</span><input type="text" bind:value={draft.smtpHost} /></label>
          <label class="field narrow"><span>Port</span><input type="number" bind:value={draft.smtpPort} /></label>
        </div>

        <button type="button" class="disclosure" on:click={() => (showAdvanced = !showAdvanced)}>
          {showAdvanced ? 'Hide' : 'Advanced'} connection settings
        </button>
        {#if showAdvanced}
          <div class="advanced">
            <span class="adv-label">IMAP security</span>
            <div class="seg" role="radiogroup" aria-label="IMAP security">
              <button type="button" class:on={imapTLS === 'ssl'} on:click={() => setTLS('ssl')}>SSL / TLS</button>
              <button type="button" class:on={imapTLS === 'starttls'} on:click={() => setTLS('starttls')}>STARTTLS</button>
            </div>
            <p class="adv-hint">
              SSL / TLS uses port 993, STARTTLS uses port 143. Most servers use SSL / TLS; pick STARTTLS only
              if your provider requires it.
            </p>
          </div>
        {/if}

        {#if testOk === true}<p class="ok"><IconCheck size={14} stroke={2} /> Connection works.</p>{/if}
        {#if error}<p class="err">{error}</p>{/if}

        <div class="actions">
          <button type="button" class="ghost" on:click={test} disabled={testing || !canSubmitPassword}>
            {testing ? 'Testing…' : 'Test connection'}
          </button>
          <button type="button" class="primary" on:click={addPassword} disabled={!canSubmitPassword}>
            Add mailbox
          </button>
        </div>
      {:else if step === 'oauth'}
        <h3>Sign in to {preset?.label}</h3>
        <p class="note">
          OAuth uses your own registered client id (a desktop OAuth client). Paste it below; sign-in opens in
          your browser and no client secret is stored.
        </p>

        <label class="field">
          <span>Email</span>
          <input type="email" bind:value={draft.email} placeholder="you@example.com" />
        </label>
        <label class="field">
          <span>Display name</span>
          <input type="text" bind:value={draft.displayName} placeholder="Your name" />
        </label>
        <label class="field">
          <span>OAuth client id</span>
          <input type="text" bind:value={draft.clientId} placeholder="xxxxx.apps.googleusercontent.com" />
        </label>

        {#if preset?.allowClientSecret}
          <button type="button" class="disclosure" on:click={() => (showAdvanced = !showAdvanced)}>
            {showAdvanced ? 'Hide' : 'Advanced'}
          </button>
          {#if showAdvanced}
            <label class="field">
              <span>OAuth client secret (optional)</span>
              <input type="password" bind:value={draft.clientSecret} placeholder="Only for confidential-client apps" />
            </label>
            <p class="adv-hint">
              Leave blank for a normal desktop (public) app registration. Set this only if your Microsoft
              Entra app is registered as a confidential client that requires a secret.
            </p>
          {/if}
        {/if}

        {#if error}<p class="err">{error}</p>{/if}

        <div class="actions">
          <button type="button" class="primary" on:click={signIn} disabled={!canSignIn}>
            Sign in with {preset?.label}
          </button>
        </div>
      {:else if step === 'working'}
        <Spinner label={workingMessage} />
      {:else if step === 'done'}
        <div class="result">
          <IconCheck size={32} stroke={1.6} />
          <h3>Mailbox added</h3>
          <p class="note">{draft.email} is syncing now.</p>
          <button type="button" class="primary" on:click={() => dispatch('close')}>Done</button>
        </div>
      {:else if step === 'error'}
        <div class="result">
          <h3>Could not add mailbox</h3>
          <p class="err">{error}</p>
          <button type="button" class="ghost" on:click={() => (step = preset?.kind === 'oauth' ? 'oauth' : 'config')}>
            Back
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .screen {
    position: fixed;
    inset: 0;
    /* above the onboarding overlay (120) so it is usable when launched from the
       onboarding mailbox step. */
    z-index: 150;
    display: flex;
    flex-direction: column;
    background: var(--surface-base);
  }

  .head {
    display: grid;
    grid-template-columns: 40px 1fr 40px;
    align-items: center;
    padding: var(--space-3) var(--space-5);
    border-bottom: var(--hairline) solid var(--border-default);
  }

  .title {
    text-align: center;
    font-weight: var(--fw-semibold);
  }

  .icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    border-radius: var(--radius-control);
  }

  /* keep the back button off the very edge and the close button flush right so
     they sit balanced inside the header padding. */
  .head .icon:first-child {
    justify-self: start;
  }

  .head .icon:last-child {
    justify-self: end;
  }

  .icon:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .icon-spacer {
    width: 40px;
  }

  .body {
    flex: 1;
    overflow-y: auto;
  }

  .content {
    max-width: 480px;
    margin: 0 auto;
    padding: var(--space-6);
  }

  h3 {
    margin: 0 0 var(--space-2);
    font-size: var(--fz-title);
    font-weight: var(--fw-semibold);
  }

  .note {
    margin: 0 0 var(--space-4);
    color: var(--text-secondary);
    font-size: var(--fz-label);
    line-height: 1.5;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
  }

  .field span {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }

  .field input {
    height: var(--control-height);
    padding: 0 var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    outline: none;
  }

  .field input:focus {
    border-color: var(--accent);
  }

  .servers {
    display: flex;
    gap: var(--space-3);
  }

  .servers .field {
    flex: 1;
  }

  .servers .field.narrow {
    flex: 0 0 88px;
  }

  /* a lightweight text toggle that reveals the advanced settings. */
  .disclosure {
    border: none;
    background: transparent;
    color: var(--accent);
    cursor: pointer;
    font-size: var(--fz-label);
    padding: var(--space-1) 0;
    margin-bottom: var(--space-2);
  }

  .disclosure:hover {
    text-decoration: underline;
  }

  .advanced {
    margin-bottom: var(--space-3);
  }

  .adv-label {
    display: block;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    margin-bottom: var(--space-2);
  }

  /* a small two-option segmented control for the tls mode. */
  .seg {
    display: inline-flex;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    overflow: hidden;
  }

  .seg button {
    border: none;
    background: var(--surface-raised);
    color: var(--text-secondary);
    cursor: pointer;
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
  }

  .seg button + button {
    border-left: var(--hairline) solid var(--border-default);
  }

  .seg button.on {
    background: var(--accent);
    color: var(--accent-fg);
  }

  .adv-hint {
    margin: var(--space-2) 0 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-3);
    margin-top: var(--space-4);
  }

  .primary,
  .ghost {
    padding: var(--space-2) var(--space-5);
    border-radius: var(--radius-control);
    border: var(--hairline) solid var(--border-default);
    cursor: pointer;
    font-size: var(--fz-label);
  }

  .primary {
    background: var(--accent);
    color: var(--accent-fg);
    border-color: transparent;
  }

  .primary:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .ghost {
    background: var(--surface-raised);
    color: var(--text-primary);
  }

  .ghost:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .ok {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--success);
    font-size: var(--fz-label);
  }

  .err {
    color: var(--danger);
    font-size: var(--fz-label);
    word-break: break-word;
  }

  .result {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-3);
    text-align: center;
    color: var(--success);
    padding: var(--space-6) 0;
  }

  .result h3 {
    color: var(--text-primary);
  }
</style>
