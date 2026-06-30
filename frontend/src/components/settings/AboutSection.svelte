<script lang="ts">
  // the about block at the bottom of settings: app name, version and links. links
  // open in the system browser via the wails runtime, never an in-app webview.
  import { onMount } from 'svelte'
  import { IconBrandGithub, IconBug, IconLicense, IconX } from '@tabler/icons-svelte'
  import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'
  import { APP } from '../../lib/app-info'
  import { appVersion } from '../../lib/api'

  // the version is injected into the binary at build time; fall back to the
  // static value if the binding is unavailable (e.g. a stale dev build).
  let version: string = APP.version
  onMount(async () => {
    try {
      const v = await appVersion()
      if (v) {
        version = v
      }
    } catch {
      // keep the static fallback.
    }
  })

  // the licenses view is heavy (it pulls the embedded manifest), so it is
  // code-split and only mounted (in a modal) when the user opens it.
  let showLicenses = false

  function open(url: string): void {
    BrowserOpenURL(url)
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      showLicenses = false
    }
  }
</script>

<svelte:window on:keydown={showLicenses ? onKeydown : undefined} />

<div class="about">
  <div class="identity">
    <span class="name">{APP.name}</span>
    <span class="version">{version.startsWith('v') ? version : `v${version}`}</span>
  </div>
  <p class="tagline">{APP.tagline}</p>

  <div class="links">
    <button type="button" on:click={() => open(APP.repo)}>
      <IconBrandGithub size={15} stroke={1.6} />
      GitHub
    </button>
    <button type="button" on:click={() => open(APP.issues)}>
      <IconBug size={15} stroke={1.6} />
      Report an issue
    </button>
    <button type="button" on:click={() => open(APP.licenseUrl)}>
      <IconLicense size={15} stroke={1.6} />
      {APP.license}
    </button>
    <button type="button" on:click={() => (showLicenses = true)}>
      <IconLicense size={15} stroke={1.6} />
      Open-source licenses
    </button>
  </div>

  <p class="credit">Made by {APP.author}. Built with Wails, Svelte and Go.</p>
</div>

{#if showLicenses}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="overlay" on:click={() => (showLicenses = false)}>
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions a11y-no-noninteractive-element-interactions -->
    <div class="modal" role="dialog" aria-modal="true" aria-label="Open-source licenses" on:click|stopPropagation>
      <header>
        <span class="m-title">Open-source licenses</span>
        <button type="button" class="m-close" aria-label="Close" on:click={() => (showLicenses = false)}>
          <IconX size={18} stroke={1.8} />
        </button>
      </header>
      <div class="m-body">
        {#await import('./LicensesView.svelte') then m}
          <svelte:component this={m.default} />
        {/await}
      </div>
    </div>
  </div>
{/if}

<style>
  .about {
    padding: var(--space-2) 0;
  }

  .identity {
    display: flex;
    align-items: baseline;
    gap: var(--space-3);
  }

  .name {
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .version {
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .tagline {
    margin: var(--space-2) 0 var(--space-4);
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }

  .links {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }

  .links button {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .links button:hover {
    background: var(--surface-hover);
  }

  .credit {
    margin: var(--space-4) 0 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .overlay {
    position: fixed;
    inset: 0;
    z-index: 140;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-5);
    background: var(--scrim, rgba(0, 0, 0, 0.4));
  }

  .modal {
    width: 100%;
    max-width: 640px;
    max-height: 82vh;
    display: flex;
    flex-direction: column;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
    overflow: hidden;
  }

  .modal header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4) var(--space-5);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .m-title {
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .m-close {
    display: inline-flex;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    padding: var(--space-1);
    border-radius: var(--radius-control);
  }

  .m-close:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .m-body {
    padding: var(--space-4) var(--space-5);
    overflow-y: auto;
  }
</style>
