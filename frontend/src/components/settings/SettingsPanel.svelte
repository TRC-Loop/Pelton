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
    IconInfoCircle,
    IconHandMove,
    IconCloudDownload,
    IconFileImport,
    IconAddressBook,
    IconMailbox,
    IconWriting,
    IconLanguage,
    IconBatteryEco,
  } from '@tabler/icons-svelte'
  import { createEventDispatcher } from 'svelte'
  import SegmentedSetting from './SegmentedSetting.svelte'
  import StepSlider from './StepSlider.svelte'
  import AccentPicker from './AccentPicker.svelte'
  import TechToggles from './TechToggles.svelte'
  import ToastPositionPicker from './ToastPositionPicker.svelte'
  import ShortcutSettings from './ShortcutSettings.svelte'
  import SignaturesSection from './SignaturesSection.svelte'
  import AddressBookSection from './AddressBookSection.svelte'
  import MailboxesSection from './MailboxesSection.svelte'
  import ImportExportSection from './ImportExportSection.svelte'
  import RowLayoutPreview from './RowLayoutPreview.svelte'
  import AboutSection from './AboutSection.svelte'
  import ToggleSwitch from '../common/ToggleSwitch.svelte'
  import LanguageSelect from '../common/LanguageSelect.svelte'
  import DateTimePicker from '../common/DateTimePicker.svelte'
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
    setFlagColorSync,
    setShowOfflineIndicator,
    setSwipeEnabled,
    setSwipeLeftAction,
    setSwipeRightAction,
    setComposeVimMode,
    setAppVimMode,
    setDownloadIncludeAttachments,
    setLanguage,
    setLowPowerMode,
    setAutoSyncInterval,
    setDefaultEditorMode,
    setComposeAutocomplete,
    setComposeChips,
    setEmptyStateImage,
  } from '../../stores/prefs'
  import peltonLogo from '../../assets/images/icons/pelton-logo.png'
  import type { Locale } from '../../lib/i18n'
  import { downloadRange, cancelDownload } from '../../lib/api'
  import { downloadProgress } from '../../stores/progress'
  import { toastError, errorMessage } from '../../stores/toast'
  import { t } from '../../lib/i18n'
  import type { ThemePref, DensityPref, EditorMode } from '../../lib/types'

  let editorModeOptions: { key: EditorMode; label: string }[] = []
  $: editorModeOptions = [
    { key: 'plaintext', label: $t('settingsPanel.editorMode.plaintext') },
    { key: 'markdown', label: $t('settingsPanel.editorMode.markdown') },
    { key: 'wysiwyg', label: $t('settingsPanel.editorMode.wysiwyg') },
  ]

  const dispatch = createEventDispatcher<{ close: void; rerunOnboarding: void }>()
  $: currentLocale = $prefs.language as Locale

  // left-nav categories. each maps to a block rendered on the right.
  $: categories = [
    { key: 'appearance', label: $t('settingsPanel.category.appearance'), icon: IconPalette },
    { key: 'language', label: $t('settings.language'), icon: IconLanguage },
    { key: 'list', label: $t('settingsPanel.category.list'), icon: IconList },
    { key: 'sidebar', label: $t('settingsPanel.category.sidebar'), icon: IconLayoutSidebar },
    { key: 'avatars', label: $t('settingsPanel.category.avatars'), icon: IconUserCircle },
    { key: 'signatures', label: $t('settingsPanel.category.signatures'), icon: IconSignature },
    { key: 'sending', label: $t('settingsPanel.category.sending'), icon: IconSend2 },
    { key: 'privacy', label: $t('settingsPanel.category.privacy'), icon: IconShieldLock },
    { key: 'notifications', label: $t('settingsPanel.category.notifications'), icon: IconBell },
    { key: 'panes', label: $t('settings.panes'), icon: IconLayoutColumns },
    { key: 'display', label: $t('settingsPanel.category.display'), icon: IconEye },
    { key: 'gestures', label: $t('settingsPanel.category.gestures'), icon: IconHandMove },
    { key: 'offline', label: $t('settingsPanel.category.offline'), icon: IconCloudDownload },
    { key: 'power', label: $t('settingsPanel.category.power'), icon: IconBatteryEco },
    { key: 'mailboxes', label: $t('settingsPanel.category.mailboxes'), icon: IconMailbox },
    { key: 'contacts', label: $t('settingsPanel.category.contacts'), icon: IconAddressBook },
    { key: 'sync', label: $t('settingsPanel.category.importExport'), icon: IconFileImport },
    { key: 'composing', label: $t('settingsPanel.category.composing'), icon: IconWriting },
    { key: 'shortcuts', label: $t('settingsPanel.category.shortcuts'), icon: IconKeyboard },
    { key: 'about', label: $t('settingsPanel.category.about'), icon: IconInfoCircle },
  ]

  // auto-sync interval presets, in seconds (0 = off).
  $: autoSyncOptions = [
    { key: '0', label: $t('settingsPanel.unit.off') },
    { key: '30', label: $t('settingsPanel.unit.s30') },
    { key: '300', label: $t('settingsPanel.unit.m5') },
    { key: '900', label: $t('settingsPanel.unit.m15') },
    { key: '1800', label: $t('settingsPanel.unit.m30') },
    { key: '3600', label: $t('settingsPanel.unit.h1') },
    { key: '21600', label: $t('settingsPanel.unit.h6') },
    { key: '43200', label: $t('settingsPanel.unit.h12') },
    { key: '86400', label: $t('settingsPanel.unit.h24') },
  ]
  function onAutoSyncInterval(event: CustomEvent<string>): void {
    setAutoSyncInterval(Number(event.detail))
  }

  // swipe gesture actions (trackpad). shown in the two direction dropdowns.
  $: swipeActionOptions = [
    { key: 'none', label: $t('settingsPanel.swipeAction.none') },
    { key: 'delete', label: $t('settingsPanel.swipeAction.delete') },
    { key: 'read', label: $t('settingsPanel.swipeAction.read') },
    { key: 'unread', label: $t('settingsPanel.swipeAction.unread') },
    { key: 'flag', label: $t('settingsPanel.swipeAction.flag') },
    { key: 'archive', label: $t('settingsPanel.swipeAction.archive') },
    { key: 'snooze', label: $t('settingsPanel.swipeAction.snooze') },
  ]

  // offline range download state. the start date defaults to one year ago.
  let downloadStart = defaultDownloadStart()
  function defaultDownloadStart(): string {
    const d = new Date()
    d.setFullYear(d.getFullYear() - 1)
    return d.toISOString().slice(0, 10)
  }

  // quick presets sit next to the native date input so picking a common range
  // (rather than fiddling with the bare calendar widget) is one click.
  $: downloadPresets = [
    { key: '1w', label: $t('settingsPanel.preset.lastWeek'), days: 7 },
    { key: '1m', label: $t('settingsPanel.preset.lastMonth'), days: 30 },
    { key: '3m', label: $t('settingsPanel.preset.last3Months'), days: 90 },
    { key: '6m', label: $t('settingsPanel.preset.last6Months'), days: 180 },
    { key: '1y', label: $t('settingsPanel.preset.lastYear'), days: 365 },
    { key: 'all', label: $t('settingsPanel.preset.allTime'), days: 0 },
  ]
  function applyPreset(days: number): void {
    const d = new Date()
    if (days === 0) {
      d.setFullYear(d.getFullYear() - 20) // effectively "everything"
    } else {
      d.setDate(d.getDate() - days)
    }
    downloadStart = d.toISOString().slice(0, 10)
  }

  async function startDownload(): Promise<void> {
    if (!downloadStart) {
      return
    }
    try {
      await downloadRange(downloadStart, $prefs.downloadIncludeAttachments)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // select handlers (the cast lives in script; inline ts casts break the parser).
  function onSwipeLeft(event: Event): void {
    setSwipeLeftAction((event.currentTarget as HTMLSelectElement).value)
  }
  function onSwipeRight(event: Event): void {
    setSwipeRightAction((event.currentTarget as HTMLSelectElement).value)
  }
  // initialCategory deep-links the panel to a section (e.g. opened from the
  // "Manage Mailboxes" menu item); null opens the default section.
  export let initialCategory: string | null = null
  let active = initialCategory ?? 'appearance'

  $: themeOptions = [
    { key: 'system', label: $t('onboarding.theme.system') },
    { key: 'light', label: $t('onboarding.theme.light') },
    { key: 'dark', label: $t('onboarding.theme.dark') },
  ]

  $: densityOptions = [
    { key: 'compact', label: $t('onboarding.density.compact') },
    { key: 'medium', label: $t('onboarding.density.medium') },
    { key: 'luxe', label: $t('onboarding.density.luxe') },
  ]

  // interface zoom. values are string multipliers applied as css zoom.
  $: scaleOptions = [
    { key: '0.9', label: $t('settingsPanel.scale.90') },
    { key: '1', label: $t('settingsPanel.scale.100') },
    { key: '1.1', label: $t('settingsPanel.scale.110') },
    { key: '1.17', label: $t('settingsPanel.scale.117') },
    { key: '1.25', label: $t('settingsPanel.scale.125') },
    { key: '1.5', label: $t('settingsPanel.scale.150') },
  ]

  // base font size (px) for rendered email content.
  $: messageFontOptions = [
    { key: '12', label: $t('onboarding.font.small') },
    { key: '14', label: $t('onboarding.font.default') },
    { key: '16', label: $t('onboarding.font.large') },
    { key: '18', label: $t('onboarding.font.larger') },
    { key: '20', label: $t('onboarding.font.largest') },
  ]

  $: sendDelayOptions = [
    { key: '0', label: $t('settingsPanel.unit.off') },
    { key: '5', label: $t('settingsPanel.unit.s5') },
    { key: '10', label: $t('settingsPanel.unit.s10') },
    { key: '30', label: $t('settingsPanel.unit.s30') },
    { key: '60', label: $t('settingsPanel.unit.s60') },
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

  // the remote-image allowlist manager (trusted senders/domains) opens in a modal.
  let allowlistOpen = false

  // the reading-pane empty-state image is picked from a local file and stored as
  // a data uri. anything past the hard cap is refused; between the soft and hard
  // caps we warn ("here be dragons") but let the user proceed, since a large data
  // uri in settings can slow the ui down.
  let emptyImageInput: HTMLInputElement
  const maxEmptyImageBytes = 50_000_000
  const warnEmptyImageBytes = 3_000_000
  // a data uri awaiting confirmation because the chosen file is large.
  let dragonsPending: string | null = null
  function onPickEmptyImage(event: Event): void {
    const input = event.currentTarget as HTMLInputElement
    const file = input.files?.[0]
    input.value = ''
    if (!file) {
      return
    }
    if (file.size > maxEmptyImageBytes) {
      toastError($t('settingsPanel.error.imageTooLarge'))
      return
    }
    const large = file.size > warnEmptyImageBytes
    const reader = new FileReader()
    reader.onload = () => {
      const uri = String(reader.result)
      if (large) {
        dragonsPending = uri
      } else {
        setEmptyStateImage(uri)
      }
    }
    reader.onerror = () => toastError($t('settingsPanel.error.imageRead'))
    reader.readAsDataURL(file)
  }
  function confirmDragons(): void {
    if (dragonsPending) {
      setEmptyStateImage(dragonsPending)
    }
    dragonsPending = null
  }

  // sender-photo fallback chain. "Generated" never touches the network.
  $: avatarSourceOptions = [
    { key: 'bimi_gravatar', label: $t('settingsPanel.avatarSource.bimiGravatar') },
    { key: 'gravatar_bimi', label: $t('settingsPanel.avatarSource.gravatarBimi') },
    { key: 'pfp', label: $t('settingsPanel.avatarSource.generated') },
  ]

  // generated placeholder styles, previewed with a sample sender so the look is
  // obvious before choosing.
  const sampleEmail = 'potato@pelton.email'
  const sampleInitials = initials('', sampleEmail)
  let avatarStyleOptions: { key: PfpStyle; label: string }[] = []
  $: avatarStyleOptions = [
    { key: 'initials', label: $t('onboarding.avatar.classic') },
    { key: 'mono', label: $t('onboarding.avatar.mono') },
    { key: 'pixel', label: $t('onboarding.avatar.pixel') },
    { key: 'geometric', label: $t('onboarding.avatar.geometric') },
  ]
  function stylePreview(style: PfpStyle): string {
    return pfpDataUri(style, sampleEmail, sampleInitials)
  }

  $: flagOptions = [
    { key: 'flag', label: $t('onboarding.flagopt.icon') },
    { key: 'left', label: $t('onboarding.flagopt.left') },
    { key: 'both', label: $t('onboarding.flagopt.both') },
    { key: 'off', label: $t('onboarding.flagopt.off') },
  ]

  $: rowTemplateOptions = [
    { key: 'relaxed', label: $t('onboarding.row.relaxed') },
    { key: 'comfortable', label: $t('onboarding.row.comfortable') },
    { key: 'compact', label: $t('onboarding.row.compact') },
    { key: 'single', label: $t('onboarding.row.single') },
  ]

  $: previewLineOptions = [
    { key: '1', label: $t('settingsPanel.previewLines.1') },
    { key: '2', label: $t('settingsPanel.previewLines.2') },
    { key: '3', label: $t('settingsPanel.previewLines.3') },
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

<div class="screen" role="dialog" aria-modal="true" aria-label={$t('settings.title')}>
  <header class="head">
    <h2>{$t('settings.title')}</h2>
    <button type="button" class="close" aria-label={$t('settingsPanel.closeAria')} on:click={() => dispatch('close')}>
      <IconX size={20} stroke={1.8} />
    </button>
  </header>

  <div class="body">
    <nav class="nav" aria-label={$t('settingsPanel.navAria')}>
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
      {#if active === 'language'}
        <section>
          <h3>{$t('settings.language')}</h3>
          <p class="hint">{$t('settings.languageHint')}</p>
          <LanguageSelect value={currentLocale} onSelect={setLanguage} />
        </section>
      {:else if active === 'appearance'}
        <section>
          <h3>{$t('settingsPanel.category.appearance')}</h3>
          <SegmentedSetting label={$t('settingsPanel.label.theme')} value={$prefs.theme} options={themeOptions} on:change={onTheme} />
          <AccentPicker />
          <SegmentedSetting label={$t('settingsPanel.label.density')} value={$prefs.density} options={densityOptions} on:change={onDensity} />
          <SegmentedSetting
            label={$t('settingsPanel.label.interfaceScale')}
            value={$prefs.uiScale}
            options={scaleOptions}
            on:change={(e) => setUIScale(e.detail)}
          />
          <p class="hint">{$t('settingsPanel.hint.interfaceScale')}</p>

          <div class="field">
            <span class="row-label">{$t('settingsPanel.label.emptyStateImage')}</span>
            <p class="hint">{$t('settingsPanel.hint.emptyStateImage')}</p>
            <div class="empty-image-row">
              <div class="empty-image-preview">
                <img src={$prefs.emptyStateImage || peltonLogo} alt="" draggable="false" />
              </div>
              <div class="empty-image-actions">
                <button type="button" class="action-btn" on:click={() => emptyImageInput?.click()}>
                  {$t('settingsPanel.button.selectImage')}
                </button>
                <button
                  type="button"
                  class="action-btn"
                  disabled={!$prefs.emptyStateImage}
                  on:click={() => setEmptyStateImage('')}
                >
                  {$t('settingsPanel.button.resetImage')}
                </button>
              </div>
            </div>
            <input
              class="hidden-file"
              type="file"
              accept="image/*"
              bind:this={emptyImageInput}
              on:change={onPickEmptyImage}
            />
            {#if dragonsPending}
              <div class="warn">
                <p>{$t('settingsPanel.warn.imageLarge')}</p>
                <div class="warn-actions">
                  <button type="button" class="ghost-btn" on:click={() => (dragonsPending = null)}>{$t('settingsPanel.button.cancel')}</button>
                  <button type="button" class="danger-btn" on:click={confirmDragons}>{$t('settingsPanel.button.useAnyway')}</button>
                </div>
              </div>
            {/if}
          </div>
        </section>
      {:else if active === 'list'}
        <section>
          <h3>{$t('settingsPanel.category.list')}</h3>
          <SegmentedSetting
            label={$t('settingsPanel.label.rowLayout')}
            value={$prefs.rowTemplate}
            options={rowTemplateOptions}
            on:change={(e) => setRowTemplate(e.detail)}
          />
          <RowLayoutPreview />
          <div class="toggle" class:disabled={$prefs.rowTemplate === 'single'} title={$t('settingsPanel.hint.avatarHiddenSingleLine')}>
            <span class="row-label">{$t('settingsPanel.toggle.showSenderAvatar')}</span>
            <ToggleSwitch
              checked={$prefs.rowShowAvatar}
              disabled={$prefs.rowTemplate === 'single'}
              label={$t('settingsPanel.toggle.showSenderAvatar')}
              on:change={(e) => setRowShowAvatar(e.detail)}
            />
          </div>
          <div class="toggle" class:disabled={!snippetCapable} title={$t('settingsPanel.hint.previewShowsOn')}>
            <span class="row-label">{$t('settingsPanel.toggle.showMessagePreview')}</span>
            <ToggleSwitch
              checked={$prefs.rowShowSnippet}
              disabled={!snippetCapable}
              label={$t('settingsPanel.toggle.showMessagePreview')}
              on:change={(e) => setRowShowSnippet(e.detail)}
            />
          </div>
          {#if snippetCapable && $prefs.rowShowSnippet}
            <SegmentedSetting
              label={$t('settingsPanel.label.previewLines')}
              value={String($prefs.previewLines)}
              options={previewLineOptions}
              on:change={(e) => setPreviewLines(Number(e.detail))}
            />
          {/if}
          <SegmentedSetting
            label={$t('settingsPanel.label.flaggedHighlight')}
            value={$prefs.flagHighlight}
            options={flagOptions}
            on:change={(e) => setFlagHighlight(e.detail)}
          />
          <p class="hint">{$t('settingsPanel.hint.barIconFlag')}</p>
          <div class="toggle">
            <span class="row-label">{$t('settingsPanel.toggle.showMailboxEmail')}</span>
            <ToggleSwitch
              checked={$prefs.showAccountEmail}
              label={$t('settingsPanel.toggle.showMailboxEmail')}
              on:change={(e) => setShowAccountEmail(e.detail)}
            />
          </div>
          <div class="toggle" title={$t('settingsPanel.hint.multiSelect')}>
            <span class="row-label">{$t('settingsPanel.toggle.multiSelect')}</span>
            <ToggleSwitch
              checked={$prefs.multiSelectEnabled}
              label={$t('settingsPanel.toggle.multiSelect')}
              on:change={(e) => setMultiSelectEnabled(e.detail)}
            />
          </div>
          <div class="toggle" class:disabled={!$prefs.multiSelectEnabled}>
            <span class="row-label">{$t('settingsPanel.toggle.showSelectedCount')}</span>
            <ToggleSwitch
              checked={$prefs.showSelectedCount}
              disabled={!$prefs.multiSelectEnabled}
              label={$t('settingsPanel.toggle.showSelectedCountAria')}
              on:change={(e) => setShowSelectedCount(e.detail)}
            />
          </div>
        </section>
      {:else if active === 'sidebar'}
        <section>
          <h3>{$t('settingsPanel.category.sidebar')}</h3>
          <div class="toggle" title={$t('settingsPanel.hint.indentGuides')}>
            <span class="row-label">{$t('settingsPanel.toggle.indentGuides')}</span>
            <ToggleSwitch
              checked={$prefs.sidebarIndentGuides}
              label={$t('settingsPanel.toggle.indentGuides')}
              on:change={(e) => setSidebarIndentGuides(e.detail)}
            />
          </div>
          <div class="toggle" title={$t('settingsPanel.hint.flaggedCount')}>
            <span class="row-label">{$t('settingsPanel.toggle.flaggedCount')}</span>
            <ToggleSwitch
              checked={$prefs.showFlaggedCount}
              label={$t('settingsPanel.toggle.flaggedCount')}
              on:change={(e) => setShowFlaggedCount(e.detail)}
            />
          </div>
        </section>
      {:else if active === 'avatars'}
        <section>
          <h3>{$t('settingsPanel.category.avatars')}</h3>
          <SegmentedSetting
            label={$t('settingsPanel.label.senderPhotos')}
            value={$prefs.avatarSource}
            options={avatarSourceOptions}
            on:change={(e) => setAvatarSource(e.detail)}
          />
          <p class="hint">
            {$t('settingsPanel.hint.avatarSource')}
          </p>

          <div class="field">
            <span class="row-label">{$t('settingsPanel.label.generatedStyle')}</span>
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
            <p class="hint">{$t('settingsPanel.hint.previewSender')} {sampleEmail}</p>
          </div>
        </section>
      {:else if active === 'signatures'}
        <section>
          <SignaturesSection />
        </section>
      {:else if active === 'sending'}
        <section>
          <h3>{$t('settingsPanel.category.sending')}</h3>
          <SegmentedSetting
            label={$t('settingsPanel.label.undoSendWindow')}
            value={String($prefs.sendDelaySeconds)}
            options={sendDelayOptions}
            on:change={onSendDelay}
          />
          <p class="hint">{$t('settingsPanel.hint.undoSend')}</p>
        </section>
      {:else if active === 'privacy'}
        <section>
          <h3>{$t('settingsPanel.category.privacy')}</h3>
          <div class="toggle" title={$t('settingsPanel.hint.remoteImagesToggle')}>
            <span class="row-label">{$t('settingsPanel.toggle.alwaysLoadImages')}</span>
            <ToggleSwitch
              checked={$prefs.alwaysLoadImages}
              label={$t('settingsPanel.toggle.alwaysLoadImages')}
              on:change={(e) => onImagesToggle(e.detail)}
            />
          </div>
          {#if confirmImages}
            <div class="warn">
              <p>
                {$t('settingsPanel.warn.remoteImages')}
              </p>
              <div class="warn-actions">
                <button type="button" class="ghost-btn" on:click={() => (confirmImages = false)}>{$t('settingsPanel.button.cancel')}</button>
                <button type="button" class="danger-btn" on:click={confirmEnableImages}>{$t('settingsPanel.button.enableAnyway')}</button>
              </div>
            </div>
          {/if}

          <div class="field">
            <span class="row-label">{$t('settingsPanel.label.manageWhitelist')}</span>
            <p class="hint">{$t('settingsPanel.hint.manageWhitelist')}</p>
            <button type="button" class="action-btn" on:click={() => (allowlistOpen = true)}>
              {$t('settingsPanel.button.manageWhitelist')}
            </button>
          </div>
        </section>
      {:else if active === 'notifications'}
        <section>
          <h3>{$t('settingsPanel.category.notifications')}</h3>
          <div class="row">
            <span class="row-label">{$t('settings.toastPosition')}</span>
            <ToastPositionPicker
              value={$prefs.toastPosition}
              on:change={(e) => setToastPosition(e.detail)}
            />
          </div>
        </section>
      {:else if active === 'panes'}
        <section>
          <h3>{$t('settings.panes')}</h3>
          <div class="toggle" title={$t('settingsPanel.hint.lockPanes')}>
            <span class="row-label">{$t('settings.lockPanes')}</span>
            <ToggleSwitch
              checked={$prefs.paneLocked}
              label={$t('settings.lockPanes')}
              on:change={(e) => setPaneLocked(e.detail)}
            />
          </div>
        </section>
      {:else if active === 'display'}
        <section>
          <h3>{$t('settingsPanel.category.display')}</h3>
          <SegmentedSetting
            label={$t('onboarding.extras.fontSize')}
            value={String($prefs.messageFontSize)}
            options={messageFontOptions}
            on:change={(e) => setMessageFontSize(Number(e.detail))}
          />
          <p class="hint">{$t('settingsPanel.hint.fontSize')}</p>
          <TechToggles />
        </section>
      {:else if active === 'gestures'}
        <section>
          <h3>{$t('settingsPanel.heading.gestures')}</h3>
          <div class="toggle" title={$t('settingsPanel.hint.swipeEnable')}>
            <span class="row-label">{$t('settingsPanel.toggle.swipeEnabled')}</span>
            <ToggleSwitch
              checked={$prefs.swipeEnabled}
              label={$t('settingsPanel.toggle.swipeEnabled')}
              on:change={(e) => setSwipeEnabled(e.detail)}
            />
          </div>
          <div class="row" class:disabled={!$prefs.swipeEnabled}>
            <span class="row-label">{$t('settingsPanel.label.swipeLeft')}</span>
            <select
              class="select"
              disabled={!$prefs.swipeEnabled}
              value={$prefs.swipeLeftAction}
              on:change={onSwipeLeft}
            >
              {#each swipeActionOptions as opt}
                <option value={opt.key}>{opt.label}</option>
              {/each}
            </select>
          </div>
          <div class="row" class:disabled={!$prefs.swipeEnabled}>
            <span class="row-label">{$t('settingsPanel.label.swipeRight')}</span>
            <select
              class="select"
              disabled={!$prefs.swipeEnabled}
              value={$prefs.swipeRightAction}
              on:change={onSwipeRight}
            >
              {#each swipeActionOptions as opt}
                <option value={opt.key}>{opt.label}</option>
              {/each}
            </select>
          </div>
          <p class="hint">{$t('settingsPanel.hint.swipeWork')}</p>
        </section>
      {:else if active === 'power'}
        <section>
          <h3>{$t('settingsPanel.category.power')}</h3>
          <div class="toggle" title={$t('settingsPanel.hint.lowPowerToggle')}>
            <span class="row-label">{$t('settingsPanel.toggle.lowPowerMode')}</span>
            <ToggleSwitch
              checked={$prefs.lowPowerMode}
              label={$t('settingsPanel.toggle.lowPowerMode')}
              on:change={(e) => setLowPowerMode(e.detail)}
            />
          </div>
          <p class="hint">
            {$t('settingsPanel.hint.lowPowerDetail')}
          </p>
          <StepSlider
            label={$t('settingsPanel.label.autoSyncInterval')}
            value={String($prefs.autoSyncIntervalSeconds)}
            options={autoSyncOptions}
            on:change={onAutoSyncInterval}
          />
          <p class="hint">
            {$t('settingsPanel.hint.autoSyncDetail')}
          </p>
        </section>
      {:else if active === 'offline'}
        <section>
          <h3>{$t('settingsPanel.category.offline')}</h3>
          <div class="toggle" title={$t('settingsPanel.hint.offlineIndicator')}>
            <span class="row-label">{$t('settingsPanel.toggle.offlineIndicator')}</span>
            <ToggleSwitch
              checked={$prefs.showOfflineIndicator}
              label={$t('settingsPanel.toggle.offlineIndicator')}
              on:change={(e) => setShowOfflineIndicator(e.detail)}
            />
          </div>
          <div class="toggle" title={$t('settingsPanel.hint.flagColorSync')}>
            <span class="row-label">{$t('settingsPanel.toggle.flagColorSync')}</span>
            <ToggleSwitch
              checked={$prefs.flagColorSync}
              label={$t('settingsPanel.toggle.flagColorSync')}
              on:change={(e) => setFlagColorSync(e.detail)}
            />
          </div>

          <div class="field">
            <span class="row-label">{$t('settingsPanel.label.downloadOffline')}</span>
            <p class="hint">
              {$t('settingsPanel.hint.downloadOffline')}
            </p>
            <div class="download-presets">
              {#each downloadPresets as p (p.key)}
                <button type="button" class="preset-btn" on:click={() => applyPreset(p.days)}>{p.label}</button>
              {/each}
            </div>
            <div class="download-row">
              <div class="download-date">
                <DateTimePicker mode="date" bind:value={downloadStart} />
              </div>
              {#if $downloadProgress && $downloadProgress.running}
                <button type="button" class="action-btn" on:click={() => cancelDownload()}>
                  {$t('settingsPanel.button.cancelDownload')}
                </button>
              {:else}
                <button type="button" class="action-btn" on:click={startDownload}>
                  {$t('settingsPanel.button.download')}
                </button>
              {/if}
            </div>
            <div class="toggle" title={$t('settingsPanel.hint.includeAttachments')}>
              <span class="row-label">{$t('settingsPanel.toggle.includeAttachments')}</span>
              <ToggleSwitch
                checked={$prefs.downloadIncludeAttachments}
                label={$t('settingsPanel.toggle.includeAttachments')}
                on:change={(e) => setDownloadIncludeAttachments(e.detail)}
              />
            </div>
          </div>
        </section>
      {:else if active === 'mailboxes'}
        <section>
          <MailboxesSection />
        </section>
      {:else if active === 'contacts'}
        <section>
          <AddressBookSection />
        </section>
      {:else if active === 'sync'}
        <section>
          <h3>{$t('settingsPanel.category.importExport')}</h3>
          <ImportExportSection />
        </section>
      {:else if active === 'composing'}
        <section>
          <h3>{$t('settingsPanel.category.composing')}</h3>
          <SegmentedSetting
            label={$t('settingsPanel.label.defaultEditor')}
            value={$prefs.defaultEditorMode}
            options={editorModeOptions}
            on:change={(e) => setDefaultEditorMode(e.detail as EditorMode)}
          />
          <p class="hint">{$t('settingsPanel.hint.defaultEditor')}</p>
          <div class="toggle" title={$t('settingsPanel.hint.autocomplete')}>
            <span class="row-label">{$t('settingsPanel.toggle.autocomplete')}</span>
            <ToggleSwitch
              checked={$prefs.composeAutocomplete}
              label={$t('settingsPanel.toggle.autocomplete')}
              on:change={(e) => setComposeAutocomplete(e.detail)}
            />
          </div>
          <div class="toggle" title={$t('settingsPanel.hint.chipRecipients')}>
            <span class="row-label">{$t('settingsPanel.toggle.chipRecipients')}</span>
            <ToggleSwitch
              checked={$prefs.composeChips}
              label={$t('settingsPanel.toggle.chipRecipients')}
              on:change={(e) => setComposeChips(e.detail)}
            />
          </div>
          {#if !$prefs.composeChips}
            <p class="hint">{$t('settingsPanel.hint.plainRecipients')}</p>
          {/if}
          <div class="toggle" title={$t('settingsPanel.hint.vimEditor')}>
            <span class="row-label">{$t('onboarding.extras.vimEditor')} <span class="badge-experimental">{$t('common.experimental')}</span></span>
            <ToggleSwitch
              checked={$prefs.composeVimMode}
              label={$t('onboarding.extras.vimEditor')}
              on:change={(e) => setComposeVimMode(e.detail)}
            />
          </div>
          <p class="hint">{$t('settingsPanel.hint.vimEditorDetail')}</p>
        </section>
      {:else if active === 'shortcuts'}
        <section>
          <h3>{$t('settings.shortcuts')}</h3>
          <div class="toggle">
            <span class="row-label">{$t('settingsPanel.toggle.shortcutHints')}</span>
            <ToggleSwitch
              checked={$prefs.showShortcutHints}
              label={$t('settingsPanel.toggle.shortcutHints')}
              on:change={(e) => setShortcutHints(e.detail)}
            />
          </div>
          <div class="toggle" title={$t('settingsPanel.hint.appVim')}>
            <span class="row-label">{$t('onboarding.extras.appVim')} <span class="badge-experimental">{$t('common.experimental')}</span></span>
            <ToggleSwitch
              checked={$prefs.appVimMode}
              label={$t('onboarding.extras.appVim')}
              on:change={(e) => setAppVimMode(e.detail)}
            />
          </div>
          <p class="hint">{$t('settingsPanel.hint.appVimDetail')}</p>
          <ShortcutSettings />
        </section>
      {:else if active === 'about'}
        <section>
          <h3>{$t('settingsPanel.category.about')}</h3>
          <AboutSection on:rerunOnboarding={() => dispatch('rerunOnboarding')} />
        </section>
      {/if}
    </div>
  </div>
</div>

<!-- the allowlist modal is code-split so its list logic loads only on demand. -->
{#if allowlistOpen}
  {#await import('./ImageAllowlistModal.svelte') then m}
    <svelte:component
      this={m.default}
      on:close={() => (allowlistOpen = false)}
      on:openMessage={() => {
        allowlistOpen = false
        dispatch('close')
      }}
    />
  {/await}
{/if}

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

  /* a small marker for features that are still rough around the edges. */
  .badge-experimental {
    display: inline-block;
    margin-left: var(--space-2);
    padding: 1px 6px;
    border-radius: var(--radius-control);
    background: var(--warning-bg, var(--surface-sunken));
    color: var(--warning, var(--text-secondary));
    font-size: var(--fz-meta);
    font-weight: var(--fw-semibold);
    text-transform: uppercase;
    letter-spacing: 0.03em;
    vertical-align: middle;
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

  .action-btn:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .empty-image-row {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    margin-top: var(--space-2);
  }

  .empty-image-preview {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 72px;
    height: 72px;
    flex-shrink: 0;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-sunken);
    overflow: hidden;
  }

  .empty-image-preview img {
    max-width: 100%;
    max-height: 100%;
    object-fit: contain;
  }

  .empty-image-actions {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    align-items: flex-start;
  }

  .hidden-file {
    display: none;
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

  .row.disabled {
    opacity: 0.45;
  }

  .select {
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font: inherit;
    cursor: pointer;
  }

  .download-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-top: var(--space-3);
  }

  .download-date {
    width: 160px;
  }

  .download-presets {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    margin-top: var(--space-3);
  }

  .preset-btn {
    padding: var(--space-1) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: 999px;
    background: var(--surface-raised);
    color: var(--text-secondary);
    font-size: var(--fz-meta);
    cursor: pointer;
  }

  .preset-btn:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }
</style>
