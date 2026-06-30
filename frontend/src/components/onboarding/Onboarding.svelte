<script lang="ts">
  // the first-run onboarding. it is a deliberately unhurried, full-screen flow:
  // a welcome, a few honest selling points, then theme / accent / density chosen
  // with live previews (every pick applies to the whole app immediately), then
  // adding a mailbox or skipping, and finally a celebratory "all set". a skip
  // affordance is always available. completion is persisted by the parent so this
  // only ever shows once, with a settings action to run it again.
  import { createEventDispatcher } from 'svelte'
  import { fly, fade, scale } from 'svelte/transition'
  import { quintOut, backOut } from 'svelte/easing'
  import {
    IconArrowRight,
    IconArrowLeft,
    IconCheck,
    IconDeviceDesktop,
    IconSun,
    IconMoon,
    IconShieldLock,
    IconBolt,
    IconAdjustmentsHorizontal,
    IconPackageExport,
    IconBrandOpenSource,
    IconSparkles,
    IconBrandGoogle,
    IconBrandApple,
    IconServer,
    IconDots,
    IconPalette,
  } from '@tabler/icons-svelte'
  import Confetti from '../common/Confetti.svelte'
  import AppSkeleton from './AppSkeleton.svelte'
  import ColorPicker from '../common/ColorPicker.svelte'
  import AddMailboxWizard from '../wizard/AddMailboxWizard.svelte'
  import { prefs, setTheme, setDensity, setAccent, setUIScale } from '../../stores/prefs'
  import { ACCENT_PRESETS } from '../../theme/accent'
  import { shortcutLabel } from '../../lib/i18n'
  import logo from '../../assets/images/icons/pelton-logo.png'
  import type { ThemePref, DensityPref, Account } from '../../lib/types'

  const dispatch = createEventDispatcher<{ finish: void; added: Account }>()

  // the ordered steps. "done" is the finale and has no skip/progress chrome.
  type Step = 'welcome' | 'features' | 'theme' | 'accent' | 'density' | 'scale' | 'mailbox' | 'done'
  const order: Step[] = ['welcome', 'features', 'theme', 'accent', 'density', 'scale', 'mailbox', 'done']
  let index = 0
  $: step = order[index]
  // direction drives the slide of the step transition (forward vs back).
  let dir = 1

  let showWizard = false
  let skippedMailbox = false

  function next(): void {
    dir = 1
    index = Math.min(index + 1, order.length - 1)
  }

  function back(): void {
    dir = -1
    index = Math.max(index - 1, 0)
  }

  function finish(): void {
    dispatch('finish')
  }

  const themeChoices: { key: ThemePref; label: string; icon: typeof IconSun }[] = [
    { key: 'system', label: 'System', icon: IconDeviceDesktop },
    { key: 'light', label: 'Light', icon: IconSun },
    { key: 'dark', label: 'Dark', icon: IconMoon },
  ]

  // each density card illustrates its spacing with bars set apart by a real
  // spacing token, so the comparison is honest and stays theme-driven. selecting
  // also applies the density to the whole app live.
  const densityChoices: { key: DensityPref; label: string; gap: string }[] = [
    { key: 'compact', label: 'Compact', gap: 'var(--space-1)' },
    { key: 'medium', label: 'Medium', gap: 'var(--space-2)' },
    { key: 'luxe', label: 'Luxe', gap: 'var(--space-3)' },
  ]

  // interface zoom choices, shown live so the whole onboarding scales as you pick.
  const scaleChoices: { key: string; label: string }[] = [
    { key: '0.9', label: '90%' },
    { key: '1', label: '100%' },
    { key: '1.1', label: '110%' },
    { key: '1.25', label: '125%' },
    { key: '1.5', label: '150%' },
  ]

  const features = [
    { icon: IconShieldLock, title: 'Full privacy', body: 'Your data stays on your machine. Zero tracking, zero telemetry, and complete control over your inbox.' },
    { icon: IconBolt, title: 'Fast search', body: 'Find what you need instantly. The search engine is optimized for speed and handles large local mailboxes with ease.' },
    { icon: IconAdjustmentsHorizontal, title: 'Highly customizable', body: 'Tailor the client to fit your exact workflow and aesthetic preferences.' },
    { icon: IconPackageExport, title: 'Portable configuration', body: 'Export your entire setup, including accounts, preferences and custom layouts, into a single easily transferable file.' },
    { icon: IconBrandOpenSource, title: 'FOSS & cross-platform', body: 'Truly open source and built to run beautifully across different operating systems.' },
  ]

  const addMailboxHint = shortcutLabel('mod+m')

  // the provider that the wizard should open straight into, or null for the full
  // provider grid.
  let wizardProvider: string | null = null

  function openProvider(id: string | null): void {
    wizardProvider = id
    showWizard = true
  }

  // quick provider cards shown directly on the mailbox step.
  const quickProviders = [
    { id: 'gmail', label: 'Gmail', icon: IconBrandGoogle, sub: 'Sign in with Google' },
    { id: 'icloud', label: 'iCloud', icon: IconBrandApple, sub: 'App-specific password' },
    { id: 'custom', label: 'Custom IMAP / SMTP', icon: IconServer, sub: 'Auto-detected servers' },
  ]

  // custom accent: a popover color picker (hue square + slider + hex), opened from
  // the custom swatch. it applies live through the same accent setter.
  let pickerOpen = false
  $: isPresetAccent = ACCENT_PRESETS.some((c) => c.toLowerCase() === $prefs.accent.toLowerCase())

  function onMailboxAdded(event: CustomEvent<Account>): void {
    showWizard = false
    dispatch('added', event.detail)
    dir = 1
    index = order.indexOf('done')
  }

  function skipMailbox(): void {
    skippedMailbox = true
    next()
  }

  // slide params: enter from the side we are moving toward, leave to the other.
  $: enterX = dir * 36
