<script lang="ts">
  // a labelled segmented control used by the theme and density settings. generic
  // over the option value so each setting stays a one-liner.
  import { createEventDispatcher } from 'svelte'

  export let label: string
  export let value: string
  export let options: { key: string; label: string }[]

  const dispatch = createEventDispatcher<{ change: string }>()
</script>

<div class="setting">
  <span class="label">{label}</span>
  <div class="segments" role="group" aria-label={label}>
    {#each options as opt (opt.key)}
      <button
        type="button"
        class:active={value === opt.key}
        aria-pressed={value === opt.key}
        on:click={() => dispatch('change', opt.key)}
      >
        {opt.label}
      </button>
    {/each}
  </div>
</div>

<style>
  .setting {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-3) 0;
  }

  .label {
    font-size: var(--fz-body);
    color: var(--text-primary);
  }

  .segments {
    display: inline-flex;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    overflow: hidden;
  }

  button {
    border: none;
    border-right: var(--hairline) solid var(--border-subtle);
    background: var(--surface-raised);
    color: var(--text-secondary);
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  button:last-child {
    border-right: none;
  }

  button:hover {
    background: var(--surface-hover);
  }

  button.active {
    background: var(--selection-bg);
    color: var(--text-primary);
  }
</style>
