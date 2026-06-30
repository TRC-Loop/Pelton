<script lang="ts">
  // the application shell: the three-column layout (resizable, lockable) plus the
  // compose panes, the settings screen, the outbox status bar and the toast
  // stack. it loads initial data, applies preferences, subscribes to backend and
  // menu events, and handles the app-wide keyboard shortcuts.
  import { onMount, onDestroy } from 'svelte'
  import { get } from 'svelte/store'

  import Sidebar from './components/sidebar/Sidebar.svelte'
  import MessageList from './components/list/MessageList.svelte'
  import MessageDetail from './components/detail/MessageDetail.svelte'
  import Compose from './components/compose/Compose.svelte'
  import Toasts from './components/common/Toasts.svelte'
  import StatusBar from './components/common/StatusBar.svelte'
  import ContextMenu from './components/common/ContextMenu.svelte'
  import Resizer from './components/common/Resizer.svelte'

  import { initPrefs, prefs, setPaneWidths } from './stores/prefs'
  import { loadSidebar, refreshSidebar, sidebar } from './stores/accounts'
  import { initSidebarState } from './stores/sidebarstate'
  import { loadSignatures } from './stores/signatures'
  import { loadOutbox, syncing, lastSynced } from './stores/outbox'
  import { selection } from './stores/selection'
  import { loadList } from './stores/messages'
  import { composeSessions, openCompose, initComposePrefs } from './stores/compose'
  import { triggerSync, getSetting, setSetting, SettingKeys, exportMessagePrintView } from './lib/api'
  import { onMailNew, onSyncState, onOutboxChanged, onMenu, type Unsubscribe } from './lib/events'
  import { matchShortcut, comboHasModifier, type ShortcutAction } from './lib/shortcuts'
  import { bindings, recording, initShortcuts } from './stores/shortcuts'
  import { triggerUndo } from './stores/undosend'
  import { openMessageId } from './stores/selection'
  import { errorMessage, toastError, toastInfo } from './stores/toast'

  let settingsOpen = false
  let wizardOpen = false
  let onboardingOpen = false
  const unsubscribers: Unsubscribe[] = []

  // live pane widths. they track the persisted prefs unless the user is mid-drag,
  // so a resize feels immediate and only commits on release.
  let sidebarW = 264
  let listW = 380
  let dragging = false
  $: if (!dragging) {
    sidebarW = $prefs.sidebarWidth
    listW = $prefs.listWidth
  }
  $: locked = $prefs.paneLocked

  onMount(async () => {
    await initPrefs()
    await initSidebarState()
    await initComposePrefs()
    void initShortcuts()
    void loadSignatures()
    await loadSidebar()
    await loadOutbox()

    // show the first-run onboarding until it has been completed once.
    try {
      const r = await getSetting(SettingKeys.onboarded)
      onboardingOpen = !(r.found && r.value === 'true')
    } catch {
      // if the lookup fails, do not block the app with onboarding.
      onboardingOpen = false
    }

    unsubscribers.push(
      onMailNew(() => {
        void refreshSidebar()
        void loadList(get(selection))
      }),
    )
    unsubscribers.push(
      onSyncState((e) => {
        syncing.set(e.running)
        // record the moment a sync finishes for the status bar's last-synced time.
        if (!e.running) {
          lastSynced.set(Date.now())
        }
      }),
    )
    unsubscribers.push(onOutboxChanged(() => void loadOutbox()))
    unsubscribers.push(onMenu(handleMenu))
  })

  onDestroy(() => {
    for (const off of unsubscribers) {
      off()
    }
  })

  function composeAccountId(): number | null {
    const data = get(sidebar).data
    if (!data || data.accounts.length === 0) {
      return null
    }
    const sel = get(selection)
    if (sel.kind === 'folder') {
      return sel.accountId
    }
    return data.accounts[0].id
  }

  function startCompose(): void {
    const accountId = composeAccountId()
    if (accountId === null) {
      toastError('Add a mailbox before composing.')
      return
    }
    openCompose(accountId, 'plaintext')
  }

  async function runSync(): Promise<void> {
    syncing.set(true)
    try {
      await triggerSync()
      await refreshSidebar()
      await loadList(get(selection))
      lastSynced.set(Date.now())
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      syncing.set(false)
    }
  }

  // add-mailbox opens the fullscreen wizard (lazy-loaded). once a mailbox is
  // added we reload the sidebar so the new account and its folders appear.
  function addMailbox(): void {
    wizardOpen = true
  }

  function onMailboxAdded(): void {
    wizardOpen = false
    void refreshSidebar()
    toastInfo('Mailbox added and syncing.')
  }

  // onboarding completion is persisted so it shows only once. re-run clears it
  // from settings and reopens the flow.
  function finishOnboarding(): void {
    onboardingOpen = false
    void setSetting(SettingKeys.onboarded, 'true')
  }

  function rerunOnboarding(): void {
    settingsOpen = false
    onboardingOpen = true
  }

  function onboardingAddedMailbox(): void {
    void refreshSidebar()
  }

  function focusSearch(): void {
    const input = document.querySelector<HTMLInputElement>('input[type="search"]')
    input?.focus()
  }

  // export the currently open message to a print/pdf view, or tell the user to
  // open one first.
  function exportPdf(): void {
    const id = get(openMessageId)
    if (id === null) {
      toastInfo('Open a message first to export it.')
      return
    }
    exportMessagePrintView(id).catch((err) => toastError(errorMessage(err)))
  }

  // dispatch maps an action (from a shortcut or a menu item) to its handler.
  function dispatchAction(action: ShortcutAction | 'about' | 'export-pdf'): void {
    switch (action) {
      case 'compose':
        startCompose()
        break
      case 'export-pdf':
        exportPdf()
        break
      case 'preferences':
        settingsOpen = true
        break
      case 'sync':
        void runSync()
        break
      case 'add-mailbox':
        addMailbox()
        break
      case 'search':
        focusSearch()
        break
      case 'about':
        toastInfo('Pelton, an open-source mail client.')
        break
    }
  }

  function handleMenu(action: string): void {
    dispatchAction(action as ShortcutAction | 'about' | 'export-pdf')
  }

  // suppress the webview's default context menu (inspect/reload) everywhere. the
  // one exception is when text is selected, so the native copy menu still works
  // for selected mail text. components that want a real menu (the message list)
  // open the custom one themselves.
  function onContextMenu(event: MouseEvent): void {
    const selected = window.getSelection()?.toString().trim()
    if (selected) {
      return
    }
    event.preventDefault()
  }

  function onKeydown(event: KeyboardEvent): void {
    // while the settings panel is capturing a new binding, let it have the keys.
    if ($recording) {
      return
    }
    // cmd/ctrl+z undoes a pending delayed send, when one is in its window. it
    // takes priority over other shortcuts and is swallowed only if it acted.
    if ((event.metaKey || event.ctrlKey) && !event.shiftKey && event.key.toLowerCase() === 'z') {
      if (triggerUndo()) {
        event.preventDefault()
        return
      }
    }
    const action = matchShortcut(event, $bindings)
    if (action) {
      // a modifier-less custom binding must not hijack typing in a field.
      if (!comboHasModifier($bindings[action]) && isEditableTarget(event.target)) {
        return
      }
      event.preventDefault()
      dispatchAction(action)
    }
  }

  // isEditableTarget reports whether the event originated in a text field, so
  // plain-key shortcuts do not fire while the user is typing.
  function isEditableTarget(target: EventTarget | null): boolean {
    const el = target as HTMLElement | null
    if (!el) {
      return false
    }
    const tag = el.tagName
    return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || el.isContentEditable
  }

  // pane resize. clamps keep each column usable; widths persist on release.
  function clamp(value: number, min: number, max: number): number {
    return Math.min(Math.max(value, min), max)
  }

  function resizeSidebar(event: CustomEvent<number>): void {
    dragging = true
    sidebarW = clamp(sidebarW + event.detail, 180, 480)
  }

  function resizeList(event: CustomEvent<number>): void {
    dragging = true
    listW = clamp(listW + event.detail, 260, 720)
  }

  function commitPanes(): void {
    dragging = false
    setPaneWidths(sidebarW, listW)
  }
