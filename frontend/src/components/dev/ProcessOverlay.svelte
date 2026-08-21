<script lang="ts">
  // the Go side of the app, sampled: goroutines, heap, gc runs, and what the
  // app has put on disk. Useful for the two things that are otherwise invisible
  // from the ui, a goroutine that never exits and a cache that keeps growing.
  import { onDestroy, onMount } from 'svelte'
  import DevPanel from './DevPanel.svelte'
  import { closeOverlay } from '../../stores/devoverlays'
  import { devProcessStats } from '../../lib/api'
  import type { DevProcessStats } from '../../lib/types'

  const pollMs = 2000

  let stats: DevProcessStats | null = null
  let error = ''
  let timer: ReturnType<typeof setInterval> | null = null

  onMount(() => {
    void poll()
    timer = setInterval(() => void poll(), pollMs)
  })

  onDestroy(() => {
    if (timer) {
      clearInterval(timer)
    }
  })

  async function poll(): Promise<void> {
    try {
      stats = await devProcessStats()
      error = ''
    } catch (err) {
      error = err instanceof Error ? err.message : String(err)
    }
  }

  // bytes at three digits of precision, which is all any of these numbers
  // deserve while they are moving.
  function bytes(n: number): string {
    const units = ['B', 'KB', 'MB', 'GB']
    let value = n
    let unit = 0
    while (value >= 1024 && unit < units.length - 1) {
      value /= 1024
      unit++
    }
    return `${value < 10 && unit > 0 ? value.toFixed(1) : Math.round(value)} ${units[unit]}`
  }

  function duration(seconds: number): string {
    const h = Math.floor(seconds / 3600)
    const m = Math.floor((seconds % 3600) / 60)
    const s = seconds % 60
    return h > 0 ? `${h}h ${m}m` : m > 0 ? `${m}m ${s}s` : `${s}s`
  }
</script>

<DevPanel title="Process" shortcut="F7" top="var(--space-5)" left="calc(100vw - 300px)" width="260px" onClose={() => closeOverlay('process')}>
  {#if error}
    <p class="error">{error}</p>
  {:else if !stats}
    <p class="muted">Sampling…</p>
  {:else}
    <dl>
      <dt>goroutines</dt>
      <dd>{stats.goroutines}</dd>
      <dt>heap live</dt>
      <dd>{bytes(stats.heapBytes)}</dd>
      <dt>heap from os</dt>
      <dd>{bytes(stats.heapSysBytes)}</dd>
      <dt>gc runs</dt>
      <dd>{stats.gcRuns}</dd>
      <dt>database</dt>
      <dd>{bytes(stats.databaseBytes)}</dd>
      <dt>attachments</dt>
      <dd>{bytes(stats.attachmentsBytes)}</dd>
      <dt>data dir</dt>
      <dd>{bytes(stats.dataDirBytes)}</dd>
      <dt>uptime</dt>
      <dd>{duration(stats.uptimeSeconds)}</dd>
    </dl>
    <p class="muted">Directory sizes are measured every 10s.</p>
  {/if}
</DevPanel>

<style>
  dl {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: var(--space-1) var(--space-3);
    margin: 0;
  }

  dt {
    color: var(--text-tertiary);
  }

  dd {
    margin: 0;
    text-align: right;
    color: var(--text-primary);
    font-variant-numeric: tabular-nums;
  }

  .muted {
    margin: var(--space-3) 0 0;
    color: var(--text-tertiary);
  }

  .error {
    margin: 0;
    color: var(--danger);
  }
</style>
