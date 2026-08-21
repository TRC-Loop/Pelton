<script lang="ts">
  // a small info glyph next to a field label. The browser's own title tooltip
  // takes about a second to appear, which is long enough that people give up
  // on it, so the bubble here is drawn by the app and shows in 150ms.
  import { IconInfoCircle } from '@tabler/icons-svelte'

  /** The explanation the bubble shows. */
  export let text: string
  /** Which side of the glyph the bubble opens towards. */
  export let align: 'left' | 'right' = 'left'
</script>

<button type="button" class="tip {align}" aria-label={text}>
  <IconInfoCircle size={13} stroke={1.7} />
  <span class="bubble" aria-hidden="true">{text}</span>
</button>

<style>
  .tip {
    position: relative;
    display: inline-flex;
    padding: 0;
    border: none;
    background: transparent;
    vertical-align: -2px;
    color: var(--text-tertiary);
    cursor: help;
  }
  .tip:hover,
  .tip:focus-visible {
    color: var(--text-secondary);
  }

  .bubble {
    position: absolute;
    top: calc(100% + var(--space-1));
    z-index: 400;
    width: max-content;
    max-width: 260px;
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
    color: var(--text-secondary);
    font-size: var(--fz-meta);
    font-weight: var(--fw-regular);
    line-height: 1.45;
    text-align: left;
    white-space: normal;
    opacity: 0;
    visibility: hidden;
    transition:
      opacity 80ms ease 150ms,
      visibility 0s linear 150ms;
  }
  .left .bubble {
    left: 0;
  }
  .right .bubble {
    right: 0;
  }

  .tip:hover .bubble,
  .tip:focus-visible .bubble {
    opacity: 1;
    visibility: visible;
  }
</style>
