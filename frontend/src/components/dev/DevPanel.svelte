<script lang="ts">
  // the shared chrome for a developer overlay: a title strip with the key that
  // opened it, a close button, and a body that scrolls.
  //
  // Panels are draggable by their title strip so two of them can be arranged
  // side by side. Nothing about the position is stored; the overlays are
  // scratch tools and every launch starts them closed at their default corner.
  import { IconX } from '@tabler/icons-svelte'

  /** Panel heading. */
  export let title: string
  /** The F-key that toggles this panel, shown next to the title. */
  export let shortcut: string
  /** Distance from the top of the window, as a css length. */
  export let top = 'var(--space-5)'
  /** Distance from the left of the window, as a css length. */
  export let left = 'var(--space-5)'
  /** Panel width, as a css length. */
  export let width = '360px'
  /** Called when the close button is pressed. */
  export let onClose: () => void = () => {}

  let dragging = false
  let offsetX = 0
  let offsetY = 0
  let position: { top: string; left: string } | null = null

  $: placement = position ?? { top, left }

  function startDrag(event: PointerEvent): void {
    const panel = (event.currentTarget as HTMLElement).parentElement
    if (!panel) {
      return
    }
    const box = panel.getBoundingClientRect()
    offsetX = event.clientX - box.left
    offsetY = event.clientY - box.top
    dragging = true
    ;(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId)
  }

  function drag(event: PointerEvent): void {
    if (!dragging) {
      return
    }
    position = {
      left: `${Math.max(0, event.clientX - offsetX)}px`,
      top: `${Math.max(0, event.clientY - offsetY)}px`,
    }
  }

  function endDrag(event: PointerEvent): void {
    dragging = false
    ;(event.currentTarget as HTMLElement).releasePointerCapture(event.pointerId)
  }
</script>

<section class="panel" style="top: {placement.top}; left: {placement.left}; width: {width};">
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <header
    class="bar"
    class:dragging
    on:pointerdown={startDrag}
    on:pointermove={drag}
    on:pointerup={endDrag}
    on:pointercancel={endDrag}
  >
    <span class="title">{title}</span>
    <span class="key">{shortcut}</span>
    <button type="button" class="close" aria-label="Close {title}" on:click={onClose}>
      <IconX size={13} stroke={2} />
    </button>
  </header>
  <div class="body">
    <slot />
  </div>
</section>

<style>
  .panel {
    position: fixed;
    z-index: 400;
    display: flex;
    flex-direction: column;
    max-height: 60vh;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
    overflow: hidden;
    resize: both;
  }

  .bar {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-2) var(--space-2) var(--space-3);
    border-bottom: var(--hairline) solid var(--border-subtle);
    background: var(--surface-sunken);
    cursor: grab;
    user-select: none;
  }
  .bar.dragging {
    cursor: grabbing;
  }

  .title {
    flex: 1;
    min-width: 0;
    font-size: var(--fz-label);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .key {
    padding: 0 var(--space-1);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .close {
    display: inline-flex;
    padding: var(--space-1);
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    border-radius: var(--radius-control);
    cursor: var(--cursor-action);
  }
  .close:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .body {
    overflow: auto;
    padding: var(--space-3);
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }
</style>
