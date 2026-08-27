<script lang="ts">
  // frame timing for the ui.
  //
  // A plain fps readout in a webview is close to useless: at rest it sits at
  // whatever the display refreshes at and says the same thing on a fast machine
  // and a slow one. What is worth watching is the opposite, the frames that
  // took too long, so this reports the worst frame in the last second and
  // counts every frame that missed its slot. Those numbers move when a list
  // scroll janks, which is the thing you would open this to see.
  import { onDestroy, onMount } from 'svelte'
  import DevPanel from './DevPanel.svelte'
  import { closeOverlay } from '../../stores/devoverlays'

  // how many recent frame times the strip shows.
  const sampleCount = 60

  let frames: number[] = []
  let fps = 0
  let worst = 0
  let longFrames = 0
  let raf = 0
  let last = 0
  // the display's own frame budget, learned from the fastest frame seen rather
  // than assumed to be 60Hz, so a 120Hz screen is not judged against 16.7ms.
  let budget = 16.7

  onMount(() => {
    last = performance.now()
    raf = requestAnimationFrame(tick)
  })

  onDestroy(() => cancelAnimationFrame(raf))

  function tick(now: number): void {
    const delta = now - last
    last = now
    raf = requestAnimationFrame(tick)

    // the first frame after the overlay opens carries the mount cost and says
    // nothing about the ui.
    if (delta <= 0 || delta > 2000) {
      return
    }
    if (delta < budget) {
      budget = Math.max(delta, 6)
    }
    frames = [...frames, delta].slice(-sampleCount)

    const total = frames.reduce((sum, f) => sum + f, 0)
    fps = Math.round(1000 / (total / frames.length))
    worst = Math.max(...frames)
    if (delta > budget * 1.8) {
      longFrames++
    }
  }

  function reset(): void {
    frames = []
    longFrames = 0
    worst = 0
  }

  // the strip is scaled against twice the budget, so a frame that missed its
  // slot fills more than half the bar and stands out without a legend.
  $: scale = budget * 2
</script>

<DevPanel title="Frames" shortcut="F8" top="calc(var(--space-5) + 300px)" left="calc(100vw - 300px)" width="260px" onClose={() => closeOverlay('performance')}>
  <dl>
    <dt>fps</dt>
    <dd>{fps}</dd>
    <dt>budget</dt>
    <dd>{budget.toFixed(1)} ms</dd>
    <dt>worst frame</dt>
    <dd>{worst.toFixed(1)} ms</dd>
    <dt>missed</dt>
    <dd>{longFrames}</dd>
  </dl>

  <div class="strip" aria-hidden="true">
    {#each frames as frame, i (i)}
      <span class="bar" class:over={frame > budget * 1.8} style="height: {Math.min(100, (frame / scale) * 100)}%"></span>
    {/each}
  </div>

  <button type="button" on:click={reset}>reset</button>
  <p class="muted">Idle sits at the refresh rate. Scroll a message list to see it work.</p>
</DevPanel>

<style>
  dl {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: var(--space-1) var(--space-3);
    margin: 0 0 var(--space-3);
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

  .strip {
    display: flex;
    align-items: flex-end;
    gap: 1px;
    height: 40px;
    padding: var(--space-1);
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
  }

  .bar {
    flex: 1;
    min-height: 1px;
    background: var(--accent);
    border-radius: 1px;
  }

  .bar.over {
    background: var(--warning);
  }

  button {
    margin-top: var(--space-2);
    padding: 0 var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: transparent;
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
    cursor: var(--cursor-action);
  }
  button:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .muted {
    margin: var(--space-2) 0 0;
    color: var(--text-tertiary);
  }
</style>
