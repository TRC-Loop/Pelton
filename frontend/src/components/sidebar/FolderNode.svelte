<script lang="ts">
  // a recursive folder tree node. children are the folders whose parentId is this
  // folder's id, so the server-provided hierarchy and per-server delimiter are
  // respected without the frontend parsing paths. expansion state is local.
  import { IconInbox, IconSend, IconFile, IconTrash, IconAlertTriangle, IconArchive, IconFolder } from '@tabler/icons-svelte'
  import SidebarRow from './SidebarRow.svelte'
  import Self from './FolderNode.svelte'
  import type { Folder } from '../../lib/types'
  import { selection, selectFolder } from '../../stores/selection'

  export let folder: Folder
  export let folders: Folder[]
  export let depth: number = 0

  let expanded = true

  $: children = folders.filter((f) => f.parentId === folder.id)
  $: isActive = $selection.kind === 'folder' && $selection.folderId === folder.id

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
>
  <svelte:component this={Icon} size={15} stroke={1.6} />
</SidebarRow>

{#if expanded}
  {#each children as child (child.id)}
    <Self folder={child} {folders} depth={depth + 1} />
  {/each}
{/if}
