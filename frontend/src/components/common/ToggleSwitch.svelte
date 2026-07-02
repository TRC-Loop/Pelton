<script lang="ts">
  // a small accessible on/off switch used across settings in place of native
  // checkboxes. it dispatches `change` with the new boolean; the parent owns the
  // state. disabled switches are dimmed and inert.
  import { createEventDispatcher } from 'svelte'

  export let checked: boolean = false
  export let disabled: boolean = false
  export let label: string = ''

  const dispatch = createEventDispatcher<{ change: boolean }>()

  function toggle(): void {
    if (disabled) {
      return
    }
    dispatch('change', !checked)
  }
</script>

<button
  type="button"
  role="switch"
  class="switch"
  class:on={checked}
  aria-checked={checked}
  aria-label={label || undefined}
  {disabled}
  on:click={toggle}
>
  <span class="knob" aria-hidden="true"></span>
</button>

<style>
  .switch {
    flex-shrink: 0;
    position: relative;
    width: 36px;
    height: 20px;
    padding: 0;
    /* the off state needs enough contrast to read as a control: a visible filled
       track plus a defined border, not just a faint sunken well. */
    border: var(--hairline) solid var(--text-tertiary);
    border-radius: 999px;
    background: var(--text-tertiary);
    cursor: pointer;
    transition: background 0.14s ease, border-color 0.14s ease;
  }

  .switch.on {
    background: var(--accent);
    border-color: var(--accent);
  }

  .switch:disabled {
    opacity: 0.45;
    cursor: default;
  }

  .knob {
    position: absolute;
    top: 50%;
    left: 2px;
    width: 14px;
    height: 14px;
    border-radius: 999px;
    background: var(--surface-base);
    transform: translateY(-50%);
    transition: left 0.14s ease;
    box-shadow: 0 1px 2px rgb(0 0 0 / 0.25);
  }

  .switch.on .knob {
    left: 18px;
    background: var(--accent-fg, #fff);
  }
</style>