</script>

<svelte:window on:keydown={onKeydown} on:contextmenu={onContextMenu} />

<div class="shell">
  <div class="columns" style={`grid-template-columns: ${sidebarW}px 0 ${listW}px 0 1fr`}>
    <Sidebar
      on:compose={startCompose}
      on:sync={runSync}
      on:addMailbox={addMailbox}
    />
    <Resizer disabled={locked} label="Resize sidebar" on:resize={resizeSidebar} on:end={commitPanes} />
    <MessageList />
    <Resizer disabled={locked} label="Resize message list" on:resize={resizeList} on:end={commitPanes} />
    <MessageDetail />
  </div>

  <StatusBar />
</div>

{#if $composeSessions.length > 0}
  <div class="compose-layer">
    {#each $composeSessions as session (session.id)}
      <Compose {session} />
    {/each}
  </div>
{/if}

<!-- settings and the wizard are code-split: their js/css load only when opened,
     so they cost nothing at startup. compose stays eager (used constantly). -->
{#if settingsOpen}
  {#await import('./components/settings/SettingsPanel.svelte') then m}
    <svelte:component
      this={m.default}
      on:close={() => (settingsOpen = false)}
      on:rerunOnboarding={rerunOnboarding}
    />
  {/await}
{/if}

<!-- onboarding is code-split too: shown only on first run or when re-run. -->
{#if onboardingOpen}
  {#await import('./components/onboarding/Onboarding.svelte') then m}
    <svelte:component this={m.default} on:finish={finishOnboarding} on:added={onboardingAddedMailbox} />
  {/await}
{/if}

{#if wizardOpen}
  {#await import('./components/wizard/AddMailboxWizard.svelte') then m}
    <svelte:component this={m.default} on:close={() => (wizardOpen = false)} on:added={onMailboxAdded} />
  {/await}
{/if}

<Toasts />
<ContextMenu />

<style>
  .shell {
    display: grid;
    grid-template-rows: 1fr auto;
    /* divide by the interface scale: css zoom enlarges everything but does not
       shrink vh/vw, so without this a zoomed shell overflows and its bottom row
       (the status bar) gets clipped. --ui-scale is set in theme.ts (default 1). */
    height: calc(100vh / var(--ui-scale, 1));
    width: calc(100vw / var(--ui-scale, 1));
    overflow: hidden;
  }

  /* the two zero-width tracks hold the resizer handles, which overhang via
     negative margins so they sit on the column borders without taking space. the
     single row is pinned to the shell height with a 0 minimum so each column's
     own scroll area can shrink and scroll instead of stretching the grid. */
  .columns {
    display: grid;
    grid-template-rows: minmax(0, 1fr);
    min-height: 0;
    overflow: hidden;
  }

  /* the dock spans the viewport width so extra panes wrap upward (wrap-reverse,
     bottom-anchored) instead of cascading off the left edge. it is click-through
     in the empty gaps; each pane re-enables pointer events. */
  .compose-layer {
    position: fixed;
    bottom: var(--space-5);
    right: var(--space-5);
    left: var(--space-5);
    display: flex;
    flex-wrap: wrap-reverse;
    justify-content: flex-end;
    gap: var(--space-4);
    align-items: flex-end;
    z-index: 90;
    pointer-events: none;
  }

  .compose-layer > :global(.compose) {
    pointer-events: auto;
  }
</style>
