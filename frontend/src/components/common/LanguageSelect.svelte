<script lang="ts">
  // a small language picker shared by Settings and onboarding. each language is
  // shown in its own spelling (Deutsch, Français, ...) rather than translated
  // into the currently active one, so it stays recognizable regardless of what
  // is selected. the OS-detected language gets a "Recommended" badge, but it is
  // never auto-applied: the caller decides the initial value. custom language
  // files from the locales folder are listed after the built-ins; their value
  // is "user:<id>".
  import { onMount } from 'svelte'
  import { IconCheck } from '@tabler/icons-svelte'
  import { locales, localeNames, detectOSLocale, userLocalePrefix, t } from '../../lib/i18n'
  import { listUserLocales } from '../../lib/api'
  import type { UserLocale } from '../../lib/types'

  /** the persisted language setting: a built-in code or "user:<id>". */
  export let value: string
  export let onSelect: (value: string) => void

  const recommended = detectOSLocale()

  let userLocales: UserLocale[] = []
  onMount(async () => {
    try {
      userLocales = await listUserLocales()
    } catch {
      // no storage yet (early onboarding); the built-ins are always there.
    }
  })
</script>

<div class="lang-grid" role="listbox" aria-label={$t('settings.language')}>
  {#each locales as l (l)}
    <button type="button" class="lang-card" class:active={value === l} on:click={() => onSelect(l)} role="option" aria-selected={value === l}>
      <span class="lang-row">
        <span class="lang-name">{localeNames[l]}</span>
        {#if value === l}
          <IconCheck size={14} stroke={2} class="lang-check" />
        {/if}
      </span>
      {#if l === recommended}
        <span class="lang-badge">{$t('settings.recommended')}</span>
      {/if}
    </button>
  {/each}
</div>

{#if userLocales.length}
  <p class="user-heading">{$t('language.custom')}</p>
  <div class="lang-grid" role="listbox" aria-label={$t('language.custom')}>
    {#each userLocales as l (l.id)}
      {@const v = userLocalePrefix + l.id}
      <button type="button" class="lang-card" class:active={value === v} on:click={() => onSelect(v)} role="option" aria-selected={value === v}>
        <span class="lang-row">
          <span class="lang-name">{l.name}</span>
          {#if value === v}
            <IconCheck size={14} stroke={2} class="lang-check" />
          {/if}
        </span>
        {#if l.author}
          <span class="lang-badge">{l.author}</span>
        {/if}
      </button>
    {/each}
  </div>
{/if}

<style>
  .lang-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: var(--space-3);
  }

  .lang-card {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-1);
    padding: var(--space-3) var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: pointer;
    text-align: left;
  }

  .lang-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    width: 100%;
  }

  .lang-card:hover {
    background: var(--surface-hover);
  }

  .lang-card.active {
    border-color: var(--accent);
    box-shadow: 0 0 0 1px var(--accent);
  }

  .lang-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .lang-badge {
    flex-shrink: 0;
    padding: 1px 6px;
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-tertiary);
    font-size: var(--fz-meta);
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .lang-card.active :global(.lang-check) {
    color: var(--accent);
    flex-shrink: 0;
  }

  .user-heading {
    margin: var(--space-4) 0 var(--space-3);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }
</style>
