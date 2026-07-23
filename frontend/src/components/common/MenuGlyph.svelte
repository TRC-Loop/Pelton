<script lang="ts">
  // MenuGlyph renders a custom menu item's chosen icon by name or geometry,
  // without pulling in the full tabler dataset. A theme override for the name
  // wins (so it restyles live with the theme); otherwise the stored tabler node
  // geometry renders as an inline svg matching tabler's outline defaults.
  import { iconOverrides } from '../../theme/icons'
  import type { IconNode } from '../../lib/menuactions'

  /** theme override name to prefer, if the active theme ships one. */
  export let iconName: string | undefined = undefined
  /** tabler icon geometry to render when no theme override applies. */
  export let iconNodes: IconNode[] | undefined = undefined
  /** icon box in px. */
  export let size = 15
  /** stroke width for the tabler geometry. */
  export let stroke = 1.7

  $: override = iconName ? $iconOverrides[iconName] : undefined
</script>

{#if override}
  <!-- eslint-disable-next-line svelte/no-at-html-tags -- svg sanitized by the backend -->
  <span class="glyph" style="width:{size}px;height:{size}px" aria-hidden="true">{@html override}</span>
{:else if iconNodes}
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width={stroke}
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    {#each iconNodes as [tag, attrs]}
      <svelte:element this={tag} {...attrs} />
    {/each}
  </svg>
{/if}

<style>
  .glyph {
    display: inline-flex;
    flex: none;
  }

  .glyph :global(svg) {
    width: 100%;
    height: 100%;
  }
</style>
