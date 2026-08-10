<script lang="ts">
  // a recursive folder tree node. children are the folders whose parentId is this
  // folder's id, so the server-provided hierarchy and per-server delimiter are
  // respected without the frontend parsing paths. expansion state is local.
  //
  // the whole node (row plus children) is one element so the parent's reorder
  // container sees exactly one child per folder. this node's own children are a
  // separate container, which is what keeps a drag inside its sibling group.
  import {
    IconInbox,
    IconSend,
    IconFile,
    IconTrash,
    IconAlertTriangle,
    IconArchive,
    IconFolder,
    IconFolderPlus,
    IconPencil,
    IconPin,
    IconPinnedOff,
    IconTags,
  } from '@tabler/icons-svelte'
  import SidebarRow from './SidebarRow.svelte'
  import Self from './FolderNode.svelte'
  import type { Folder } from '../../lib/types'
  import { reorder, type ReorderDetail } from '../../lib/reorder'
  import { selection, selectFolder } from '../../stores/selection'
  import { openContextMenu, type MenuEntry } from '../../stores/contextmenu'
  import {
    openCreateFolder,
    openRenameFolder,
    openDeleteFolder,
    openEmptyTrash,
    openFolderRole,
  } from '../../stores/folderdialog'
  import { startFolderDrag, endFolderDrag } from '../../stores/sidebardrag'
  import { refreshSidebar } from '../../stores/accounts'
  import { reorderFolders, setFolderPinned } from '../../lib/api'
  import { toastError, errorMessage } from '../../stores/toast'
  import { t } from '../../lib/i18n'

  export let folder: Folder
  export let folders: Folder[]
  export let depth: number = 0

  let expanded = true

  $: children = folders.filter((f) => f.parentId === folder.id)
  $: isActive = $selection.kind === 'folder' && $selection.folderId === folder.id
  // special mailboxes are refused by the backend (renaming Sent out from under
  // the app would leave sent mail nowhere to go), so the ui does not offer it.
  $: manageable = folder.role === 'normal'
  // nesting needs a hierarchy delimiter; a flat server has none.
  $: nestable = folder.delimiter !== ''
  // emptying is only offered on the trash, and only when there is something in
  // it, so the menu never carries an action that would do nothing.
  $: emptiable = folder.role === 'trash' && folder.totalCount > 0

  async function onReorder(event: CustomEvent<ReorderDetail>): Promise<void> {
    try {
      await reorderFolders(event.detail.ids.map(Number))
    } catch (err) {
      toastError(errorMessage(err))
    }
    // reload either way, so the rows end up showing what was actually stored.
    await refreshSidebar()
  }

  async function togglePinned(): Promise<void> {
    try {
      await setFolderPinned(folder.id, !folder.pinned)
      await refreshSidebar()
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  function onContext(event: MouseEvent): void {
    const entries: MenuEntry[] = [
      {
        label: folder.pinned ? $t('folders.unpin') : $t('folders.pin'),
        icon: folder.pinned ? IconPinnedOff : IconPin,
        action: () => void togglePinned(),
      },
      // offered on every folder, including the inbox: a mailbox the server
      // flagged wrongly is exactly the one that needs correcting.
      {
        label: $t('folders.roleAction'),
        icon: IconTags,
        action: () => openFolderRole(folder),
      },
    ]
    if (emptiable) {
      entries.push('separator')
      entries.push({
        label: $t('folders.emptyTrash'),
        icon: IconTrash,
        danger: true,
        action: () => openEmptyTrash(folder),
      })
    }
    if (nestable) {
      entries.push('separator')
      entries.push({
        label: $t('folders.newSubfolder'),
        icon: IconFolderPlus,
        action: () => openCreateFolder(folder.accountId, folder),
      })
    }
    if (manageable) {
      if (!nestable) {
        entries.push('separator')
      }
      entries.push({
        label: $t('folders.rename'),
        icon: IconPencil,
        action: () => openRenameFolder(folder),
      })
      entries.push({
        label: $t('folders.delete'),
        icon: IconTrash,
        danger: true,
        action: () => openDeleteFolder(folder),
      })
    }
    openContextMenu(event.clientX, event.clientY, entries)
  }

  // pick an icon by role so special folders read at a glance.
  const roleIcons: Record<string, typeof IconFolder> = {
    inbox: IconInbox,
    sent: IconSend,
    drafts: IconFile,
    trash: IconTrash,
    junk: IconAlertTriangle,
    archive: IconArchive,
    normal: IconFolder,
  }
  $: Icon = roleIcons[folder.role] ?? IconFolder
</script>

<div class="node" data-reorder-id={folder.id}>
  <SidebarRow
    label={folder.name}
    count={folder.unreadCount}
    active={isActive}
    {depth}
    expandable={children.length > 0}
    {expanded}
    reorderable
    on:select={() => selectFolder(folder)}
    on:toggle={() => (expanded = !expanded)}
    on:contextmenu={(e) => onContext(e.detail)}
  >
    <svelte:component this={Icon} size={15} stroke={1.6} />
  </SidebarRow>

  {#if expanded && children.length > 0}
    <div
      use:reorder
      on:reorderstart={() => startFolderDrag(folder.accountId)}
      on:reorderend={endFolderDrag}
      on:reorder={onReorder}
    >
      {#each children as child (child.id)}
        <Self folder={child} {folders} depth={depth + 1} />
      {/each}
    </div>
  {/if}
</div>
