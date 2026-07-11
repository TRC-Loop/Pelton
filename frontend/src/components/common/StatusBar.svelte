<script lang="ts">
  // the bottom status line. the left side surfaces the durable send queue: a
  // clickable "N sending" that opens the outbox panel, plus a failed count. the
  // right side shows live sync state and the last-synced relative time. it is the
  // honest, always-visible window into background activity.
  import { onDestroy, onMount } from 'svelte'
  import { IconSend, IconAlertTriangle, IconRefresh, IconCheck, IconDownload, IconBatteryEco, IconX, IconBug } from '@tabler/icons-svelte'
  import { outbox, syncing, lastSynced } from '../../stores/outbox'
  import { downloadProgress, attachmentProgress } from '../../stores/progress'
  import { formatRelative } from '../../lib/format'
  import { cancelDownload, isDevMode } from '../../lib/api'
  import { prefs, setLowPowerMode } from '../../stores/prefs'
  import OutboxPanel from './OutboxPanel.svelte'
  import { t } from '../../lib/i18n'

  // devMode is read once at startup: it's fixed for the lifetime of the
  // process (set by the PELTON_DEV env var a dev run launches with), not
  // something that can change while the app is open.
  let devMode = false
  onMount(async () => {
    devMode = await isDevMode().catch(() => false)
  })

  // format an eta in seconds as m:ss (or just seconds under a minute).
  function formatEta(sec: number): string {
    if (sec <= 0) {
      return ''
    }
    if (sec < 60) {
      return `${sec}s`
    }
    const m = Math.floor(sec / 60)
    const s = sec % 60
    return `${m}:${String(s).padStart(2, '0')}`
  }

  // attachment percent from bytes, guarding a zero total.
  $: attPercent =
    $attachmentProgress && $attachmentProgress.bytesTotal > 0
      ? Math.round(($attachmentProgress.bytesDone / $attachmentProgress.bytesTotal) * 100)
      : 0

  let panelOpen = false

  $: pending = $outbox.filter((r) => r.state === 'queued' || r.state === 'sending')
  $: failed = $outbox.filter((r) => r.state === 'failed')

  // re-render the relative time on a slow tick so "2m ago" stays current. the
  // tick variable is referenced in the reactive label so it recomputes.
  let tick = Date.now()
  const timer = setInterval(() => (tick = Date.now()), 30000)
  onDestroy(() => clearInterval(timer))
  $: syncedLabel = $lastSynced ? relativeAt($lastSynced, tick) : ''
  function relativeAt(ts: number, _tick: number): string {
    return formatRelative(ts)
  }
</script>

