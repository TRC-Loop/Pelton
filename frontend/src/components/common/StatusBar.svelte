<script lang="ts">
  // the bottom status line. the left side surfaces the durable send queue: a
  // clickable "N sending" that opens the outbox panel, plus a failed count. the
  // right side shows live sync state and the last-synced relative time. it is the
  // honest, always-visible window into background activity.
  import { onDestroy } from 'svelte'
  import { IconSend, IconAlertTriangle, IconRefresh, IconCheck } from '@tabler/icons-svelte'
  import { outbox, syncing, lastSynced } from '../../stores/outbox'
  import { formatRelative } from '../../lib/format'
  import OutboxPanel from './OutboxPanel.svelte'

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
    {#if pending.length > 0 || failed.length > 0}
      <button type="button" class="outbox-btn" class:has-failed={failed.length > 0} on:click={() => (panelOpen = !panelOpen)}>
        {#if pending.length > 0}
          <IconSend size={13} stroke={1.7} />
          <span>{pending.length} sending</span>
        {/if}
        {#if failed.length > 0}
          <IconAlertTriangle size={13} stroke={1.7} class="fail-icon" />
          <span class="fail">{failed.length} failed</span>
        {/if}
      </button>
    {/if}

    {#if panelOpen}
      <div class="popover">
        <OutboxPanel on:close={() => (panelOpen = false)} />
      </div>
    {/if}
  </div>

  <div class="right">
    {#if $syncing}
      <span class="sync syncing">
        <IconRefresh size={13} stroke={1.7} class="spin" />
        Syncing…
      </span>
    {:else if $lastSynced}
      <span class="sync">
        <IconCheck size={13} stroke={1.7} />
        Synced {syncedLabel}
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

  .sync {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
  }

  .sync.syncing {
    color: var(--text-secondary);
  }

  .sync :global(.spin) {
    animation: statusspin 0.8s linear infinite;
  }

  @keyframes statusspin {
    to {
      transform: rotate(-360deg);
    }
  }
</style>
