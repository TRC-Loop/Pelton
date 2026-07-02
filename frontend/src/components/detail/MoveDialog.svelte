<script lang="ts">
  // the "Move to folder" dialog. it lists the message's account folders and moves
  // it on the server, reusing the same undo machinery as archive (cmd+z moves it
  // back). the current folder is excluded; a search box filters long trees.
  import { fade, scale } from 'svelte/transition'
  import { IconX, IconFolder, IconSearch } from '@tabler/icons-svelte'
  import { moveTarget, closeMove } from '../../stores/move'
  import { listFolders, moveMessage } from '../../lib/api'
  import { removeFromList } from '../../stores/messages'
  import { recordArchived } from '../../stores/undoarchive'
  import { openMessageId } from '../../stores/selection'
  import { errorMessage, toastError, toastInfo } from '../../stores/toast'
  import { t } from '../../lib/i18n'
  import type { Folder } from '../../lib/types'

  let folders: Folder[] = []
  let query = ''
  let loadedFor = -1
  let busy = false

  // load the account's folders when the target changes.
  $: if ($moveTarget && $moveTarget.accountId !== loadedFor) {
    loadedFor = $moveTarget.accountId
    void load($moveTarget.accountId)
  }
  $: if (!$moveTarget) {
    loadedFor = -1
    query = ''
  }

  async function load(accountId: number): Promise<void> {
    try {
      folders = await listFolders(accountId)
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // candidate folders: everything except the message's current folder, filtered
  // by the search box.
  $: candidates = folders
    .filter((f) => !$moveTarget || f.id !== $moveTarget.folderId)
    .filter((f) => (query.trim() ? f.name.toLowerCase().includes(query.toLowerCase()) : true))

  async function move(folder: Folder): Promise<void> {
    const target = $moveTarget
    if (!target || busy) {
      return
    }
    busy = true
    // optimistic: drop it from the list right away.
    removeFromList(target.id)
    if ($openMessageId === target.id) {
      openMessageId.set(null)
    }
    closeMove()
    try {
      const undo = await moveMessage(target.id, folder.id)
      if (undo.messageId) {
        recordArchived(target, undo.messageId, undo.originalFolderId)
      }
      toastInfo($t('detail.moveDialog.moved').replace('{folder}', folder.name))
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      busy = false
    }
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      closeMove()
    }
  }
</script>

<svelte:window on:keydown={$moveTarget ? onKeydown : undefined} />

{#if $moveTarget}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="backdrop" transition:fade={{ duration: 120 }} on:click={closeMove}></div>
  <div class="dialog" role="dialog" aria-modal="true" aria-label={$t('detail.moveDialog.dialogLabel')} transition:scale={{ duration: 150, start: 0.94 }}>
    <header>
      <h2>{$t('detail.moveDialog.title')}</h2>
      <button type="button" class="close" aria-label={$t('detail.attachments.close')} on:click={closeMove}>
        <IconX size={16} stroke={1.8} />
      </button>
    </header>

    <div class="search">
      <IconSearch size={15} stroke={1.7} />
      <input type="search" placeholder={$t('detail.moveDialog.searchPlaceholder')} bind:value={query} />
    </div>

    <ul class="list">
      {#each candidates as f (f.id)}
        <li>
          <button type="button" class="folder" disabled={busy} on:click={() => move(f)}>
            <IconFolder size={16} stroke={1.6} />
            <span class="f-name">{f.name}</span>
          </button>
        </li>
      {/each}
      {#if candidates.length === 0}
        <li class="empty">{query ? $t('detail.moveDialog.noMatches') : $t('detail.moveDialog.noOthers')}</li>
      {/if}
    </ul>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 300;
    background: var(--scrim, rgba(0, 0, 0, 0.4));
    backdrop-filter: blur(2px);
  }

  .dialog {
    position: fixed;
    z-index: 301;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: min(400px, calc(100vw - 2 * var(--space-5)));
    max-height: 70vh;
    padding: var(--space-4);
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  h2 {
    margin: 0;
    font-size: var(--fz-title);
    font-weight: var(--fw-semibold);
    color: var(--text-primary);
  }
  .close {
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: var(--space-1);
    border-radius: var(--radius-control);
  }
  .close:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .search {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-tertiary);
  }
  .search input {
    flex: 1;
    border: none;
    background: transparent;
    color: var(--text-primary);
    font: inherit;
    outline: none;
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    overflow-y: auto;
    min-height: 0;
  }

  .folder {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    width: 100%;
    padding: var(--space-2) var(--space-3);
    border: none;
    background: transparent;
    color: var(--text-primary);
    cursor: pointer;
    text-align: left;
    border-radius: var(--radius-control);
    font-size: var(--fz-label);
  }
  .folder:hover {
    background: var(--surface-hover);
  }
  .folder :global(svg) {
    color: var(--text-tertiary);
    flex-shrink: 0;
  }
  .f-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .empty {
    padding: var(--space-3);
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }
</style>
