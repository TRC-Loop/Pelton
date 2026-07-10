<script lang="ts">
  // column 2: the message list. it loads the page for the current selection,
  // runs fts5 search, and owns keyboard navigation (up/down to move, enter to
  // open). opening a message marks it seen optimistically. when multi-select is
  // enabled, cmd/ctrl-click toggles rows and shift-click selects a range, with a
  // selection toolbar for bulk mark/flag/delete. loading, empty and error states
  // are all explicit so the pane is never blank.
  import { tick, onMount } from 'svelte'
  import MessageRow from './MessageRow.svelte'
  import SearchBar from './SearchBar.svelte'
  import Spinner from '../common/Spinner.svelte'
  import EmptyState from '../common/EmptyState.svelte'
  import ErrorState from '../common/ErrorState.svelte'
  import {
    IconMail,
    IconArrowBackUp,
    IconArrowForwardUp,
    IconMailOpened,
    IconMailFilled,
    IconFlag,
    IconFlagFilled,
    IconTrash,
    IconX,
    IconClockPause,
    IconDownload,
    IconDownloadOff,
    IconFolderSymlink,
  } from '@tabler/icons-svelte'
  import { selection, searchQuery, openMessageId, openMessage } from '../../stores/selection'
  import {
    messageList,
    loadList,
    loadMore,
    runSearch,
    patchInList,
    removeFromList,
    restoreToList,
    emptyFilter,
    filterActive,
    type SearchFilter,
  } from '../../stores/messages'
  import {
    selectedIds,
    clearSelection,
    toggleSelect,
    selectRange,
  } from '../../stores/listselect'
  import { prefs } from '../../stores/prefs'
  import {
    setSeen,
    setFlagged,
    deleteMessage,
    getMessage,
    setFlagColor,
    downloadMessageOffline,
    removeOffline,
    archiveMessage,
  } from '../../lib/api'
  import { openReply, openForward } from '../../stores/compose'
  import { openSnooze } from '../../stores/snooze'
  import { openMove } from '../../stores/move'
  import { recordDeleted } from '../../stores/undodelete'
  import { recordArchived } from '../../stores/undoarchive'
  import { openContextMenu, type MenuEntry } from '../../stores/contextmenu'
  import { errorMessage, toastError, toastSuccess } from '../../stores/toast'
  import type { Selection, MessageSummary, SwipeAction, EditorMode } from '../../lib/types'
  import { t } from '../../lib/i18n'

  let listEl: HTMLDivElement
  let activeIndex = -1

  // virtualization. the list can hold thousands of rows, so we only render the
  // window around the viewport plus a little overscan, padding the rest with
  // spacers so the scrollbar stays accurate. all rows share a template, so a
  // single measured row height is enough; we fall back to a per-template estimate
  // until the first real row is measured.
  const OVERSCAN = 8
  const ROW_ESTIMATES: Record<string, number> = {
    relaxed: 76,
    comfortable: 62,
    compact: 50,
    single: 34,
  }
  let scrollTop = 0
  let viewportHeight = 600
  let rowHeight = 0
  $: estRowHeight = rowHeight || ROW_ESTIMATES[$prefs.rowTemplate] || 64

  // re-measure when anything that changes a row's height changes.
  $: rowMetricKey = `${$prefs.rowTemplate}|${$prefs.density}|${$prefs.previewLines}|${$prefs.rowShowSnippet}|${$prefs.rowShowAvatar}|${$prefs.showDateTime}`
  let lastMetricKey = ''
  $: if (rowMetricKey !== lastMetricKey) {
    lastMetricKey = rowMetricKey
    rowHeight = 0
    void measureRow()
  }

  // measureRow reads the first rendered row's height so the window math matches
  // the real layout.
  async function measureRow(): Promise<void> {
    await tick()
    const node = listEl?.querySelector('[role="option"]')
    if (node instanceof HTMLElement && node.offsetHeight > 0) {
      rowHeight = node.offsetHeight
    }
  }

  onMount(() => {
    if (listEl) {
      viewportHeight = listEl.clientHeight || viewportHeight
    }
  })

  // selectionKey identifies a selection so we reload only when it actually
  // changes, not on unrelated store updates.
  function selectionKey(sel: Selection): string {
    return sel.kind === 'view' ? `view:${sel.view}` : `folder:${sel.folderId}`
  }

  let lastKey = ''
  $: if ($selection && selectionKey($selection) !== lastKey) {
    lastKey = selectionKey($selection)
    activeIndex = -1
    clearSelection()
    void loadList($selection)
  }

  $: items = $messageList.data?.items ?? []
  $: hasMore = !($messageList.data?.searching ?? false) && items.length < ($messageList.data?.total ?? 0)

  // measure once rows first appear and the height is still unknown.
  $: if (items.length > 0 && rowHeight === 0) {
    void measureRow()
  }

  // the rendered window: a contiguous slice around the viewport, with spacer
  // heights above and below so the scroll range stays correct.
  $: startIndex = Math.max(0, Math.floor(scrollTop / estRowHeight) - OVERSCAN)
  $: visibleCount = Math.ceil(viewportHeight / estRowHeight) + OVERSCAN * 2
  $: endIndex = Math.min(items.length, startIndex + visibleCount)
  $: windowItems = items.slice(startIndex, endIndex)
  $: topPad = startIndex * estRowHeight
  $: bottomPad = Math.max(0, (items.length - endIndex) * estRowHeight)

  // the live multi-selection, intersected with what is still loaded so bulk
  // actions never act on rows that have scrolled out of the data.
  $: selectedItems = items.filter((m) => $selectedIds.has(m.id))
  $: selectionCount = selectedItems.length

  // keep the keyboard-nav highlight in sync with whichever message is open,
  // however it got opened (click, Enter, or the in-app vim motions), so a
  // stale arrow-key position never leaves two rows looking selected at once.
  $: if ($openMessageId !== null) {
    const openIdx = items.findIndex((m) => m.id === $openMessageId)
    if (openIdx !== -1) {
      activeIndex = openIdx
    }
  }

  // search handling. the list shows ranked results when there is a query or an
  // active date filter, and the selection's normal list otherwise.
  let searchFilter: SearchFilter = emptyFilter

  function applySearch(query: string, filter: SearchFilter): void {
    activeIndex = -1
    clearSelection()
    if (query === '' && !filterActive(filter)) {
      void loadList($selection)
    } else {
      void runSearch(query, filter)
    }
  }

  function onSearch(event: CustomEvent<string>): void {
    searchQuery.set(event.detail)
    applySearch(event.detail.trim(), searchFilter)
  }

  function onFilter(event: CustomEvent<SearchFilter>): void {
    searchFilter = event.detail
    applySearch($searchQuery.trim(), searchFilter)
  }

  // open marks the message seen if needed and shows it in the detail pane. a
  // plain open also clears any multi-selection.
  async function open(index: number): Promise<void> {
    const item = items[index]
    if (!item) {
      return
    }
    activeIndex = index
    clearSelection()
    openMessage(item.id)
    if (!item.seen) {
      patchInList(item.id, { seen: true })
      try {
        await setSeen(item.id, true)
      } catch (err) {
        toastError(errorMessage(err))
      }
    }
  }

  // onRowClick routes a click to either multi-selection (cmd/ctrl or shift, when
  // enabled) or opening the message.
  function onRowClick(event: MouseEvent, index: number): void {
    const item = items[index]
    if (!item) {
      return
    }
    if ($prefs.multiSelectEnabled && (event.metaKey || event.ctrlKey)) {
      toggleSelect(item.id)
      activeIndex = index
      return
    }
    if ($prefs.multiSelectEnabled && event.shiftKey) {
      selectRange(items.map((m) => m.id), item.id)
      activeIndex = index
      return
    }
    void open(index)
  }

  // keyboard navigation over the rows.
  async function onKeydown(event: KeyboardEvent): Promise<void> {
    if (event.key === 'Escape' && selectionCount > 0) {
      clearSelection()
      return
    }
    if (items.length === 0) {
      return
    }
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      activeIndex = Math.min(activeIndex + 1, items.length - 1)
      await scrollActiveIntoView()
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      activeIndex = Math.max(activeIndex - 1, 0)
      await scrollActiveIntoView()
    } else if (event.key === 'Enter' && activeIndex >= 0) {
      event.preventDefault()
      await open(activeIndex)
    }
  }

  // scrollActiveIntoView keeps the highlighted row visible during arrow nav.
  // because rows are virtualized, it scrolls by computed offset rather than
  // asking a (possibly unrendered) node to scroll itself into view.
  async function scrollActiveIntoView(): Promise<void> {
    if (!listEl || activeIndex < 0) {
      return
    }
    const top = activeIndex * estRowHeight
    const bottom = top + estRowHeight
    if (top < listEl.scrollTop) {
      listEl.scrollTop = top
    } else if (bottom > listEl.scrollTop + listEl.clientHeight) {
      listEl.scrollTop = bottom - listEl.clientHeight
    }
    scrollTop = listEl.scrollTop
    await tick()
  }

  // toggleSeen / toggleFlag / remove act on a single row from the context menu,
  // updating the list optimistically and surfacing any backend error.
  async function toggleSeen(item: MessageSummary): Promise<void> {
    patchInList(item.id, { seen: !item.seen })
    try {
      await setSeen(item.id, !item.seen)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function toggleFlag(item: MessageSummary): Promise<void> {
    patchInList(item.id, { flagged: !item.flagged })
    try {
      await setFlagged(item.id, !item.flagged)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function remove(item: MessageSummary): Promise<void> {
    try {
      await deleteMessage(item.id)
      recordDeleted(item)
      removeFromList(item.id)
      if ($openMessageId === item.id) {
        openMessageId.set(null)
      }
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // bulk actions operate on the whole multi-selection, then clear it.
  async function bulkSetSeen(seen: boolean): Promise<void> {
    const targets = selectedItems
    for (const item of targets) {
      patchInList(item.id, { seen })
    }
    clearSelection()
    await Promise.all(
      targets.map((item) =>
        setSeen(item.id, seen).catch((err) => toastError(errorMessage(err))),
      ),
    )
  }

  async function bulkSetFlagged(flagged: boolean): Promise<void> {
    const targets = selectedItems
    for (const item of targets) {
      patchInList(item.id, { flagged })
    }
    clearSelection()
    await Promise.all(
      targets.map((item) =>
        setFlagged(item.id, flagged).catch((err) => toastError(errorMessage(err))),
      ),
    )
  }

  async function bulkDelete(): Promise<void> {
    const targets = selectedItems
    clearSelection()
    for (const item of targets) {
      try {
        await deleteMessage(item.id)
        recordDeleted(item)
        removeFromList(item.id)
        if ($openMessageId === item.id) {
          openMessageId.set(null)
        }
      } catch (err) {
        toastError(errorMessage(err))
      }
    }
  }

  // reply/forward need the full message (for quoting), so load it first.
  $: editorMode = $prefs.defaultEditorMode as EditorMode

  async function replyTo(item: MessageSummary, all: boolean): Promise<void> {
    try {
      const detail = await getMessage(item.id)
      openReply(detail, editorMode, all)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  async function forward(item: MessageSummary): Promise<void> {
    try {
      const detail = await getMessage(item.id)
      openForward(detail, editorMode)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // setColor applies a flag color to a row (0 clears), optimistically.
  async function setColor(item: MessageSummary, color: number): Promise<void> {
    patchInList(item.id, { flagColor: color })
    try {
      await setFlagColor(item.id, color)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // archive moves a row to the account's Archive folder, removing it from the
  // current view optimistically.
  async function archive(item: MessageSummary): Promise<void> {
    removeFromList(item.id)
    if ($openMessageId === item.id) {
      openMessageId.set(null)
    }
    try {
      const undo = await archiveMessage(item.id)
      if (undo.messageId) {
        recordArchived(item, undo.messageId, undo.originalFolderId)
      }
    } catch (err) {
      toastError(errorMessage(err))
      // the move failed; bring the row back so nothing is silently lost.
      restoreToList(item)
    }
  }

  // toggleOffline pins or unpins a row for offline availability.
  async function toggleOffline(item: MessageSummary): Promise<void> {
    const next = !item.offline
    patchInList(item.id, { offline: next })
    try {
      if (next) {
        await downloadMessageOffline(item.id)
        toastSuccess($t('messageList.toast.savedOffline'))
      } else {
        await removeOffline(item.id)
      }
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // performSwipe runs the configured action for a swipe direction on a row.
  function performSwipe(item: MessageSummary, dir: 'left' | 'right'): void {
    const action = (dir === 'left' ? $prefs.swipeLeftAction : $prefs.swipeRightAction) as SwipeAction
    switch (action) {
      case 'delete':
        void remove(item)
        break
      case 'read':
      case 'unread':
        void toggleSeen(item)
        break
      case 'flag':
        void toggleFlag(item)
        break
      case 'archive':
        void archive(item)
        break
      case 'snooze':
        openSnooze(item.id, item.subject)
        break
      case 'none':
      default:
        break
    }
  }

  // onContext builds and opens the right-click menu. when the row is part of a
  // multi-selection of more than one, the menu offers bulk actions instead.
  function onContext(event: MouseEvent, item: MessageSummary): void {
    event.preventDefault()
    if (selectionCount > 1 && $selectedIds.has(item.id)) {
      const anyUnread = selectedItems.some((m) => !m.seen)
      const anyUnflagged = selectedItems.some((m) => !m.flagged)
      const entries: MenuEntry[] = [
        anyUnread
          ? { label: $t('messageList.bulk.markReadCount').replace('{n}', String(selectionCount)), icon: IconMailOpened, action: () => void bulkSetSeen(true) }
          : { label: $t('messageList.bulk.markUnreadCount').replace('{n}', String(selectionCount)), icon: IconMailFilled, action: () => void bulkSetSeen(false) },
        anyUnflagged
          ? { label: $t('messageList.bulk.flagCount').replace('{n}', String(selectionCount)), icon: IconFlagFilled, action: () => void bulkSetFlagged(true) }
          : { label: $t('messageList.bulk.unflagCount').replace('{n}', String(selectionCount)), icon: IconFlag, action: () => void bulkSetFlagged(false) },
        'separator',
        { label: $t('messageList.bulk.deleteCount').replace('{n}', String(selectionCount)), icon: IconTrash, danger: true, action: () => void bulkDelete() },
      ]
      openContextMenu(event.clientX, event.clientY, entries)
      return
    }
    const entries: MenuEntry[] = [
      { label: $t('messageList.menu.open'), icon: IconMail, action: () => open(items.indexOf(item)) },
      { label: $t('action.reply'), icon: IconArrowBackUp, action: () => void replyTo(item, false) },
      { label: $t('shortcut.replyAll'), icon: IconArrowBackUp, action: () => void replyTo(item, true) },
      { label: $t('action.forward'), icon: IconArrowForwardUp, action: () => void forward(item) },
      'separator',
      item.seen
        ? { label: $t('shortcut.markUnread'), icon: IconMailFilled, action: () => void toggleSeen(item) }
        : { label: $t('shortcut.markRead'), icon: IconMailOpened, action: () => void toggleSeen(item) },
      item.flagged
        ? { label: $t('messageList.unflag'), icon: IconFlag, action: () => void toggleFlag(item) }
        : { label: $t('messageList.flag'), icon: IconFlagFilled, action: () => void toggleFlag(item) },
      { kind: 'colors', current: item.flagColor, onPick: (color) => void setColor(item, color) },
      'separator',
      { label: $t('messageList.menu.snooze'), icon: IconClockPause, action: () => openSnooze(item.id, item.subject) },
      { label: $t('messageList.menu.moveTo'), icon: IconFolderSymlink, action: () => openMove(item) },
      item.offline
        ? { label: $t('messageList.menu.removeOffline'), icon: IconDownloadOff, action: () => void toggleOffline(item) }
        : { label: $t('shortcut.downloadOffline'), icon: IconDownload, action: () => void toggleOffline(item) },
      'separator',
      { label: $t('action.delete'), icon: IconTrash, danger: true, action: () => void remove(item) },
    ]
    openContextMenu(event.clientX, event.clientY, entries)
  }

  // onScroll updates the virtualization window and pages in more rows near the
  // bottom.
  function onScroll(): void {
    scrollTop = listEl.scrollTop
    viewportHeight = listEl.clientHeight
    if (!hasMore || $messageList.status === 'loading') {
      return
    }
    const nearBottom = listEl.scrollTop + listEl.clientHeight >= listEl.scrollHeight - 200
    if (nearBottom) {
      void loadMore()
    }
  }

  // keep the viewport height current when the window resizes.
  function onResize(): void {
    if (listEl) {
      viewportHeight = listEl.clientHeight
    }
  }
</script>

<svelte:window on:resize={onResize} />

<section class="list-col">
  <div class="header">
    <SearchBar value={$searchQuery} on:search={onSearch} on:filter={onFilter} />
  </div>

  {#if selectionCount > 0}
    <div class="select-bar">
      <button type="button" class="clear" aria-label={$t('messageList.clearSelection')} on:click={clearSelection}>
        <IconX size={15} stroke={1.9} />
      </button>
      {#if $prefs.showSelectedCount}
        <span class="sel-count">{selectionCount} {$t('messageList.selectedSuffix')}</span>
      {/if}
      <span class="sel-spacer"></span>
      {#if selectedItems.some((m) => !m.seen)}
        <button type="button" class="act" title={$t('shortcut.markRead')} on:click={() => bulkSetSeen(true)}>
          <IconMailOpened size={16} stroke={1.7} />
        </button>
      {:else}
        <button type="button" class="act" title={$t('shortcut.markUnread')} on:click={() => bulkSetSeen(false)}>
          <IconMailFilled size={16} stroke={1.7} />
        </button>
      {/if}
      {#if selectedItems.some((m) => !m.flagged)}
        <button type="button" class="act" title={$t('messageList.flag')} on:click={() => bulkSetFlagged(true)}>
          <IconFlagFilled size={16} stroke={1.7} />
        </button>
      {:else}
        <button type="button" class="act" title={$t('messageList.unflag')} on:click={() => bulkSetFlagged(false)}>
          <IconFlag size={16} stroke={1.7} />
        </button>
      {/if}
      <button type="button" class="act danger" title={$t('action.delete')} on:click={bulkDelete}>
        <IconTrash size={16} stroke={1.7} />
      </button>
    </div>
  {:else}
    <div class="meta-bar">
      <span class="title">{$selection.label}</span>
      {#if $messageList.data}
        <span class="count">
          {#if $messageList.data.searching}
            {items.length} {items.length === 1 ? $t('messageList.result') : $t('messageList.results')}
          {:else}
            {$messageList.data.total} {$messageList.data.total === 1 ? $t('messageList.message') : $t('messageList.messages')}
          {/if}
        </span>
      {/if}
    </div>
  {/if}

  <div
    class="rows"
    role="listbox"
    tabindex="0"
    aria-label={$t('messageList.ariaLabel')}
    aria-multiselectable={$prefs.multiSelectEnabled}
    aria-activedescendant={activeIndex >= 0 ? `msg-${items[activeIndex]?.id}` : undefined}
    bind:this={listEl}
    on:keydown={onKeydown}
    on:scroll={onScroll}
  >
    {#if $messageList.status === 'loading' && items.length === 0}
      <Spinner label={$t('messageList.loading')} />
    {:else if $messageList.status === 'error'}
      <ErrorState message={$messageList.error} onRetry={() => loadList($selection)} />
    {:else if items.length === 0}
      <EmptyState
        title={$searchQuery ? $t('messageList.empty.noMatch') : $t('messageList.empty.noMessages')}
        detail={$searchQuery ? $t('messageList.empty.tryDifferentSearch') : $t('messageList.empty.viewEmpty')}
      >
        <IconMail size={28} stroke={1.4} />
      </EmptyState>
    {:else}
      {#if topPad > 0}
        <div class="spacer" style={`height:${topPad}px`} aria-hidden="true"></div>
      {/if}
      {#each windowItems as item, i (item.id)}
        {@const index = startIndex + i}
        <div id={`msg-${item.id}`}>
          <MessageRow
            message={item}
            selected={item.id === $openMessageId || index === activeIndex}
            checked={$selectedIds.has(item.id)}
            on:click={(e) => onRowClick(e, index)}
            on:contextmenu={(e) => onContext(e, item)}
            on:swipe={(e) => performSwipe(item, e.detail)}
          />
        </div>
      {/each}
      {#if bottomPad > 0}
        <div class="spacer" style={`height:${bottomPad}px`} aria-hidden="true"></div>
      {/if}
      {#if $messageList.status === 'loading'}
        <Spinner label={$t('messageList.loadingMore')} inline />
      {/if}
    {/if}
  </div>
</section>

<style>
  .list-col {
    display: grid;
    grid-template-rows: auto auto 1fr;
    height: 100%;
    background: var(--surface-raised);
    border-right: var(--hairline) solid var(--border-default);
    min-width: 0;
  }

  .header {
    padding: var(--space-3);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .meta-bar {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3);
    padding: var(--space-2) var(--row-pad-x);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .title {
    font-size: var(--fz-label);
    font-weight: var(--fw-semibold);
    color: var(--text-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .count {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    flex-shrink: 0;
  }

  /* the selection toolbar replaces the meta bar while rows are selected. */
  .select-bar {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-1) var(--space-2);
    border-bottom: var(--hairline) solid var(--border-subtle);
    background: var(--selection-bg);
  }

  .sel-count {
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--text-primary);
  }

  .sel-spacer {
    flex: 1;
  }

  .clear,
  .act {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    padding: var(--space-2);
    border-radius: var(--radius-control);
  }

  .clear:hover,
  .act:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .act.danger:hover {
    color: var(--danger);
  }

  .rows {
    /* min-height:0 lets this 1fr grid track shrink below its content so it
       actually scrolls instead of growing the column past the viewport. */
    min-height: 0;
    overflow-y: auto;
    outline: none;
  }

  /* virtualization spacers stand in for the rows outside the rendered window. */
  .spacer {
    flex-shrink: 0;
  }
</style>
