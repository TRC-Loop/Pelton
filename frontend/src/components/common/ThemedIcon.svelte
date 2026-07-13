<script lang="ts">
  // ThemedIcon renders an icon that a custom theme may override: if the
  // active theme ships an svg for `name`, that svg is inlined (it inherits
  // currentColor like the bundled icons); otherwise the given tabler
  // component renders as before. Call sites convert to this wrapper
  // progressively - every converted site becomes themeable.
  import type { ComponentType } from 'svelte'
  import { iconOverrides } from '../../theme/icons'

  /** icon name a theme can target: the tabler name without the Icon prefix,
   * lowercase kebab (IconPalette -> "palette"). */
  export let name: string
  /** bundled tabler icon component used when no override exists. */
  export let icon: ComponentType
  /** icon box in px, matching the tabler size prop. */
  export let size = 16
  /** stroke width passed to the tabler icon (overrides bring their own). */
  export let stroke = 1.6
</script>

{#if $iconOverrides[name]}
  <!-- eslint-disable-next-line svelte/no-at-html-tags -- svg sanitized by the backend -->
  <span class="themed-icon" style="width:{size}px;height:{size}px" aria-hidden="true">{@html $iconOverrides[name]}</span>
{:else}
  <svelte:component this={icon} {size} {stroke} />
{/if}

<style>
  .themed-icon {
    display: inline-flex;
    flex: none;
  }

  .themed-icon :global(svg) {
    width: 100%;
    height: 100%;
  }
</style>
