<script lang="ts">
  // a live view of the app's log: sync passes, imap and smtp errors, outbox
  // activity, everything the backend already writes. It reads the same redacted
  // lines that reach stderr and the log file, so a stored password cannot show
  // up here either.
  //
  // What it can show is limited to what is logged at the current level. Below
  // debug most of the sync path says nothing, so the level is on screen and
  // PELTON_DEBUG is how you get more.
  import { onDestroy, onMount, tick } from 'svelte'
  import DevPanel from './DevPanel.svelte'
  import { closeOverlay } from '../../stores/devoverlays'
  import { devActivity, clearDevActivity } from '../../lib/api'
  import type { DevLogLine } from '../../lib/types'

  // how often new lines are pulled. Fast enough to feel live, slow enough that
  // the poll itself is not what the overlay is showing you.
  const pollMs = 700
  // how many lines stay on screen. The backend buffer is the real limit; this
  // keeps the dom from growing past it.
  const maxLines = 500

  let lines: DevLogLine[] = []
  let level = ''
  let error = ''
  let next = 0
  let follow = true
  let body: HTMLElement | null = null
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
      const page = await devActivity(next)
      error = ''
      level = page.level
      next = page.next
      if (page.lines.length > 0) {
        lines = [...lines, ...page.lines].slice(-maxLines)
        if (follow) {
          await tick()
          body?.scrollTo({ top: body.scrollHeight })
        }
      }
    } catch (err) {
      error = err instanceof Error ? err.message : String(err)
    }
  }

  async function clear(): Promise<void> {
    try {
      await clearDevActivity()
      lines = []
    } catch (err) {
      error = err instanceof Error ? err.message : String(err)
    }
  }

  // the seq numbers tell whether the buffer dropped lines while the overlay was
  // closed or busy, which matters when reading back what happened.
  $: dropped = lines.length > 0 && lines[0].seq > 0 && lines.length < lines[lines.length - 1].seq + 1
</script>

<DevPanel title="Activity" shortcut="F6" top="var(--space-5)" left="var(--space-5)" width="520px" onClose={() => closeOverlay('activity')}>
  <div class="controls">
    <span class="level">level {level || '?'}</span>
    <label class="follow">
      <input type="checkbox" bind:checked={follow} />
      follow
    </label>
    <button type="button" on:click={clear}>clear</button>
  </div>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="log" bind:this={body}>
    {#if lines.length === 0}
      <p class="empty">Nothing logged yet. Set PELTON_DEBUG=1 for the full sync detail.</p>
    {:else}
      {#if dropped}
        <p class="empty">Older lines were dropped from the buffer.</p>
      {/if}
      {#each lines as line (line.seq)}
        <div class="line">{line.text}</div>
      {/each}
    {/if}
  </div>
</DevPanel>

<style>
  .controls {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-2);
  }

  .level {
    color: var(--text-tertiary);
  }

  .follow {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    cursor: var(--cursor-action);
  }

  button {
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

  .log {
    max-height: 40vh;
    overflow: auto;
  }

  .line {
    padding: 1px 0;
    white-space: pre-wrap;
    word-break: break-word;
    line-height: 1.45;
    border-bottom: var(--hairline) solid var(--border-subtle);
    color: var(--text-secondary);
  }

  .empty,
  .error {
    margin: 0 0 var(--space-2);
    color: var(--text-tertiary);
  }

  .error {
    color: var(--danger);
  }
</style>
