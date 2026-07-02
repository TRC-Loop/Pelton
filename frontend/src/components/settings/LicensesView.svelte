<script lang="ts">
  // the open-source licenses view, code-split and loaded only when the user opens
  // it from the about section. it fetches the embedded manifest from the backend
  // (so the text never ships in the frontend bundle), lists each dependency with
  // its license, expandable to the full text, and offers Pelton's own GPL-3.0
  // text on demand.
  import { onMount } from 'svelte'
  import { IconChevronRight, IconLicense } from '@tabler/icons-svelte'
  import { licenses, programLicense, type LicenseEntry } from '../../lib/api'
  import { errorMessage } from '../../stores/toast'
  import { t } from '../../lib/i18n'

  let entries: LicenseEntry[] = []
  let loading = true
  let error = ''
  let openName: string | null = null

  // the program's own license text, loaded lazily on first request.
  let gplText = ''
  let gplOpen = false

  onMount(async () => {
    try {
      entries = await licenses()
    } catch (err) {
      error = errorMessage(err)
    } finally {
      loading = false
    }
  })

  function toggle(name: string): void {
    openName = openName === name ? null : name
  }

  async function toggleGpl(): Promise<void> {
    gplOpen = !gplOpen
    if (gplOpen && !gplText) {
      try {
        gplText = await programLicense()
      } catch (err) {
        gplText = errorMessage(err)
      }
    }
  }

  $: goEntries = entries.filter((e) => e.group === 'go')
  $: npmEntries = entries.filter((e) => e.group === 'npm')
  $: sections = [
    { label: $t('settingsPanel.goModules'), list: goEntries },
    { label: $t('settingsPanel.frontendPackages'), list: npmEntries },
  ]
</script>

<div class="licenses">
  <button type="button" class="program" on:click={toggleGpl}>
    <IconLicense size={15} stroke={1.6} />
    <span>Pelton {$t('settingsPanel.isLicensedUnder')} GPL-3.0</span>
    <IconChevronRight size={14} stroke={1.8} class={gplOpen ? 'chev open' : 'chev'} />
  </button>
  {#if gplOpen}
    <pre class="text selectable">{gplText || $t('settingsPanel.loading')}</pre>
  {/if}

  {#if loading}
    <p class="muted">{$t('settingsPanel.loadingLicenses')}</p>
  {:else if error}
    <p class="err">{error}</p>
  {:else if entries.length === 0}
    <p class="muted">{$t('settingsPanel.noManifest')} <code>make licenses</code> {$t('settingsPanel.andRebuild')}</p>
  {:else}
    {#each sections as section}
      {#if section.list.length > 0}
        <h4>{section.label}</h4>
        <ul>
          {#each section.list as e (e.group + e.name)}
            <li>
              <button type="button" class="row" on:click={() => toggle(e.name)}>
                <IconChevronRight size={13} stroke={1.8} class={openName === e.name ? 'chev open' : 'chev'} />
                <span class="name">{e.name}</span>
                <span class="badge">{e.license}</span>
              </button>
              {#if openName === e.name}
                <pre class="text selectable">{e.text}</pre>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
    {/each}
  {/if}
</div>

<style>
  .licenses {
    margin-top: var(--space-3);
  }

  .program {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .program span {
    flex: 1;
    text-align: left;
  }

  h4 {
    margin: var(--space-4) 0 var(--space-2);
    font-size: var(--fz-label);
    font-weight: var(--fw-semibold);
    color: var(--text-secondary);
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    width: 100%;
    padding: var(--space-2) var(--space-1);
    border: none;
    background: transparent;
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: pointer;
    text-align: left;
    border-radius: var(--radius-control);
  }

  .row:hover {
    background: var(--surface-hover);
  }

  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
  }

  .badge {
    flex-shrink: 0;
    padding: 1px var(--space-2);
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-control);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  :global(.licenses .chev) {
    transition: transform 0.15s ease;
    flex-shrink: 0;
    color: var(--text-tertiary);
  }

  :global(.licenses .chev.open) {
    transform: rotate(90deg);
  }

  .text {
    margin: var(--space-2) 0 var(--space-3);
    padding: var(--space-3);
    max-height: 280px;
    overflow: auto;
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-word;
    color: var(--text-secondary);
  }

  .muted {
    color: var(--text-tertiary);
    font-size: var(--fz-label);
  }

  .err {
    color: var(--danger);
    font-size: var(--fz-label);
  }

  code {
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
  }
</style>
