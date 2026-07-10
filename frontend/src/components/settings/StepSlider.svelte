<script lang="ts">
  // a labelled slider over a fixed set of discrete steps, used where a setting
  // has an ordered range of presets (e.g. the auto-sync interval) and a long
  // segmented control would wrap. the track snaps to each step; the current
  // step's label is shown above and the end labels anchor the range.
  import { createEventDispatcher } from 'svelte'

  /** Setting label shown above the slider. */
  export let label: string
  /** The currently selected option key. */
  export let value: string
  /** Ordered steps; the slider index maps to these in order. */
  export let options: { key: string; label: string }[]

  const dispatch = createEventDispatcher<{ change: string }>()

  // map between the option key and the slider's numeric index. an unknown value
  // falls back to the first step so the thumb always has a valid position.
  $: index = Math.max(
    0,
    options.findIndex((o) => o.key === value),
  )
  $: current = options[index] ?? options[0]

  function onInput(event: Event): void {
    const i = Number((event.currentTarget as HTMLInputElement).value)
    const opt = options[i]
    if (opt && opt.key !== value) {
      dispatch('change', opt.key)
    }
  }
</script>

<div class="setting">
  <div class="head">
    <span class="label">{label}</span>
    <span class="value">{current?.label}</span>
  </div>
  <input
    type="range"
    min="0"
    max={options.length - 1}
    step="1"
    value={index}
    aria-label={label}
    aria-valuetext={current?.label}
    on:input={onInput}
  />
  <div class="ends">
    <span>{options[0]?.label}</span>
    <span>{options[options.length - 1]?.label}</span>
  </div>
</div>

<style>
  .setting {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-3) 0;
  }

  .head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-4);
  }

  .label {
    font-size: var(--fz-body);
    color: var(--text-primary);
  }

  .value {
    font-size: var(--fz-label);
    font-weight: var(--fw-semibold);
    color: var(--accent);
    font-variant-numeric: tabular-nums;
  }

  input[type='range'] {
    -webkit-appearance: none;
    appearance: none;
    width: 100%;
    height: 4px;
    border-radius: 999px;
    background: var(--surface-hover);
    cursor: pointer;
  }

  input[type='range']:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 4px;
  }

  input[type='range']::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 16px;
    height: 16px;
    border-radius: 999px;
    background: var(--accent);
    border: 2px solid var(--surface-raised);
    box-shadow: var(--shadow-overlay);
  }

  input[type='range']::-moz-range-thumb {
    width: 16px;
    height: 16px;
    border-radius: 999px;
    background: var(--accent);
    border: 2px solid var(--surface-raised);
  }

  .ends {
    display: flex;
    justify-content: space-between;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    font-variant-numeric: tabular-nums;
  }
</style>
