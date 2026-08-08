<script lang="ts">
  // a recursive folder tree node. children are the folders whose parentId is this
  // folder's id, so the server-provided hierarchy and per-server delimiter are
  // respected without the frontend parsing paths. expansion state is local.
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
  } from '@tabler/icons-svelte'
  import SidebarRow from './SidebarRow.svelte'
  import Self from './FolderNode.svelte'
  import type { Folder } from '../../lib/types'
  import { selection, selectFolder } from '../../stores/selection'
  import { openContextMenu, type MenuEntry } from '../../stores/contextmenu'
  import { openCreateFolder, openRenameFolder, openDeleteFolder } from '../../stores/folderdialog'
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

  function onContext(event: MouseEvent): void {
    const entries: MenuEntry[] = []
    if (nestable) {
      entries.push({
        label: $t('folders.newSubfolder'),
        icon: IconFolderPlus,
        action: () => openCreateFolder(folder.accountId, folder),
      })
    }
    if (manageable) {
      if (entries.length > 0) {
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
    if (entries.length === 0) {
      return
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

<SidebarRow
  label={folder.name}
  count={folder.unreadCount}
  active={isActive}
  {depth}
  expandable={children.length > 0}
  {expanded}
  on:select={() => selectFolder(folder)}
  on:toggle={() => (expanded = !expanded)}
  on:contextmenu={(e) => onContext(e.detail)}
>
  <svelte:component this={Icon} size={15} stroke={1.6} />
</SidebarRow>

{#if expanded}
  {#each children as child (child.id)}
    <Self folder={child} {folders} depth={depth + 1} />
  {/each}
{/if}