</script>

<div class="screen" role="dialog" aria-modal="true" aria-label="Welcome to Pelton">
  {#if step !== 'done'}
    <button type="button" class="skip" on:click={finish}>Skip</button>
  {/if}

  <div class="stage">
    {#key step}
      <div class="step" in:fly={{ x: enterX, y: 8, duration: 380, easing: quintOut, delay: 90 }} out:fade={{ duration: 120 }}>
        {#if step === 'welcome'}
          <div class="welcome">
            <img class="logo" src={logo} alt="Pelton" in:scale={{ start: 0.7, duration: 600, easing: backOut }} />
            <h1 in:fly={{ y: 12, duration: 500, delay: 160, easing: quintOut }}>Welcome to Pelton</h1>
            <p class="lede" in:fly={{ y: 12, duration: 500, delay: 260, easing: quintOut }}>
              A calm, private, open-source home for your email. Let's make it yours.
            </p>
            <button class="primary big" on:click={next} in:fly={{ y: 12, duration: 500, delay: 380, easing: quintOut }}>
              Get started <IconArrowRight size={18} stroke={1.8} />
            </button>
          </div>
        {:else if step === 'features'}
          <div class="features">
            <h2>Why you'll like it here</h2>
            <ul>
              {#each features as f, i}
                <li in:fly={{ y: 16, duration: 460, delay: 120 + i * 90, easing: quintOut }}>
                  <span class="f-icon"><svelte:component this={f.icon} size={22} stroke={1.6} /></span>
                  <span class="f-text">
                    <span class="f-title">{f.title}</span>
                    <span class="f-body">{f.body}</span>
                  </span>
                </li>
              {/each}
            </ul>
            <div class="nav">
              <button class="ghost" on:click={back}><IconArrowLeft size={16} stroke={1.8} /> Back</button>
              <button class="primary" on:click={next}>Continue <IconArrowRight size={16} stroke={1.8} /></button>
            </div>
          </div>
        {:else if step === 'theme'}
          <div class="choose">
            <h2>Pick a theme</h2>
            <div class="cards three">
              {#each themeChoices as c}
                <button class="card" class:active={$prefs.theme === c.key} on:click={() => setTheme(c.key)}>
                  <span class="card-icon"><svelte:component this={c.icon} size={26} stroke={1.5} /></span>
                  <span class="card-label">{c.label}</span>
                  {#if $prefs.theme === c.key}<span class="tick"><IconCheck size={14} stroke={2.4} /></span>{/if}
                </button>
              {/each}
            </div>
            <div class="nav">
              <button class="ghost" on:click={back}><IconArrowLeft size={16} stroke={1.8} /> Back</button>
              <button class="primary" on:click={next}>Continue <IconArrowRight size={16} stroke={1.8} /></button>
            </div>
          </div>
        {:else if step === 'accent'}
          <div class="choose">
            <h2>Choose an accent</h2>
            <AppSkeleton />
            <div class="swatches">
              {#each ACCENT_PRESETS as color}
                <button
                  class="swatch"
                  class:active={$prefs.accent.toLowerCase() === color.toLowerCase()}
                  style={`--sw:${color}`}
                  aria-label={`Accent ${color}`}
                  on:click={() => setAccent(color)}
                >
                  {#if $prefs.accent.toLowerCase() === color.toLowerCase()}
                    <IconCheck size={16} stroke={2.6} />
                  {/if}
                </button>
              {/each}
              <button
                type="button"
                class="custom-swatch"
                class:active={!isPresetAccent}
                title="Custom color"
                aria-label="Custom color"
                on:click={() => (pickerOpen = !pickerOpen)}
              >
                <IconPalette size={16} stroke={1.8} />
              </button>
            </div>
            {#if pickerOpen}
              <div class="picker-wrap" transition:fade={{ duration: 120 }}>
                <ColorPicker value={$prefs.accent} on:change={(e) => setAccent(e.detail)} />
              </div>
            {/if}
            <div class="nav">
              <button class="ghost" on:click={back}><IconArrowLeft size={16} stroke={1.8} /> Back</button>
              <button class="primary" on:click={next}>Continue <IconArrowRight size={16} stroke={1.8} /></button>
            </div>
          </div>
        {:else if step === 'density'}
          <div class="choose">
            <h2>Set the density</h2>
            <AppSkeleton />
            <div class="cards three dense">
              {#each densityChoices as c}
                <button class="card" class:active={$prefs.density === c.key} on:click={() => setDensity(c.key)}>
                  <span class="bars" style={`gap:${c.gap}`}>
                    <span class="bar"></span><span class="bar short"></span><span class="bar"></span>
                  </span>
                  <span class="card-label">{c.label}</span>
                  {#if $prefs.density === c.key}<span class="tick"><IconCheck size={14} stroke={2.4} /></span>{/if}
                </button>
              {/each}
            </div>
            <div class="nav">
              <button class="ghost" on:click={back}><IconArrowLeft size={16} stroke={1.8} /> Back</button>
              <button class="primary" on:click={next}>Continue <IconArrowRight size={16} stroke={1.8} /></button>
            </div>
          </div>
        {:else if step === 'scale'}
          <div class="choose">
            <h2>Set the interface size</h2>
            <p class="sub">Make everything bigger or smaller. You can change this any time in Settings.</p>
            <AppSkeleton />
            <div class="cards five">
              {#each scaleChoices as c}
                <button class="card" class:active={$prefs.uiScale === c.key} on:click={() => setUIScale(c.key)}>
                  <span class="card-label">{c.label}</span>
                  {#if $prefs.uiScale === c.key}<span class="tick"><IconCheck size={14} stroke={2.4} /></span>{/if}
                </button>
              {/each}
            </div>
            <div class="nav">
              <button class="ghost" on:click={back}><IconArrowLeft size={16} stroke={1.8} /> Back</button>
              <button class="primary" on:click={next}>Continue <IconArrowRight size={16} stroke={1.8} /></button>
            </div>
          </div>
        {:else if step === 'mailbox'}
          <div class="choose">
            <h2>Add your first mailbox</h2>
            <p class="sub">Connect an account now, or do it later.</p>
            <div class="providers">
              {#each quickProviders as p}
                <button class="provider" on:click={() => openProvider(p.id)}>
                  <span class="p-icon"><svelte:component this={p.icon} size={24} stroke={1.5} /></span>
                  <span class="p-text">
                    <span class="p-title">{p.label}</span>
                    <span class="p-sub">{p.sub}</span>
                  </span>
                  <IconArrowRight size={16} stroke={1.8} />
                </button>
              {/each}
              <button class="provider more" on:click={() => openProvider(null)}>
                <span class="p-icon"><IconDots size={24} stroke={1.5} /></span>
                <span class="p-text">
                  <span class="p-title">More providers</span>
                  <span class="p-sub">Outlook, Yahoo, Fastmail and others</span>
                </span>
                <IconArrowRight size={16} stroke={1.8} />
              </button>
            </div>
            <div class="nav">
              <button class="ghost" on:click={back}><IconArrowLeft size={16} stroke={1.8} /> Back</button>
              <button class="primary" on:click={skipMailbox}>Skip for now</button>
            </div>
          </div>
        {:else if step === 'done'}
          <div class="welcome done">
            <span class="done-mark" in:scale={{ start: 0.5, duration: 600, easing: backOut }}>
              <IconSparkles size={40} stroke={1.5} />
            </span>
            <h1 in:fly={{ y: 12, duration: 500, delay: 160, easing: quintOut }}>All set!</h1>
            <p class="lede" in:fly={{ y: 12, duration: 500, delay: 260, easing: quintOut }}>
              {#if skippedMailbox}
                You can add mailboxes any time with <kbd>{addMailboxHint}</kbd> or from the Mailbox menu.
              {:else}
                Pelton is ready. Enjoy a calmer inbox.
              {/if}
            </p>
            <button class="primary big" on:click={finish} in:fly={{ y: 12, duration: 500, delay: 380, easing: quintOut }}>
              Start using Pelton
            </button>
          </div>
        {/if}
      </div>
    {/key}
  </div>

  {#if step !== 'done'}
    <div class="dots" aria-hidden="true">
      {#each order.slice(0, -1) as _, i}
        <span class="dot" class:on={i === index}></span>
      {/each}
    </div>
  {/if}
</div>

{#if step === 'done'}
  <Confetti />
{/if}

{#if showWizard}
  <AddMailboxWizard
    initialProviderId={wizardProvider}
    on:close={() => (showWizard = false)}
    on:added={onMailboxAdded}
  />
{/if}

<style>
  .screen {
    position: fixed;
    inset: 0;
    z-index: 120;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: var(--surface-base);
    overflow: hidden;
  }

  .skip {
    position: absolute;
    top: var(--space-5);
    right: var(--space-6);
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    font-size: var(--fz-label);
    cursor: pointer;
    padding: var(--space-2) var(--space-3);
    border-radius: var(--radius-control);
  }

  .skip:hover {
    background: var(--surface-hover);
    color: var(--text-secondary);
  }

  .stage {
    width: 100%;
    max-width: 560px;
    padding: 0 var(--space-6);
    display: grid;
  }

  /* every step occupies the same grid cell so transitions cross-fade in place. */
  .step {
    grid-area: 1 / 1;
  }

  h1 {
    margin: var(--space-4) 0 0;
    font-size: 2rem;
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
    letter-spacing: -0.02em;
  }

  h2 {
    margin: 0 0 var(--space-2);
    font-size: var(--fz-title);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
    text-align: center;
  }

  .sub,
  .lede {
    color: var(--text-secondary);
    line-height: 1.55;
  }

  .sub {
    margin: 0 0 var(--space-5);
    font-size: var(--fz-label);
    text-align: center;
  }

  /* welcome + done are centered hero layouts. */
  .welcome {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .logo {
    width: 96px;
    height: 96px;
    object-fit: contain;
    filter: drop-shadow(0 8px 24px rgba(0, 0, 0, 0.18));
  }

  .lede {
    margin: var(--space-3) 0 var(--space-6);
    font-size: var(--fz-body);
    max-width: 42ch;
  }

  .done-mark {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 84px;
    height: 84px;
    border-radius: 999px;
    background: var(--surface-sunken);
    color: var(--accent);
    border: var(--hairline) solid var(--border-subtle);
  }

  kbd {
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    padding: 1px var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
  }

  /* features list. */
  .features ul {
    list-style: none;
    margin: var(--space-5) 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .features li {
    display: flex;
    gap: var(--space-4);
    align-items: flex-start;
  }

  .f-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 42px;
    flex-shrink: 0;
    border-radius: var(--radius-card);
    background: var(--surface-sunken);
    color: var(--accent);
    border: var(--hairline) solid var(--border-subtle);
  }

  .f-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .f-title {
    font-size: var(--fz-body);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .f-body {
    font-size: var(--fz-label);
    color: var(--text-secondary);
    line-height: 1.5;
  }

  /* choice cards (theme / density). */
  .cards.three {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-3);
  }

  /* the interface-size step: five compact choices in a row. */
  .cards.five {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: var(--space-2);
    margin-bottom: var(--space-6);
  }

  .cards.five .card {
    padding: var(--space-4) var(--space-2);
  }

  .card {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-5) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-raised);
    color: var(--text-primary);
    cursor: pointer;
    transition: border-color 0.15s ease, transform 0.15s ease;
  }

  .card:hover {
    transform: translateY(-2px);
  }

  .card.active {
    border-color: var(--accent);
    box-shadow: 0 0 0 1px var(--accent) inset;
  }

  .card-icon {
    color: var(--text-secondary);
  }

  .card.active .card-icon {
    color: var(--accent);
  }

  .card-label {
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
  }

  .tick {
    position: absolute;
    top: var(--space-2);
    right: var(--space-2);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    border-radius: 999px;
    background: var(--accent);
    color: var(--accent-fg);
  }

  /* density illustration bars. */
  .bars {
    display: flex;
    flex-direction: column;
    width: 100%;
    padding: 0 var(--space-3);
  }

  .bar {
    height: 6px;
    border-radius: 999px;
    background: var(--border-strong, var(--text-tertiary));
    opacity: 0.5;
  }

  .bar.short {
    width: 60%;
  }

  /* accent swatches + live preview. */
  .swatches {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--space-3);
    margin-bottom: var(--space-5);
  }

  .swatch {
    width: 40px;
    height: 40px;
    border-radius: 999px;
    border: 2px solid transparent;
    background: var(--sw);
    color: #fff;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.15s ease;
  }

  .swatch:hover {
    transform: scale(1.08);
  }

  .swatch.active {
    border-color: var(--text-primary);
  }

  /* the native color picker, shown as a round swatch tinted with the current
     accent and a palette glyph. the real input is an invisible overlay. */
  .custom-swatch {
    width: 40px;
    height: 40px;
    border-radius: 999px;
    border: var(--hairline) dashed var(--border-strong, var(--border-default));
    overflow: hidden;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    position: relative;
    background: var(--accent);
    color: var(--accent-fg);
  }

  .custom-swatch.active {
    box-shadow: 0 0 0 2px var(--accent);
    border-style: solid;
  }

  .picker-wrap {
    display: flex;
    justify-content: center;
    margin-bottom: var(--space-6);
  }

  /* mailbox provider cards. */
  .providers {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin-bottom: var(--space-6);
  }

  .provider {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    width: 100%;
    padding: var(--space-4) var(--space-5);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-raised);
    color: var(--text-primary);
    cursor: pointer;
    transition: border-color 0.15s ease, transform 0.15s ease;
  }

  .provider:hover {
    transform: translateY(-2px);
    border-color: var(--accent);
  }

  .provider.more {
    background: transparent;
  }

  .p-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 44px;
    height: 44px;
    flex-shrink: 0;
    border-radius: var(--radius-card);
    background: var(--surface-sunken);
    color: var(--accent);
  }

  .p-text {
    display: flex;
    flex-direction: column;
    gap: 2px;
    flex: 1;
    text-align: left;
  }

  .p-title {
    font-size: var(--fz-body);
    font-weight: var(--fw-semibold);
  }

  .p-sub {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }

  /* choose-step rhythm: give the heading, preview and controls room to breathe
     so the steps do not feel cramped. */
  .choose {
    padding: var(--space-4) 0;
  }

  .choose h2 {
    margin-bottom: var(--space-5);
  }

  .choose :global(.mock) {
    margin-bottom: var(--space-5);
  }

  .cards.three {
    margin-bottom: var(--space-6);
  }

  /* shared nav + buttons. */
  .nav {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--space-3);
    margin-top: var(--space-6);
  }

  .primary,
  .ghost {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-4) var(--space-6);
    border-radius: var(--radius-control);
    font-size: var(--fz-body);
    font-weight: var(--fw-medium);
    cursor: pointer;
    border: var(--hairline) solid var(--border-default);
  }

  .primary {
    background: var(--accent);
    color: var(--accent-fg);
    border-color: transparent;
  }

  .primary:hover {
    filter: brightness(1.05);
  }

  .primary.big {
    padding: var(--space-5) var(--space-6);
    font-size: var(--fz-body);
  }

  .ghost {
    background: transparent;
    color: var(--text-secondary);
  }

  .ghost:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  /* progress dots. */
  .dots {
    position: absolute;
    bottom: var(--space-6);
    display: flex;
    gap: var(--space-2);
  }

  .dot {
    width: 7px;
    height: 7px;
    border-radius: 999px;
    background: var(--border-strong, var(--border-default));
    opacity: 0.5;
    transition: width 0.2s ease, opacity 0.2s ease, background 0.2s ease;
  }

  .dot.on {
    width: 20px;
    opacity: 1;
    background: var(--accent);
  }
</style>
