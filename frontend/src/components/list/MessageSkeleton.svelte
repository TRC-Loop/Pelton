<script lang="ts">
  // placeholder rows while a list loads (#320). Nothing appears for the first
  // 200ms: a page query is under a millisecond on a large mailbox, so showing a
  // placeholder for every load would put a flash in front of work that had
  // already finished. Only a load slow enough to be felt gets one, and it fades
  // in so its arrival is not itself a jolt.
  import { onDestroy } from 'svelte'
  import { prefs } from '../../stores/prefs'

  /** How many placeholder rows to draw. */
  export let rows = 8
  /** Set while something is loading. Clearing it hides the placeholders at once. */
  export let active = false
  /** Extra top border, for placeholders appended under real rows. */
  export let inline = false

  const delayMs = 200

  let show = false
  let timer: ReturnType<typeof setTimeout> | null = null

  $: schedule(active)

  function schedule(on: boolean): void {
    if (!on) {
      cancel()
      show = false
      return
    }
    if (timer || show) {
      return
    }
    timer = setTimeout(() => {
      timer = null
      show = true
    }, delayMs)
  }

  function cancel(): void {
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  onDestroy(cancel)

  // the fade is the only motion here, so honouring reduce-motion means
  // appearing outright rather than not appearing at all.
  $: fade = $prefs.reduceMotion ? 0 : 150
</script>

{#if show}
  <div
    class="skeleton"
    class:inline
    style={`--fade:${fade}ms`}
    role="status"
    aria-live="polite"
    aria-busy="true"
  >
    {#each Array(rows) as _, i (i)}
      <div class="row">
        <div class="line subject"></div>
        <div class="line snippet"></div>
      </div>
    {/each}
  </div>
{/if}

<style>
  .skeleton {
    display: flex;
    flex-direction: column;
    animation: appear var(--fade) ease-out both;
  }

  .skeleton.inline {
    border-top: var(--hairline) solid var(--border-subtle);
  }

  .row {
    display: flex;
    flex-direction: column;
    gap: var(--row-gap);
    padding: var(--row-pad-y) var(--row-pad-x);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .line {
    height: 9px;
    border-radius: var(--radius-control);
    background: var(--surface-hover);
  }

  .subject {
    width: 62%;
  }

  .snippet {
    width: 88%;
    opacity: 0.6;
  }

  @keyframes appear {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
</style>
