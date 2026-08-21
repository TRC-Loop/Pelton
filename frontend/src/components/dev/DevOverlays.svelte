<script lang="ts">
  // the developer overlays (#188), and the F-keys that toggle them.
  //
  // The whole thing renders nothing unless the backend says the overlays are
  // available, which it only does for a run started with PELTON_DEV or
  // PELTON_DEVTOOLS. A release build never binds the keys.
  //
  // The panels are code-split: a normal build downloads none of this.
  import { onMount } from 'svelte'
  import {
    devToolsAvailable,
    openOverlays,
    initDevTools,
    handleOverlayKey,
  } from '../../stores/devoverlays'

  onMount(initDevTools)

  function onKeydown(event: KeyboardEvent): void {
    if (handleOverlayKey(event)) {
      event.preventDefault()
    }
  }
</script>

<svelte:window on:keydown={onKeydown} />

{#if $devToolsAvailable}
  {#if $openOverlays.has('activity')}
    {#await import('./ActivityOverlay.svelte') then m}
      <svelte:component this={m.default} />
    {/await}
  {/if}
  {#if $openOverlays.has('process')}
    {#await import('./ProcessOverlay.svelte') then m}
      <svelte:component this={m.default} />
    {/await}
  {/if}
  {#if $openOverlays.has('performance')}
    {#await import('./PerformanceOverlay.svelte') then m}
      <svelte:component this={m.default} />
    {/await}
  {/if}
{/if}
