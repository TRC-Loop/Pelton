<script lang="ts">
  // the settings screen. it fills the window and uses a two-column layout: a
  // category nav on the left and the selected category's controls on the right,
  // so the (many) preferences stay übersichtlich instead of one long scroll.
  // every control has a short hint/tooltip. no preference was removed in the
  // reorganization — they are only grouped.
  import {
    IconX,
    IconPalette,
    IconList,
    IconLayoutSidebar,
    IconUserCircle,
    IconSignature,
    IconSend2,
    IconShieldLock,
    IconBell,
    IconLayoutColumns,
    IconEye,
    IconKeyboard,
    IconRocket,
    IconInfoCircle,
  } from '@tabler/icons-svelte'
  import { createEventDispatcher } from 'svelte'
  import SegmentedSetting from './SegmentedSetting.svelte'
  import AccentPicker from './AccentPicker.svelte'
  import TechToggles from './TechToggles.svelte'
  import ToastPositionPicker from './ToastPositionPicker.svelte'
  import ShortcutSettings from './ShortcutSettings.svelte'
  import SignaturesSection from './SignaturesSection.svelte'
  import AboutSection from './AboutSection.svelte'
  import ToggleSwitch from '../common/ToggleSwitch.svelte'
  import { pfpDataUri, type PfpStyle } from '../../lib/pfp'
  import { initials } from '../../lib/format'
  import {
    prefs,
    setTheme,
    setDensity,
    setUIScale,
    setMessageFontSize,
    setToastPosition,
    setPaneLocked,
    setSendDelay,
    setFlagHighlight,
    setShortcutHints,
    setShowAccountEmail,
    setAlwaysLoadImages,
    setAvatarSource,
    setAvatarStyle,
    setMultiSelectEnabled,
    setShowSelectedCount,
    setSidebarIndentGuides,
    setShowFlaggedCount,
    setRowTemplate,
    setRowShowAvatar,
    setRowShowSnippet,
    setPreviewLines,
  } from '../../stores/prefs'
  import { t } from '../../lib/i18n'
  import type { ThemePref, DensityPref } from '../../lib/types'

  const dispatch = createEventDispatcher<{ close: void; rerunOnboarding: void }>()

  // left-nav categories. each maps to a block rendered on the right.
  const categories = [
    { key: 'appearance', label: 'Appearance', icon: IconPalette },
    { key: 'list', label: 'Message list', icon: IconList },
    { key: 'sidebar', label: 'Sidebar', icon: IconLayoutSidebar },
    { key: 'avatars', label: 'Avatars', icon: IconUserCircle },
    { key: 'signatures', label: 'Signatures', icon: IconSignature },
    { key: 'sending', label: 'Sending', icon: IconSend2 },
    { key: 'privacy', label: 'Privacy', icon: IconShieldLock },
    { key: 'notifications', label: 'Notifications', icon: IconBell },
    { key: 'panes', label: 'Panes', icon: IconLayoutColumns },
    { key: 'display', label: 'Message display', icon: IconEye },
    { key: 'shortcuts', label: 'Shortcuts', icon: IconKeyboard },
    { key: 'start', label: 'Getting started', icon: IconRocket },
    { key: 'about', label: 'About', icon: IconInfoCircle },
  ]
  let active = 'appearance'

  const themeOptions = [
    { key: 'system', label: 'System' },
    { key: 'light', label: 'Light' },
    { key: 'dark', label: 'Dark' },
  ]

  const densityOptions = [
    { key: 'compact', label: 'Compact' },
    { key: 'medium', label: 'Medium' },
    { key: 'luxe', label: 'Luxe' },
  ]

  // interface zoom. values are string multipliers applied as css zoom.
  const scaleOptions = [
    { key: '0.9', label: '90%' },
    { key: '1', label: '100%' },
    { key: '1.1', label: '110%' },
    { key: '1.25', label: '125%' },
    { key: '1.5', label: '150%' },
  ]

  // base font size (px) for rendered email content.
  const messageFontOptions = [
    { key: '12', label: 'Small' },
    { key: '14', label: 'Default' },
    { key: '16', label: 'Large' },
    { key: '18', label: 'Larger' },
    { key: '20', label: 'Largest' },
  ]

  const sendDelayOptions = [
    { key: '0', label: 'Off' },
    { key: '5', label: '5s' },
    { key: '10', label: '10s' },
    { key: '30', label: '30s' },
    { key: '60', label: '60s' },
  ]

  function onSendDelay(event: CustomEvent<string>): void {
    setSendDelay(Number(event.detail))
  }

  // enabling the global remote-image override is guarded by a tracking warning.
  let confirmImages = false
  function onImagesToggle(checked: boolean): void {
    if (checked && !$prefs.alwaysLoadImages) {
      confirmImages = true
    } else {
      setAlwaysLoadImages(false)
    }
  }
  function confirmEnableImages(): void {
    setAlwaysLoadImages(true)
    confirmImages = false
  }

  // sender-photo fallback chain. "Generated" never touches the network.
  const avatarSourceOptions = [
    { key: 'bimi_gravatar', label: 'Logo → Gravatar' },
    { key: 'gravatar_bimi', label: 'Gravatar → Logo' },
    { key: 'pfp', label: 'Generated' },
  ]

  // generated placeholder styles, previewed with a sample sender so the look is
  // obvious before choosing.
  const sampleEmail = 'potato@pelton.email'
  const sampleInitials = initials('', sampleEmail)
  const avatarStyleOptions: { key: PfpStyle; label: string }[] = [
    { key: 'initials', label: 'Classic' },
    { key: 'mono', label: 'Monochrome' },
    { key: 'pixel', label: 'Pixel' },
    { key: 'geometric', label: 'Geometric' },
  ]
  function stylePreview(style: PfpStyle): string {
    return pfpDataUri(style, sampleEmail, sampleInitials)
  }

  const flagOptions = [
    { key: 'flag', label: 'Flag icon' },
    { key: 'left', label: 'Left bar' },
    { key: 'both', label: 'Bar + icon' },
    { key: 'off', label: 'Off' },
  ]

  const rowTemplateOptions = [
    { key: 'relaxed', label: 'Relaxed' },
    { key: 'comfortable', label: 'Comfortable' },
    { key: 'compact', label: 'Compact' },
    { key: 'single', label: 'Single line' },
  ]

  const previewLineOptions = [
    { key: '1', label: '1 line' },
    { key: '2', label: '2 lines' },
    { key: '3', label: '3 lines' },
  ]

  $: snippetCapable = $prefs.rowTemplate === 'relaxed' || $prefs.rowTemplate === 'comfortable'

  function onTheme(event: CustomEvent<string>): void {
    setTheme(event.detail as ThemePref)
  }

  function onDensity(event: CustomEvent<string>): void {
    setDensity(event.detail as DensityPref)
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      dispatch('close')
    }
  }
