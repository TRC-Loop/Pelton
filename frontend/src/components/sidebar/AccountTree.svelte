<script lang="ts">
  // one account's section in the sidebar: a header with the address and the full
  // folder tree below it. the header chevron collapses the whole account. root
  // folders are those without a parent; the rest nest via FolderNode recursion.
  import { IconChevronRight, IconFolderPlus } from '@tabler/icons-svelte'
  import FolderNode from './FolderNode.svelte'
  import DragGrip from './DragGrip.svelte'
  import { prefs } from '../../stores/prefs'
  import { collapsedAccounts, toggleAccountCollapsed } from '../../stores/sidebarstate'
  import { openCreateFolder } from '../../stores/folderdialog'
  import { folderDragAccount, startFolderDrag, endFolderDrag } from '../../stores/sidebardrag'
  import { reorder, type ReorderDetail } from '../../lib/reorder'
  import { refreshSidebar } from '../../stores/accounts'
  import { reorderFolders } from '../../lib/api'
  import { toastError, errorMessage } from '../../stores/toast'
  import type { Account, Folder } from '../../lib/types'
  import { t } from '../../lib/i18n'

  export let account: Account
  export let folders: Folder[]

  // the account section is expanded by default; the collapsed set (persisted) is
  // the source of truth so the choice survives restarts.
  $: expanded = !$collapsedAccounts.has(account.id)

  // roots have no parent. loose equality catches both null and undefined.
  $: roots = folders.filter((f) => f.parentId == null)
  // a folder never moves between accounts, so a drag elsewhere dims this section
  // to show where the drop can land.
  $: dimmed = $folderDragAccount !== null && $folderDragAccount !== account.id

  async function onReorder(event: CustomEvent<ReorderDetail>): Promise<void> {
    try {
      await reorderFolders(event.detail.ids.map(Number))
    } catch (err) {
      toastError(errorMessage(err))
    }
    // reload either way, so the rows end up showing what was actually stored.
    await refreshSidebar()
  }

  // the header shows the display name, or the email when the user prefers it (or
  // when there is no display name to show). Local Folders has no real address,
  // so it is named rather than addressed.
  $: label = account.local
    ? $t('sidebar.localFolders')
    : $prefs.showAccountEmail
      ? account.email
      : account.displayName || account.email
</script>

<section class="account" class:dimmed data-reorder-id={account.id}>
  <div class="account-row">
    <button
      type="button"
      class="account-head"
      class:open={expanded}
      title={account.local ? $t('sidebar.localFoldersHint') : account.email}
      aria-expanded={expanded}
      on:click={() => toggleAccountCollapsed(account.id)}
    >
      <span class="account-caret" aria-hidden="true"><IconChevronRight size={13} stroke={1.9} /></span>
      <span class="account-name">{label}</span>
    </button>
    <button
      type="button"
      class="new-folder"
      title={$t('folders.newFolder')}
      aria-label={$t('folders.newFolder')}
      on:click={() => openCreateFolder(account.id)}
    >
      <IconFolderPlus size={14} stroke={1.7} />
    </button>
    <DragGrip />
  </div>
  {#if expanded}
    {#if roots.length === 0}
      <p class="empty-note">{$t('sidebar.account.noFolders')}</p>
    {:else}
      <div
        use:reorder
        on:reorderstart={() => startFolderDrag(account.id)}
        on:reorderend={endFolderDrag}
        on:reorder={onReorder}
      >
        {#each roots as folder (folder.id)}
          <FolderNode {folder} {folders} depth={0} />
        {/each}
      </div>
    {/if}
  {/if}
</section>

<style>
  .account {
    margin-top: var(--space-4);
  }

  /* a folder drag can only land among its own siblings, so every other account
     fades back while one is in progress. pointer-events stay on: dnd needs to
     keep tracking the pointer, it just cannot drop here. */
  .account.dimmed {
    opacity: 0.4;
    transition: opacity 0.12s ease;
  }

  /* the header and its new-folder button share a row; the button only appears
     on hover so the sidebar stays quiet at rest. */
  .account-row {
    display: flex;
    align-items: center;
    border-radius: var(--radius-control);
  }

  .account-head {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    flex: 1;
    min-width: 0;
    padding: var(--space-2) var(--space-3);
    border: none;
    background: transparent;
    cursor: var(--cursor-action);
    text-align: left;
    border-radius: var(--radius-control);
  }

  .account-row:hover {
    background: var(--surface-hover);
  }

  .new-folder {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    padding: var(--space-1);
    margin-right: var(--space-2);
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    border-radius: var(--radius-control);
    cursor: var(--cursor-action);
    opacity: 0;
  }

  .account-row:hover .new-folder,
  .new-folder:focus-visible {
    opacity: 1;
  }

  .new-folder:hover {
    color: var(--text-primary);
    background: var(--surface-sunken);
  }

  .account-caret {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--text-tertiary);
    flex-shrink: 0;
  }

  .account-caret :global(svg) {
    transition: transform 0.12s ease;
  }

  .account-head.open .account-caret :global(svg) {
    transform: rotate(90deg);
  }

  .account-name {
    font-size: var(--fz-label);
    font-weight: var(--fw-semibold);
    color: var(--text-tertiary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
  }

  .empty-note {
    margin: 0;
    padding: var(--space-2) var(--space-3) var(--space-2) var(--space-5);
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }
</style>
