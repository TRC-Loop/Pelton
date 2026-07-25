<script lang="ts">
  // the External settings: integrations that talk to services or agents outside
  // Pelton. Today that is the read-only MCP server (off by default) an AI agent
  // can connect to on loopback with a bearer token; future external integrations
  // live under the same category.
  import { onMount } from 'svelte'
  import { IconCheck, IconCopy, IconRefresh } from '@tabler/icons-svelte'
  import ToggleSwitch from '../common/ToggleSwitch.svelte'
  import { getMCPConfig, setMCPEnabled, setMCPPort, regenerateMCPToken } from '../../lib/api'
  import { errorMessage, toastError, toastSuccess } from '../../stores/toast'
  import { t } from '../../lib/i18n'
  import type { MCPConfig } from '../../lib/types'

  let cfg: MCPConfig = { enabled: false, port: 8765, token: '', url: '', running: false }
  let loading = true
  let busy = false
  // the port field is edited locally and only committed on save, so typing does
  // not restart the server on every keystroke.
  let portDraft = 8765

  onMount(async () => {
    await reload()
    loading = false
  })

  async function reload(): Promise<void> {
    try {
      cfg = await getMCPConfig()
      portDraft = cfg.port
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function onToggle(enabled: boolean): Promise<void> {
    busy = true
    try {
      await setMCPEnabled(enabled)
      await reload()
    } catch (err) {
      toastError(errorMessage(err))
      await reload()
    } finally {
      busy = false
    }
  }

  async function savePort(): Promise<void> {
    busy = true
    try {
      await setMCPPort(portDraft)
      await reload()
      toastSuccess($t('mcp.saved'))
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      busy = false
    }
  }

  async function regenerate(): Promise<void> {
    busy = true
    try {
      await regenerateMCPToken()
      await reload()
      toastSuccess($t('mcp.tokenRegenerated'))
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      busy = false
    }
  }

  async function copy(text: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(text)
      toastSuccess($t('mcp.copied'))
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // the ready-to-paste client config for an MCP host that speaks streamable HTTP.
  $: clientConfig = JSON.stringify(
    {
      mcpServers: {
        pelton: {
          url: cfg.url,
          headers: { Authorization: `Bearer ${cfg.token}` },
        },
      },
    },
    null,
    2,
  )
</script>

<h3>{$t('settingsPanel.category.external')}</h3>
<p class="hint">{$t('external.hint')}</p>

{#if loading}
  <p class="hint">{$t('mailboxes.loading')}</p>
{:else}
  <div class="card">
    <div class="card-head">
      <div>
        <span class="card-title">{$t('mcp.title')}</span>
        <span class="badge" class:on={cfg.running}>{cfg.running ? $t('mcp.running') : $t('mcp.stopped')}</span>
      </div>
      <ToggleSwitch checked={cfg.enabled} label={$t('mcp.title')} disabled={busy} on:change={(e) => onToggle(e.detail)} />
    </div>
    <p class="hint">{$t('mcp.hint')}</p>

    {#if cfg.enabled}
      <div class="servers">
        <label class="field narrow">
          <span>{$t('mcp.port')}</span>
          <input type="number" bind:value={portDraft} min="1024" max="65535" />
        </label>
        <button type="button" class="ghost" on:click={savePort} disabled={busy || portDraft === cfg.port}>
          <IconCheck size={14} stroke={2} />
          {$t('mcp.save')}
        </button>
      </div>

      <label class="field">
        <span>{$t('mcp.url')}</span>
        <div class="copy-row">
          <input type="text" readonly value={cfg.url} />
          <button type="button" class="ghost" on:click={() => copy(cfg.url)} aria-label={$t('mcp.copy')}>
            <IconCopy size={14} stroke={2} />
          </button>
        </div>
      </label>

      <label class="field">
        <span>{$t('mcp.token')}</span>
        <div class="copy-row">
          <input type="text" readonly value={cfg.token} />
          <button type="button" class="ghost" on:click={() => copy(cfg.token)} aria-label={$t('mcp.copy')}>
            <IconCopy size={14} stroke={2} />
          </button>
          <button type="button" class="ghost" on:click={regenerate} disabled={busy} aria-label={$t('mcp.regenerate')}>
            <IconRefresh size={14} stroke={2} />
          </button>
        </div>
        <p class="hint">{$t('mcp.tokenHint')}</p>
      </label>

      <div class="field">
        <span>{$t('mcp.clientConfig')}</span>
        <p class="hint">{$t('mcp.clientConfigHint')}</p>
        <pre>{clientConfig}</pre>
        <button type="button" class="ghost self-start" on:click={() => copy(clientConfig)}>
          <IconCopy size={14} stroke={2} />
          {$t('mcp.copy')}
        </button>
      </div>
    {/if}
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

  .card {
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    padding: var(--space-4);
    background: var(--surface-raised);
  }

  .card-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-2);
  }

  .card-title {
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .badge {
    margin-left: var(--space-2);
    padding: 2px var(--space-2);
    border-radius: var(--radius-pill, 999px);
    font-size: var(--fz-caption, 0.72rem);
    background: var(--surface-sunken);
    color: var(--text-tertiary);
  }

  .badge.on {
    background: color-mix(in srgb, var(--accent) 18%, transparent);
    color: var(--accent);
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

  .field input,
  .copy-row input {
    height: var(--control-height);
    padding: 0 var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    outline: none;
    width: 100%;
  }

  .field input:focus {
    border-color: var(--accent);
  }

  .servers {
    display: flex;
    align-items: flex-end;
    gap: var(--space-3);
  }

  .servers .field.narrow {
    flex: 0 0 120px;
    margin-bottom: 0;
  }

  .copy-row {
    display: flex;
    gap: var(--space-2);
  }

  pre {
    margin: 0;
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-secondary);
    font-size: var(--fz-caption, 0.75rem);
    overflow-x: auto;
    white-space: pre;
  }

  .ghost {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4);
    border-radius: var(--radius-control);
    border: var(--hairline) solid var(--border-default);
    background: var(--surface-raised);
    color: var(--text-primary);
    cursor: pointer;
    font-size: var(--fz-label);
    white-space: nowrap;
  }

  .ghost:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .self-start {
    align-self: flex-start;
  }
</style>