<footer class="statusbar">
  <div class="left">
    {#if devMode}
      <span class="dev-badge" title={$t('common.statusBar.devModeTitle')}>
        <IconBug size={13} stroke={1.8} />
        {$t('common.statusBar.devMode')}
      </span>
    {/if}
    {#if pending.length > 0 || failed.length > 0}
      <button type="button" class="outbox-btn" class:has-failed={failed.length > 0} on:click={() => (panelOpen = !panelOpen)}>
        {#if pending.length > 0}
          <IconSend size={13} stroke={1.7} />
          <span>{pending.length} {$t('common.outbox.sendingSuffix')}</span>
        {/if}
        {#if failed.length > 0}
          <IconAlertTriangle size={13} stroke={1.7} class="fail-icon" />
          <span class="fail">{failed.length} {$t('common.outbox.failedSuffix')}</span>
        {/if}
      </button>
    {/if}

    {#if panelOpen}
      <div class="popover">
        <OutboxPanel on:close={() => (panelOpen = false)} />
      </div>
    {/if}
  </div>

  <div class="center">
    {#if $downloadProgress}
      <div class="progress" title={$downloadProgress.label}>
        <IconDownload size={12} stroke={1.7} />
        <span class="p-label">
          {#if $downloadProgress.error}
            {$t('common.statusBar.downloadFailed')}
          {:else if $downloadProgress.total > 0}
            {$downloadProgress.done}/{$downloadProgress.total}
          {:else}
            {$downloadProgress.label || $t('common.statusBar.downloading')}
          {/if}
        </span>
        {#if $downloadProgress.total > 0 && !$downloadProgress.error}
          <span class="bar"><span class="fill" style={`width:${$downloadProgress.percent}%`}></span></span>
          <span class="p-num">{$downloadProgress.percent}%</span>
          {#if $downloadProgress.etaSeconds > 0}
            <span class="p-eta">~{formatEta($downloadProgress.etaSeconds)}</span>
          {/if}
        {/if}
        {#if $downloadProgress.running}
          <button
            type="button"
            class="cancel-dl"
            aria-label={$t('common.statusBar.cancelDownload')}
            title={$t('common.statusBar.cancelDownload')}
            on:click={() => cancelDownload()}
          >
            <IconX size={12} stroke={2} />
          </button>
        {/if}
      </div>
    {:else if $attachmentProgress}
      <div class="progress">
        <IconDownload size={12} stroke={1.7} />
        <span class="p-label">
          {#if $attachmentProgress.filesTotal > 1}
            {$t('common.statusBar.saving')} {$attachmentProgress.filesDone + 1}/{$attachmentProgress.filesTotal}
          {:else}
            {$t('common.statusBar.saving')}
          {/if}
        </span>
        <span class="bar"><span class="fill" style={`width:${attPercent}%`}></span></span>
      </div>
    {/if}
  </div>

  <div class="right">
    {#if $prefs.lowPowerMode}
      <button
        type="button"
        class="low-power"
        title={$t('common.statusBar.lowPowerTitle')}
        on:click={() => setLowPowerMode(false)}
      >
        <IconBatteryEco size={13} stroke={1.7} />
        {$t('common.statusBar.lowPower')}
      </button>
    {/if}
    {#if $syncing}
      <span class="sync syncing">
        <IconRefresh size={13} stroke={1.7} class="spin" />
        {$t('common.statusBar.syncing')}
      </span>
    {:else if $lastSynced}
      <span class="sync">
        <IconCheck size={13} stroke={1.7} />
        {$t('common.statusBar.synced')} {syncedLabel}
      </span>
    {/if}
  </div>
</footer>

{#if panelOpen}
  <!-- click-away closes the popover. -->
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="scrim" on:click={() => (panelOpen = false)}></div>
{/if}

<style>
  .statusbar {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    height: 26px;
    padding: 0 var(--space-4);
    background: var(--surface-sunken);
    border-top: var(--hairline) solid var(--border-subtle);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .left,
  .right {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    flex: 1;
    min-width: 0;
  }
  .right {
    justify-content: flex-end;
  }

  .center {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .progress {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--text-secondary);
    white-space: nowrap;
  }

  .bar {
    width: 120px;
    height: 5px;
    border-radius: 999px;
    background: var(--surface-hover);
    overflow: hidden;
  }
  .fill {
    display: block;
    height: 100%;
    background: var(--accent);
    border-radius: 999px;
    transition: width 0.2s ease;
  }
  .p-num {
    font-variant-numeric: tabular-nums;
    color: var(--text-primary);
  }
  .p-eta {
    color: var(--text-tertiary);
    font-variant-numeric: tabular-nums;
  }

  .cancel-dl {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: 1px;
    border-radius: var(--radius-control);
  }
  .cancel-dl:hover {
    background: var(--surface-hover);
    color: var(--danger);
  }

  .outbox-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    border: none;
    background: transparent;
    color: var(--text-secondary);
    font-size: var(--fz-meta);
    cursor: pointer;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-control);
  }

  .outbox-btn:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .outbox-btn .fail {
    color: var(--danger);
  }

  .outbox-btn :global(.fail-icon) {
    color: var(--danger);
  }

  .popover {
    position: absolute;
    bottom: calc(100% + 6px);
    left: var(--space-3);
    z-index: 90;
  }

  .scrim {
    position: fixed;
    inset: 0;
    z-index: 80;
  }

  .low-power {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    border: none;
    background: transparent;
    color: var(--warning, var(--text-secondary));
    font-size: var(--fz-meta);
    cursor: pointer;
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-control);
  }

  .dev-badge {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) var(--space-2);
    border-radius: var(--radius-control);
    background: var(--danger-bg, var(--warning-bg, var(--surface-sunken)));
    color: var(--danger, var(--warning));
    font-size: var(--fz-meta);
    font-weight: var(--fw-semibold);
    letter-spacing: 0.02em;
  }

  .low-power:hover {
    background: var(--surface-hover);
  }

  .sync {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
  }

  .sync.syncing {
    color: var(--text-secondary);
  }

  .sync :global(.spin) {
    /* svg transform-origin defaults differ across the webviews wails embeds
       per platform (webkit vs webview2); pin it explicitly so the icon spins
       in place instead of wobbling around an off-center pivot on some os. */
    transform-box: border-box;
    transform-origin: 50% 50%;
    animation: statusspin 0.8s linear infinite;
  }

  @keyframes statusspin {
    to {
      transform: rotate(-360deg);
    }
  }
</style>
