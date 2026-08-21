<script lang="ts">
  // managing profiles (#270): create, edit, delete, switch.
  //
  // The sharing model is one control per area with three choices, rather than a
  // "share?" switch and a separate "copy from?" step. Share is a live link to
  // the main profile, copy is a one-time duplicate, fresh is the defaults. That
  // is the whole model, and it reads the same when creating and when editing
  // (where copy has already happened, so it shows as its result: not shared).
  import { onMount } from 'svelte'
  import { IconPencil, IconTrash, IconPlus, IconCheck, IconX, IconStarFilled } from '@tabler/icons-svelte'
  import ToggleSwitch from '../common/ToggleSwitch.svelte'
  import { prefs, setPaletteProfiles } from '../../stores/prefs'
  import { profiles, currentProfile, loadProfiles, switchTo } from '../../stores/profiles'
  import { allAccounts, createProfile, updateProfile, deleteProfile } from '../../lib/api'
  import { errorMessage, toastError } from '../../stores/toast'
  import { t } from '../../lib/i18n'
  import type { Account, Profile, ProfileDraft, ProfileStart } from '../../lib/types'

  // the areas a profile can own or share, in the order the form lists them.
  const areas = ['settings', 'signatures', 'views'] as const
  type Area = (typeof areas)[number]

  const starts: ProfileStart[] = ['share', 'copy', 'fresh']

  // the icons on offer. A picker rather than a text field: this is a label you
  // recognise at a glance in the status bar, not somewhere to paste anything.
  const iconChoices = [
    // people and places
    '🧑', '👤', '👥', '🧑‍💻', '🧑‍🎓', '🧑‍🏫', '🧑‍⚕️', '🧑‍🔧', '🧑‍🍳', '🧑‍🎨',
    '🏠', '🏡', '🏢', '🏫', '🏥', '🏦', '🏭', '🏛️', '⛺', '🌆',
    // work and study
    '💼', '📁', '🗂️', '📋', '📌', '📎', '🖇️', '📐', '📏', '✏️',
    '🖊️', '🖋️', '📝', '📚', '📖', '🎓', '🧾', '💳', '💰', '📊',
    '📈', '📉', '🗓️', '⏰', '⌛', '🔔', '📣', '🗣️', '🤝', '🧠',
    // mail and comms
    '✉️', '📮', '📨', '📬', '📭', '📦', '📡', '☎️', '📱', '💬',
    // making and fixing
    '🛠️', '🔧', '🔩', '⚙️', '🧰', '🔬', '🧪', '⚗️', '🧬', '🔭',
    '💡', '🔌', '🔋', '🖥️', '💻', '⌨️', '🖱️', '🖨️', '💾', '🗄️',
    // art, music, play
    '🎨', '🖌️', '🎭', '🎬', '📷', '📹', '🎧', '🎵', '🎸', '🎹',
    '🥁', '🎤', '🎮', '🕹️', '🎲', '♟️', '🧩', '🎯', '🎳', '🏆',
    // sport and outdoors
    '⚽', '🏀', '🏈', '🎾', '🏐', '🏓', '🥊', '🚴', '🏃', '🧗',
    '🏊', '⛷️', '🏕️', '🥾', '🚵', '🛹', '🛼', '🪁', '🎣', '🧘',
    // travel
    '✈️', '🚀', '🚗', '🚌', '🚂', '🚢', '⛵', '🛵', '🚲', '🗺️',
    '🧳', '🏝️', '🏔️', '🌋', '🗽', '🎡', '🎢', '🌉', '⛱️', '🧭',
    // nature and weather
    '🌍', '🌙', '⭐', '🌟', '☀️', '⛅', '🌈', '❄️', '🔥', '💧',
    '🌊', '🌱', '🌳', '🌵', '🍀', '🌸', '🌻', '🍁', '🍄', '🪴',
    // animals
    '🐢', '🐝', '🐞', '🦊', '🐻', '🐼', '🐨', '🐧', '🦉', '🦅',
    '🐬', '🐙', '🦀', '🐈', '🐕', '🐎', '🦋', '🐳', '🦔', '🦕',
    // food and drink
    '🥔', '☕', '🍵', '🍺', '🍷', '🥂', '🍎', '🍌', '🍓', '🍇',
    '🍕', '🍔', '🌮', '🍜', '🍣', '🥐', '🥗', '🍪', '🎂', '🍫',
    // symbols
    '❤️', '💙', '💚', '💛', '💜', '🖤', '🤍', '🧡', '💫', '✨',
    '🔒', '🔑', '🛡️', '⚡', '♻️', '⚓', '🧿', '🔮', '🎁', '🏁',
  ]

  let accounts: Account[] = []
  let loading = true
  let draft: ProfileDraft | null = null
  // whether the open form is creating rather than editing, which decides
  // whether "copy" is still on offer.
  let creating = false
  let saving = false
  let confirmingId: number | null = null

  // editing the main profile: it owns the rows the others read, so it has no
  // sharing choices of its own to make.
  $: editingMain = !creating && $profiles.some((p) => p.id === draft?.id && p.main)

  onMount(load)

  async function load(): Promise<void> {
    loading = true
    try {
      accounts = await allAccounts()
      await loadProfiles()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      loading = false
    }
  }

  function startCreate(): void {
    confirmingId = null
    creating = true
    draft = {
      id: 0,
      name: '',
      icon: '',
      accountIds: [],
      startSettings: 'copy',
      startSignatures: 'fresh',
      startViews: 'fresh',
    }
  }

  function startEdit(profile: Profile): void {
    confirmingId = null
    creating = false
    draft = {
      id: profile.id,
      name: profile.name,
      icon: profile.icon,
      accountIds: [...profile.accountIds],
      // an area that is not shared is the profile's own, however it got there.
      startSettings: profile.shareSettings ? 'share' : 'fresh',
      startSignatures: profile.shareSignatures ? 'share' : 'fresh',
      startViews: profile.shareViews ? 'share' : 'fresh',
    }
  }

  function cancel(): void {
    draft = null
    creating = false
  }

  function setStart(area: Area, value: ProfileStart): void {
    if (!draft) {
      return
    }
    if (area === 'settings') {
      draft.startSettings = value
    } else if (area === 'signatures') {
      draft.startSignatures = value
    } else {
      draft.startViews = value
    }
  }

  function startOf(area: Area): ProfileStart {
    if (!draft) {
      return 'fresh'
    }
    return area === 'settings' ? draft.startSettings : area === 'signatures' ? draft.startSignatures : draft.startViews
  }

  // the row at the top of the form: one click to put every area on the same
  // footing, which is what most people want.
  function setAll(value: ProfileStart): void {
    for (const area of areas) {
      setStart(area, value)
    }
  }

  function toggleAccount(id: number): void {
    if (!draft) {
      return
    }
    draft.accountIds = draft.accountIds.includes(id)
      ? draft.accountIds.filter((a) => a !== id)
      : [...draft.accountIds, id]
  }

  async function save(): Promise<void> {
    if (!draft || saving || draft.name.trim() === '') {
      return
    }
    saving = true
    try {
      if (creating) {
        await createProfile(draft)
      } else {
        await updateProfile(draft)
      }
      await loadProfiles()
      cancel()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      saving = false
    }
  }

  async function confirmDelete(id: number): Promise<void> {
    try {
      await deleteProfile(id)
      confirmingId = null
      await loadProfiles()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  function accountLabel(account: Account): string {
    return account.displayName && account.displayName !== account.email
      ? `${account.displayName} (${account.email})`
      : account.email
  }
</script>

<div class="head">
  <div>
    <h3>{$t('settingsPanel.category.profiles')}</h3>
    <p class="hint">{$t('profiles.hint')}</p>
  </div>
  <button type="button" class="add-btn" on:click={startCreate}>
    <IconPlus size={14} stroke={2} />
    {$t('profiles.add')}
  </button>
</div>

<div class="toggle" title={$t('profiles.paletteHint')}>
  <span class="row-label">{$t('profiles.paletteToggle')}</span>
  <ToggleSwitch
    checked={$prefs.paletteProfiles}
    label={$t('profiles.paletteToggle')}
    on:change={(e) => setPaletteProfiles(e.detail)}
  />
</div>
<p class="hint">{$t('profiles.paletteHint')}</p>

{#if loading}
  <p class="empty">{$t('mailboxes.loading')}</p>
{:else}
  <ul class="list">
    {#each $profiles as profile (profile.id)}
      <li>
        <div class="who">
          <span class="name">
            {#if profile.icon}<span class="glyph">{profile.icon}</span>{/if}
            {profile.name}
            {#if profile.main}
              <span class="star" title={$t('profiles.mainTitle')} aria-label={$t('profiles.mainTitle')}>
                <IconStarFilled size={11} />
              </span>
            {/if}
          </span>
          <span class="meta">
            {#if profile.id === $currentProfile?.id}
              {$t('profiles.currentTag')}
            {:else}
              {$t('profiles.accountCount').replace('{n}', String(profile.accountIds.length))}
            {/if}
          </span>
        </div>

        {#if confirmingId === profile.id}
          <div class="confirm">
            <span class="warn">{$t('profiles.deleteConfirm')}</span>
            <button type="button" class="danger" on:click={() => confirmDelete(profile.id)}>{$t('action.delete')}</button>
            <button type="button" class="ghost" on:click={() => (confirmingId = null)}>{$t('mailboxes.cancel')}</button>
          </div>
        {:else}
          {#if profile.id !== $currentProfile?.id}
            <button type="button" class="ghost" on:click={() => void switchTo(profile.id)}>{$t('profiles.switchTo')}</button>
          {/if}
          <button type="button" class="icon" aria-label={`${$t('mailboxes.edit')} ${profile.name}`} on:click={() => startEdit(profile)}>
            <IconPencil size={15} stroke={1.7} />
          </button>
          {#if !profile.main}
            <button type="button" class="icon del" aria-label={`${$t('action.delete')} ${profile.name}`} on:click={() => (confirmingId = profile.id)}>
              <IconTrash size={15} stroke={1.7} />
            </button>
          {/if}
        {/if}
      </li>
    {/each}
  </ul>
{/if}

{#if draft}
  <div class="editor">
    <h4>{creating ? $t('profiles.newProfile') : $t('profiles.editProfile')}</h4>

    <div class="fields">
      <label class="field">
        <span>{$t('profiles.field.name')}</span>
        <input type="text" bind:value={draft.name} placeholder={$t('profiles.field.namePlaceholder')} />
      </label>
    </div>

    <h5>{$t('profiles.field.icon')}</h5>
    <div class="icons" role="radiogroup" aria-label={$t('profiles.field.icon')}>
      <button
        type="button"
        class="icon-choice none"
        class:on={draft.icon === ''}
        role="radio"
        aria-checked={draft.icon === ''}
        title={$t('profiles.icon.none')}
        aria-label={$t('profiles.icon.none')}
        on:click={() => draft && (draft.icon = '')}
      >
        <IconX size={14} stroke={2} />
      </button>
      {#each iconChoices as choice (choice)}
        <button
          type="button"
          class="icon-choice"
          class:on={draft.icon === choice}
          role="radio"
          aria-checked={draft.icon === choice}
          on:click={() => draft && (draft.icon = choice)}
        >
          {choice}
        </button>
      {/each}
    </div>

    <h5>{$t('profiles.accounts')}</h5>
    <p class="hint">{$t('profiles.accountsHint')}</p>
    {#if accounts.length === 0}
      <p class="empty">{$t('mailboxes.empty')}</p>
    {:else}
      <div class="accounts">
        {#each accounts as account (account.id)}
          <label class="check">
            <input
              type="checkbox"
              checked={draft.accountIds.includes(account.id)}
              on:change={() => toggleAccount(account.id)}
            />
            <span>{accountLabel(account)}</span>
          </label>
        {/each}
      </div>
    {/if}

    <!-- main is what the other profiles share from, so it has nothing to
         share with and the whole block is off the page for it. -->
    {#if !editingMain}
      <h5>{$t('profiles.sharing')}</h5>
      <p class="hint">{$t('profiles.sharingHint')}</p>
    {/if}

    {#if creating}
      <div class="row all">
        <span class="row-label">{$t('profiles.area.all')}</span>
        <div class="seg" role="group" aria-label={$t('profiles.area.all')}>
          {#each starts as value (value)}
            <button type="button" on:click={() => setAll(value)}>{$t(`profiles.start.${value}`)}</button>
          {/each}
        </div>
      </div>
    {/if}

    {#each editingMain ? [] : areas as area (area)}
      <div class="row">
        <span class="row-label">{$t(`profiles.area.${area}`)}</span>
        <div class="seg" role="radiogroup" aria-label={$t(`profiles.area.${area}`)}>
          {#each starts as value (value)}
            <!-- copying is a one-off, so it is only on offer while creating.
                 On an existing profile the choice left is shared or not. -->
            {#if creating || value !== 'copy'}
              <button type="button" class:on={startOf(area) === value} on:click={() => setStart(area, value)}>
                {$t(`profiles.start.${value}`)}
              </button>
            {/if}
          {/each}
        </div>
      </div>
    {/each}

    <div class="edit-actions">
      <button type="button" class="ghost" on:click={cancel}>
        <IconX size={14} stroke={2} />
        {$t('mailboxes.cancel')}
      </button>
      <button type="button" class="primary" disabled={saving || draft.name.trim() === ''} on:click={save}>
        <IconCheck size={14} stroke={2} />
        {saving ? $t('mailboxes.saving') : $t('mailboxes.save')}
      </button>
    </div>
  </div>
{/if}

<style>
  .head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
  }

  h3 {
    margin: 0 0 var(--space-3);
    font-size: var(--fz-heading);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  h4 {
    margin: 0 0 var(--space-3);
    font-size: var(--fz-title);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  h5 {
    margin: var(--space-4) 0 var(--space-1);
    font-size: var(--fz-label);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }

  .hint {
    margin: 0 0 var(--space-3);
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  .empty {
    padding: var(--space-3) 0;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }

  .add-btn {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    flex-shrink: 0;
    padding: var(--space-2) var(--space-4);
    border: none;
    border-radius: var(--radius-control);
    background: var(--accent);
    color: var(--accent-fg);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    cursor: var(--cursor-action);
  }
  .add-btn:hover {
    filter: brightness(1.05);
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  li {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-1);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .who {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
  }

  .name {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fz-body);
    color: var(--text-primary);
  }

  .meta {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .icon {
    flex-shrink: 0;
    padding: var(--space-1);
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    border-radius: var(--radius-control);
    cursor: var(--cursor-action);
  }
  .icon:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }
  .icon.del:hover {
    color: var(--danger);
  }

  .confirm {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .warn {
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .editor {
    margin-top: var(--space-4);
    padding: var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-sunken);
  }

  .fields {
    display: flex;
    gap: var(--space-3);
  }

  .toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-2) 0;
  }

  .star {
    display: inline-flex;
    color: var(--warning);
  }

  /* a big grid that scrolls rather than pushing the rest of the form down the
     page. Roomy enough to scan, capped so it never owns the screen. */
  .icons {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(30px, 1fr));
    gap: var(--space-1);
    max-height: 190px;
    overflow-y: auto;
    padding: var(--space-1);
    margin-bottom: var(--space-2);
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    scrollbar-width: thin;
  }

  .icon-choice {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    font-size: 15px;
    border: var(--hairline) solid transparent;
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-tertiary);
    cursor: var(--cursor-action);
  }
  .icon-choice:hover {
    border-color: var(--border-default);
  }
  .icon-choice.on {
    border-color: var(--accent);
    background: var(--surface-hover);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    flex: 1;
  }
  .field span {
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .field input {
    padding: var(--space-2) var(--space-3);
    font-family: var(--font-ui);
    font-size: var(--fz-body);
    color: var(--text-primary);
    background: var(--surface-raised);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
  }
  .field input:focus {
    outline: none;
    border-color: var(--accent);
  }

  .accounts {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }

  .check {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fz-label);
    color: var(--text-secondary);
    cursor: var(--cursor-action);
  }

  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-2) 0;
  }

  /* the quick control sits above the per-area rows and is separated from them,
     since it sets all three rather than being one of them. */
  .row.all {
    border-bottom: var(--hairline) solid var(--border-subtle);
    margin-bottom: var(--space-2);
  }

  .row-label {
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }

  .seg {
    display: inline-flex;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    overflow: hidden;
  }

  .seg button {
    padding: var(--space-1) var(--space-3);
    border: none;
    background: transparent;
    color: var(--text-secondary);
    font-size: var(--fz-meta);
    cursor: var(--cursor-action);
  }
  .seg button:hover {
    background: var(--surface-hover);
  }
  .seg button.on {
    background: var(--accent);
    color: var(--accent-fg);
  }

  .edit-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-top: var(--space-4);
  }

  .ghost,
  .danger,
  .primary {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
    border-radius: var(--radius-control);
    cursor: var(--cursor-action);
  }

  .ghost {
    border: var(--hairline) solid var(--border-default);
    background: transparent;
    color: var(--text-secondary);
  }
  .ghost:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .danger {
    border: none;
    background: var(--danger);
    color: var(--accent-fg);
  }

  .primary {
    border: none;
    background: var(--accent);
    color: var(--accent-fg);
    font-weight: var(--fw-medium);
  }
  .primary:disabled {
    opacity: 0.5;
    cursor: default;
  }
</style>
