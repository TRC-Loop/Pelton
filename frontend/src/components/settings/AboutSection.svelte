<script lang="ts">
  // the about block at the bottom of settings: app name, version, links, and a
  // couple of dev/getting-started actions that do not deserve their own left-nav
  // category. links open in the system browser via the wails runtime, never an
  // in-app webview.
  import { onDestroy, onMount } from 'svelte'
  import { createEventDispatcher } from 'svelte'
  import { IconBrandGithub, IconBug, IconLicense, IconX, IconRefresh, IconUsers } from '@tabler/icons-svelte'
  import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'
  import { APP } from '../../lib/app-info'
  import { appVersion, checkForUpdates, type UpdateCheckResult } from '../../lib/api'
  import { onUpdateAvailable } from '../../lib/events'
  import { prefs, setUpdateCheckFrequency } from '../../stores/prefs'
  import SegmentedSetting from './SegmentedSetting.svelte'
  import { t } from '../../lib/i18n'

  const dispatch = createEventDispatcher<{ rerunOnboarding: void }>()

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

  // update checking is off by default; it only ever talks to GitHub's public
  // releases api (no telemetry, no other endpoint) to compare tags.
  $: updateCheckOptions = [
    { key: 'off', label: $t('about.update.off') },
    { key: 'startup', label: $t('about.update.startup') },
    { key: 'weekly', label: $t('about.update.weekly') },
    { key: 'monthly', label: $t('about.update.monthly') },
  ]

  let updateResult: UpdateCheckResult | null = null
  let checkingUpdate = false

  const offUpdate = onUpdateAvailable((e) => {
    updateResult = e
  })
  onDestroy(offUpdate)

  function onUpdateFrequency(event: CustomEvent<string>): void {
    setUpdateCheckFrequency(event.detail)
  }

  async function checkNow(): Promise<void> {
    checkingUpdate = true
    try {
      updateResult = await checkForUpdates()
    } catch {
      updateResult = { checked: false, available: false, currentVersion: version, latestVersion: '', releaseUrl: '', error: 'failed' }
    } finally {
      checkingUpdate = false
    }
  }

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
      {$t('about.reportIssue')}
    </button>
    <button type="button" on:click={() => open(APP.contributors)}>
      <IconUsers size={15} stroke={1.6} />
      {$t('about.contributors')}
    </button>
    <button type="button" on:click={() => open(APP.licenseUrl)}>
      <IconLicense size={15} stroke={1.6} />
      {APP.license}
    </button>
    <button type="button" on:click={() => (showLicenses = true)}>
      <IconLicense size={15} stroke={1.6} />
      {$t('about.licenses')}
    </button>
  </div>

  <div class="row">
    <span class="row-label">{$t('about.rerunTourLabel')}</span>
    <button type="button" class="action-btn" on:click={() => dispatch('rerunOnboarding')}>
      {$t('about.rerunOnboarding')}
    </button>
  </div>

  <div class="update-block">
    <SegmentedSetting label={$t('about.update.label')} value={$prefs.updateCheckFrequency} options={updateCheckOptions} on:change={onUpdateFrequency} />
    <p class="hint">{$t('about.update.hint')}</p>

    <div class="row">
      <span class="row-label">
        {#if updateResult?.checked && updateResult.available}
          {$t('about.update.available').replace('{version}', updateResult.latestVersion)}
        {:else if updateResult?.checked}
          {$t('about.update.upToDate')}
        {:else if updateResult && updateResult.error}
          {$t('about.update.checkFailed')}
        {:else}
          {$t('about.update.neverChecked')}
        {/if}
      </span>
      <div class="update-actions">
        {#if updateResult?.checked && updateResult.available}
          {@const releaseUrl = updateResult.releaseUrl}
          <button type="button" class="action-btn" on:click={() => open(releaseUrl)}>
            {$t('about.update.viewRelease')}
          </button>
        {/if}
        <button type="button" class="action-btn" disabled={checkingUpdate} on:click={checkNow}>
          <IconRefresh size={14} stroke={1.8} />
          {checkingUpdate ? $t('about.update.checking') : $t('about.update.checkNow')}
        </button>
      </div>
    </div>
  </div>

  <p class="credit">{$t('about.madeBy')} {APP.author}. {$t('about.builtWith')}</p>
</div>

{#if showLicenses}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="overlay" on:click={() => (showLicenses = false)}>
    <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions a11y-no-noninteractive-element-interactions -->
    <div class="modal" role="dialog" aria-modal="true" aria-label={$t('about.licenses')} tabindex="-1" on:click|stopPropagation>
      <header>
        <span class="m-title">{$t('about.licenses')}</span>
        <button type="button" class="m-close" aria-label={$t('about.close')} on:click={() => (showLicenses = false)}>
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

  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-3) 0;
    border-top: var(--hairline) solid var(--border-subtle);
    margin-top: var(--space-3);
  }

  .row-label {
    display: flex;
    flex-direction: column;
    gap: 2px;
    font-size: var(--fz-body);
    color: var(--text-primary);
  }

  .update-block {
    padding-top: var(--space-3);
    border-top: var(--hairline) solid var(--border-subtle);
    margin-top: var(--space-3);
  }

  .update-block .row {
    border-top: none;
    margin-top: 0;
    padding-top: var(--space-3);
  }

  .hint {
    margin: var(--space-2) 0 0;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .update-actions {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
  }

  .action-btn {
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
    white-space: nowrap;
  }

  .action-btn:hover {
    background: var(--surface-hover);
  }

  .action-btn:disabled {
    opacity: 0.6;
    cursor: default;
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
