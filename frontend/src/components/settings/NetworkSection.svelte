<script lang="ts">
  // the network settings: an outbound proxy for all of Pelton's connections
  // (IMAP, SMTP and the app's own web calls). Off and System need no fields;
  // Manual exposes a SOCKS5 or HTTP proxy with optional credentials. The
  // password is write-only - the backend never sends the stored one back, so an
  // untouched field with hasPassword shows a placeholder and keeps the secret.
  import { onMount } from 'svelte'
  import { IconCheck } from '@tabler/icons-svelte'
  import SegmentedSetting from './SegmentedSetting.svelte'
  import { getProxyConfig, setProxyConfig, testProxy } from '../../lib/api'
  import { errorMessage, toastError, toastSuccess } from '../../stores/toast'
  import { t } from '../../lib/i18n'
  import type { ProxyConfig } from '../../lib/types'

  let cfg: ProxyConfig = { mode: 'off', scheme: 'socks5', host: '', port: 1080, username: '', password: '', hasPassword: false }
  let loading = true
  let saving = false
  let testing = false
  // whether the user typed a new password this session (so we send it).
  let passwordTouched = false

  onMount(async () => {
    try {
      cfg = await getProxyConfig()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      loading = false
    }
  })

  function setMode(mode: string): void {
    cfg.mode = mode
  }

  function setScheme(scheme: string): void {
    cfg.scheme = scheme
    // nudge the port to the scheme's conventional default when it still holds
    // the other scheme's default, so switching type does the expected thing.
    if (scheme === 'http' && (cfg.port === 1080 || !cfg.port)) cfg.port = 8080
    if (scheme === 'socks5' && cfg.port === 8080) cfg.port = 1080
  }

  // the payload to send: drop the password unless the user typed one, so the
  // backend keeps the stored secret behind its placeholder.
  function payload(): ProxyConfig {
    return { ...cfg, password: passwordTouched ? cfg.password : '' }
  }

  async function save(): Promise<void> {
    saving = true
    try {
      await setProxyConfig(payload())
      cfg = await getProxyConfig()
      passwordTouched = false
      toastSuccess($t('network.proxy.saved'))
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      saving = false
    }
  }

  async function test(): Promise<void> {
    testing = true
    try {
      await testProxy(payload())
      toastSuccess($t('network.proxy.testOk'))
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      testing = false
    }
  }

  const modeOptions = [
    { key: 'off', label: $t('network.proxy.mode.off') },
    { key: 'system', label: $t('network.proxy.mode.system') },
    { key: 'manual', label: $t('network.proxy.mode.manual') },
  ]
  const schemeOptions = [
    { key: 'socks5', label: 'SOCKS5' },
    { key: 'http', label: 'HTTP' },
  ]
</script>

<h3>{$t('settingsPanel.category.network')}</h3>
<p class="hint">{$t('network.proxy.hint')}</p>

{#if loading}
  <p class="hint">{$t('mailboxes.loading')}</p>
{:else}
  <SegmentedSetting label={$t('network.proxy.mode')} value={cfg.mode} options={modeOptions} on:change={(e) => setMode(e.detail)} />

  {#if cfg.mode === 'system'}
    <p class="hint">{$t('network.proxy.systemHint')}</p>
  {/if}

  {#if cfg.mode === 'manual'}
    <div class="manual">
      <SegmentedSetting label={$t('network.proxy.type')} value={cfg.scheme} options={schemeOptions} on:change={(e) => setScheme(e.detail)} />

      <div class="servers">
        <label class="field"><span>{$t('network.proxy.host')}</span><input type="text" bind:value={cfg.host} placeholder="127.0.0.1" /></label>
        <label class="field narrow"><span>{$t('wizard.field.port')}</span><input type="number" bind:value={cfg.port} /></label>
      </div>
      <label class="field">
        <span>{$t('network.proxy.username')}</span>
        <input type="text" bind:value={cfg.username} placeholder={$t('network.proxy.optional')} />
      </label>
      <label class="field">
        <span>{$t('network.proxy.password')}</span>
        <input
          type="password"
          bind:value={cfg.password}
          on:input={() => (passwordTouched = true)}
          placeholder={cfg.hasPassword && !passwordTouched ? $t('network.proxy.passwordStored') : $t('network.proxy.optional')}
        />
      </label>
    </div>
  {/if}

  <div class="actions">
    {#if cfg.mode !== 'off'}
      <button type="button" class="ghost" on:click={test} disabled={testing || saving}>
        {testing ? $t('wizard.testing') : $t('network.proxy.test')}
      </button>
    {/if}
    <button type="button" class="primary" on:click={save} disabled={saving}>
      <IconCheck size={14} stroke={2} />
      {saving ? $t('mailboxes.saving') : $t('mailboxes.save')}
    </button>
  </div>
{/if}

<style>
  h3 {
    margin: 0 0 var(--space-3);
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .hint {
    margin: 0 0 var(--space-4);
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .manual {
    margin-top: var(--space-2);
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

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-3);
    margin-top: var(--space-4);
  }

  .primary,
  .ghost {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
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

  .primary:disabled,
  .ghost:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .ghost {
    background: var(--surface-raised);
    color: var(--text-primary);
  }
</style>