</script>

<svelte:window on:keydown={onKeydown} />

<div class="screen" role="dialog" aria-modal="true" aria-label="Settings">
  <header class="head">
    <h2>Settings</h2>
    <button type="button" class="close" aria-label="Close settings" on:click={() => dispatch('close')}>
      <IconX size={20} stroke={1.8} />
    </button>
  </header>

  <div class="body">
    <nav class="nav" aria-label="Settings categories">
      {#each categories as cat (cat.key)}
        <button
          type="button"
          class="nav-item"
          class:active={active === cat.key}
          aria-current={active === cat.key}
          on:click={() => (active = cat.key)}
        >
          <span class="nav-icon"><svelte:component this={cat.icon} size={17} stroke={1.6} /></span>
          <span>{cat.label}</span>
        </button>
      {/each}
    </nav>

    <div class="content">
      {#if active === 'appearance'}
        <section>
          <h3>Appearance</h3>
          <SegmentedSetting label="Theme" value={$prefs.theme} options={themeOptions} on:change={onTheme} />
          <AccentPicker />
          <SegmentedSetting label="Density" value={$prefs.density} options={densityOptions} on:change={onDensity} />
          <SegmentedSetting
            label="Interface scale"
            value={$prefs.uiScale}
            options={scaleOptions}
            on:change={(e) => setUIScale(e.detail)}
          />
          <p class="hint">Zooms the whole interface. Make everything a bit bigger or smaller.</p>
        </section>
      {:else if active === 'list'}
        <section>
          <h3>Message list</h3>
          <SegmentedSetting
            label="Row layout"
            value={$prefs.rowTemplate}
            options={rowTemplateOptions}
            on:change={(e) => setRowTemplate(e.detail)}
          />
          <div class="toggle" class:disabled={$prefs.rowTemplate === 'single'} title="The avatar is hidden on the single-line layout.">
            <span class="row-label">Show sender avatar</span>
            <ToggleSwitch
              checked={$prefs.rowShowAvatar}
              disabled={$prefs.rowTemplate === 'single'}
              label="Show sender avatar"
              on:change={(e) => setRowShowAvatar(e.detail)}
            />
          </div>
          <div class="toggle" class:disabled={!snippetCapable} title="Previews show on the relaxed and comfortable layouts.">
            <span class="row-label">Show message preview</span>
            <ToggleSwitch
              checked={$prefs.rowShowSnippet}
              disabled={!snippetCapable}
              label="Show message preview"
              on:change={(e) => setRowShowSnippet(e.detail)}
            />
          </div>
          {#if snippetCapable && $prefs.rowShowSnippet}
            <SegmentedSetting
              label="Preview lines"
              value={String($prefs.previewLines)}
              options={previewLineOptions}
              on:change={(e) => setPreviewLines(Number(e.detail))}
            />
          {/if}
          <SegmentedSetting
            label="Flagged highlight"
            value={$prefs.flagHighlight}
            options={flagOptions}
            on:change={(e) => setFlagHighlight(e.detail)}
          />
          <p class="hint">“Bar + icon” shows a left edge bar together with the flag icon.</p>
          <div class="toggle">
            <span class="row-label">Show mailbox email instead of name</span>
            <ToggleSwitch
              checked={$prefs.showAccountEmail}
              label="Show mailbox email instead of name"
              on:change={(e) => setShowAccountEmail(e.detail)}
            />
          </div>
          <div class="toggle" title="Cmd/Ctrl-click or Shift-click rows to select several at once.">
            <span class="row-label">Select multiple messages at once</span>
            <ToggleSwitch
              checked={$prefs.multiSelectEnabled}
              label="Select multiple messages at once"
              on:change={(e) => setMultiSelectEnabled(e.detail)}
            />
          </div>
          <div class="toggle" class:disabled={!$prefs.multiSelectEnabled}>
            <span class="row-label">Show “N selected” count</span>
            <ToggleSwitch
              checked={$prefs.showSelectedCount}
              disabled={!$prefs.multiSelectEnabled}
              label="Show selected count"
              on:change={(e) => setShowSelectedCount(e.detail)}
            />
          </div>
        </section>
      {:else if active === 'sidebar'}
        <section>
          <h3>Sidebar</h3>
          <div class="toggle" title="Draw vertical guide lines connecting nested folders.">
            <span class="row-label">Show folder indent guides</span>
            <ToggleSwitch
              checked={$prefs.sidebarIndentGuides}
              label="Show folder indent guides"
              on:change={(e) => setSidebarIndentGuides(e.detail)}
            />
          </div>
          <div class="toggle" title="Hide the count and bold styling on the Flagged view (the entry stays).">
            <span class="row-label">Show flagged count</span>
            <ToggleSwitch
              checked={$prefs.showFlaggedCount}
              label="Show flagged count"
              on:change={(e) => setShowFlaggedCount(e.detail)}
            />
          </div>
        </section>
      {:else if active === 'avatars'}
        <section>
          <h3>Avatars</h3>
          <SegmentedSetting
            label="Sender photos"
            value={$prefs.avatarSource}
            options={avatarSourceOptions}
            on:change={(e) => setAvatarSource(e.detail)}
          />
          <p class="hint">
            “Logo” is the sender domain's verified BIMI logo. Logo and Gravatar fetch images over the
            network when you open a message; “Generated” stays fully offline. When no photo is found,
            the generated style below is used.
          </p>

          <div class="field">
            <span class="row-label">Generated style</span>
            <div class="style-grid">
              {#each avatarStyleOptions as opt (opt.key)}
                <button
                  type="button"
                  class="style-card"
                  class:active={$prefs.avatarStyle === opt.key}
                  aria-pressed={$prefs.avatarStyle === opt.key}
                  on:click={() => setAvatarStyle(opt.key)}
                >
                  <img class="style-img" src={stylePreview(opt.key)} alt="" aria-hidden="true" />
                  <span class="style-label">{opt.label}</span>
                </button>
              {/each}
            </div>
            <p class="hint">Preview sender: {sampleEmail}</p>
          </div>
        </section>
      {:else if active === 'signatures'}
        <section>
          <SignaturesSection />
        </section>
      {:else if active === 'sending'}
        <section>
          <h3>Sending</h3>
          <SegmentedSetting
            label="Undo send window"
            value={String($prefs.sendDelaySeconds)}
            options={sendDelayOptions}
            on:change={onSendDelay}
          />
          <p class="hint">Hold outgoing mail briefly so you can undo. Press ⌘Z or use the Undo button.</p>
        </section>
      {:else if active === 'privacy'}
        <section>
          <h3>Privacy</h3>
          <div class="toggle" title="Disables remote-image blocking for every message.">
            <span class="row-label">Always load remote images</span>
            <ToggleSwitch
              checked={$prefs.alwaysLoadImages}
              label="Always load remote images"
              on:change={(e) => onImagesToggle(e.detail)}
            />
          </div>
          {#if confirmImages}
            <div class="warn">
              <p>
                Remote images can track when and where you open a message. Loading them for every
                message removes that protection. You can still trust individual senders instead.
              </p>
              <div class="warn-actions">
                <button type="button" class="ghost-btn" on:click={() => (confirmImages = false)}>Cancel</button>
                <button type="button" class="danger-btn" on:click={confirmEnableImages}>Enable anyway</button>
              </div>
            </div>
          {/if}
        </section>
      {:else if active === 'notifications'}
        <section>
          <h3>Notifications</h3>
          <div class="row">
            <span class="row-label">{t('settings.toastPosition')}</span>
            <ToastPositionPicker
              value={$prefs.toastPosition}
              on:change={(e) => setToastPosition(e.detail)}
            />
          </div>
        </section>
      {:else if active === 'panes'}
        <section>
          <h3>{t('settings.panes')}</h3>
          <div class="toggle" title="Prevents dragging the column dividers.">
            <span class="row-label">{t('settings.lockPanes')}</span>
            <ToggleSwitch
              checked={$prefs.paneLocked}
              label={t('settings.lockPanes')}
              on:change={(e) => setPaneLocked(e.detail)}
            />
          </div>
        </section>
      {:else if active === 'display'}
        <section>
          <h3>Message display</h3>
          <SegmentedSetting
            label="Email text size"
            value={String($prefs.messageFontSize)}
            options={messageFontOptions}
            on:change={(e) => setMessageFontSize(Number(e.detail))}
          />
          <p class="hint">Sets the base font size of the message content you read.</p>
          <TechToggles />
        </section>
      {:else if active === 'shortcuts'}
        <section>
          <h3>{t('settings.shortcuts')}</h3>
          <div class="toggle">
            <span class="row-label">Show keyboard shortcut hints in the app</span>
            <ToggleSwitch
              checked={$prefs.showShortcutHints}
              label="Show keyboard shortcut hints in the app"
              on:change={(e) => setShortcutHints(e.detail)}
            />
          </div>
          <ShortcutSettings />
        </section>
      {:else if active === 'start'}
        <section>
          <h3>Getting started</h3>
          <div class="row">
            <span class="row-label">Re-run the welcome tour</span>
            <button type="button" class="action-btn" on:click={() => dispatch('rerunOnboarding')}>
              Re-run onboarding
            </button>
          </div>
        </section>
      {:else if active === 'about'}
        <section>
          <h3>About</h3>
          <AboutSection />
        </section>
      {/if}
    </div>
  </div>
</div>

<style>
  .screen {
    position: fixed;
    inset: 0;
    z-index: 100;
    display: flex;
    flex-direction: column;
    background: var(--surface-base);
  }

  .head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-4) var(--space-6);
    border-bottom: var(--hairline) solid var(--border-default);
    flex-shrink: 0;
  }

  .head h2 {
    margin: 0;
    font-size: var(--fz-title);
    font-weight: var(--fw-semibold);
  }

  .close {
    display: inline-flex;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    padding: var(--space-2);
    border-radius: var(--radius-control);
  }

  .close:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  /* two columns: a fixed nav rail and a scrolling content pane. */
  .body {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: 220px 1fr;
  }

  .nav {
    border-right: var(--hairline) solid var(--border-subtle);
    padding: var(--space-3);
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    padding: var(--space-2) var(--space-3);
    border: none;
    background: transparent;
    border-radius: var(--radius-control);
    color: var(--text-secondary);
    cursor: pointer;
    text-align: left;
    font-size: var(--fz-list);
  }

  .nav-item:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .nav-item.active {
    background: var(--selection-bg);
    color: var(--text-primary);
    font-weight: var(--fw-medium);
  }

  .nav-icon {
    display: inline-flex;
    color: var(--text-tertiary);
  }

  .nav-item.active .nav-icon {
    color: var(--accent);
  }

  .content {
    overflow-y: auto;
    padding: var(--space-5) var(--space-6) var(--space-6);
  }

  section {
    max-width: 560px;
  }

  h3 {
    margin: 0 0 var(--space-4);
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-2) 0;
  }

  .field {
    padding: var(--space-3) 0;
  }

  .row-label {
    font-size: var(--fz-body);
    color: var(--text-primary);
  }

  .hint {
    margin: var(--space-2) 0 0;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  /* generated-style chooser: a small grid of previewed cards. */
  .style-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: var(--space-3);
    margin-top: var(--space-3);
  }

  .style-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-raised);
    cursor: pointer;
  }

  .style-card:hover {
    background: var(--surface-hover);
  }

  .style-card.active {
    border-color: var(--accent);
    box-shadow: 0 0 0 1px var(--accent);
  }

  .style-img {
    width: 48px;
    height: 48px;
    border-radius: 999px;
  }

  .style-label {
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .action-btn {
    padding: var(--space-2) var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .action-btn:hover {
    background: var(--surface-hover);
  }

  .warn {
    margin-top: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border: var(--hairline) solid var(--warning);
    border-radius: var(--radius-card);
    background: var(--warning-bg);
  }

  .warn p {
    margin: 0 0 var(--space-3);
    font-size: var(--fz-label);
    color: var(--text-primary);
    line-height: 1.5;
  }

  .warn-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-3);
  }

  .ghost-btn,
  .danger-btn {
    padding: var(--space-2) var(--space-4);
    border-radius: var(--radius-control);
    font-size: var(--fz-label);
    cursor: pointer;
    border: var(--hairline) solid var(--border-default);
  }

  .ghost-btn {
    background: var(--surface-raised);
    color: var(--text-primary);
  }

  .danger-btn {
    background: var(--danger);
    color: #fff;
    border-color: transparent;
  }

  .toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-2) 0;
    cursor: pointer;
  }

  .toggle.disabled {
    opacity: 0.45;
    cursor: default;
  }
</style>
