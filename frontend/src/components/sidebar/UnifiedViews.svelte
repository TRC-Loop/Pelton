<script lang="ts">
  // the unified cross-account views at the top of the sidebar. unified inbox is
  // the default selection. flagged shows a total (flagged are not "unread"), the
  // others show unread counts.
  import {
    IconInbox,
    IconFlag,
    IconSend,
    IconFile,
    IconArchive,
    IconAlertTriangle,
    IconTrash,
    IconFolder,
  } from '@tabler/icons-svelte'
  import SidebarRow from './SidebarRow.svelte'
  import type { UnifiedView, ViewKey } from '../../lib/types'
  import { selection, selectView } from '../../stores/selection'
  import { prefs } from '../../stores/prefs'

  export let views: UnifiedView[]

  const viewIcons: Record<string, typeof IconFolder> = {
    inbox: IconInbox,
    flagged: IconFlag,
    sent: IconSend,
    drafts: IconFile,
    archive: IconArchive,
    junk: IconAlertTriangle,
    trash: IconTrash,
  }

  // flagged and drafts read better as totals; the rest show unread. the flagged
  // count (and its bold styling, which follows count > 0) can be hidden while the
  // entry stays, via the setting.
  function badgeCount(view: UnifiedView, showFlagged: boolean): number {
    if (view.key === 'flagged') {
      return showFlagged ? view.totalCount : 0
    }
    return view.key === 'drafts' ? view.totalCount : view.unreadCount
  }

  // the cast lives in script; inline ts casts in markup confuse the parser.
  function choose(view: UnifiedView): void {
    selectView(view.key as ViewKey, view.label)
  }
</script>

<nav class="views" aria-label="Unified views">
  <header class="group-head">Unified</header>
  {#each views as view (view.key)}
    <SidebarRow
      label={view.label}
      count={badgeCount(view, $prefs.showFlaggedCount)}
      active={$selection.kind === 'view' && $selection.view === view.key}
      on:select={() => choose(view)}
    >
      <svelte:component this={viewIcons[view.key] ?? IconFolder} size={15} stroke={1.6} />
    </SidebarRow>
  {/each}
</nav>

<style>
  .views {
    display: flex;
    flex-direction: column;
    gap: 1px;
  }

  .group-head {
    padding: var(--space-2) var(--space-3);
    font-size: var(--fz-label);
    font-weight: var(--fw-semibold);
    color: var(--text-tertiary);
  }
</style>
