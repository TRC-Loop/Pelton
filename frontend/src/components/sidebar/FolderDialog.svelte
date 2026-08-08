<script lang="ts">
  // the create/rename/delete mailbox dialog (#132). one component for all three
  // because they share a shape: a folder, a confirmation and a server round trip
  // that can fail. delete is destructive on the server, so it asks for the
  // folder's name to be typed rather than offering a plain confirm button.
  import { fade, scale } from 'svelte/transition'
  import { IconX, IconAlertTriangle } from '@tabler/icons-svelte'
  import { folderDialog, closeFolderDialog } from '../../stores/folderdialog'
  import { createFolder, renameFolder, deleteFolder } from '../../lib/api'
  import { refreshSidebar } from '../../stores/accounts'
  import { selection, selectView } from '../../stores/selection'
  import { toastError, toastInfo, errorMessage } from '../../stores/toast'
  import { t } from '../../lib/i18n'

  let name = ''
  let confirmName = ''
  let busy = false
  // the request the input was seeded from, so reopening the dialog for a
  // different folder reseeds it but typing does not fight the reactive block.
  let seededFor: unknown = null

  $: request = $folderDialog
  $: if (request !== seededFor) {
    seededFor = request
    name = request?.mode === 'rename' ? (request.folder?.name ?? '') : ''
    confirmName = ''
    busy = false
  }

  $: title =
    request?.mode === 'create'
      ? request.folder
        ? $t('folders.createSubtitle').replace('{parent}', request.folder.name)
        : $t('folders.createTitle')
      : request?.mode === 'rename'
        ? $t('folders.renameTitle')
        : $t('folders.deleteTitle')

  // delete needs the exact folder name typed back; the others just need a name
  // that is non-empty and, for rename, actually different.
  $: canSubmit =
    request?.mode === 'delete'
      ? confirmName.trim() === request.folder?.name
      : name.trim() !== '' && !(request?.mode === 'rename' && name.trim() === request.folder?.name)

  function close(): void {
    closeFolderDialog()
  }

  async function submit(): Promise<void> {
    if (!request || busy || !canSubmit) {
      return
    }
    busy = true
    try {
      if (request.mode === 'create') {
        await createFolder(request.accountId, request.folder?.id ?? 0, name.trim())
        toastInfo($t('folders.created').replace('{name}', name.trim()))
      } else if (request.mode === 'rename') {
        await renameFolder(request.folder!.id, name.trim())
        toastInfo($t('folders.renamed').replace('{name}', name.trim()))
      } else {
        const gone = request.folder!
        await deleteFolder(gone.id)
        // the deleted folder cannot stay selected: its messages are gone and the
        // list would keep querying a folder id that no longer exists.
        if ($selection.kind === 'folder' && $selection.folderId === gone.id) {
          selectView('inbox', $t('sidebar.unifiedInbox'))
        }
        toastInfo($t('folders.deleted').replace('{name}', gone.name))
      }
      await refreshSidebar()
      close()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      busy = false
    }
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      close()
    }
    if (event.key === 'Enter' && canSubmit) {
      void submit()
    }
  }
</script>

<svelte:window on:keydown={request ? onKeydown : undefined} />

{#if request}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
  <div class="backdrop" transition:fade={{ duration: 120 }} on:click={close}></div>
  <div
    class="dialog"
    role="dialog"
    aria-modal="true"
    aria-label={title}
    transition:scale={{ duration: 150, start: 0.94 }}
  >
    <header>
      <h2>{title}</h2>
      <button type="button" class="close" aria-label={$t('folders.cancel')} on:click={close}>
        <IconX size={16} stroke={1.8} />
      </button>
    </header>

    {#if request.mode === 'delete'}
      <div class="warn">
        <span class="warn-icon"><IconAlertTriangle size={17} stroke={1.8} /></span>
        <p>{$t('folders.deleteWarning').replace('{name}', request.folder?.name ?? '')}</p>
      </div>
      <label class="field">
        <span>{$t('folders.deleteConfirmLabel').replace('{name}', request.folder?.name ?? '')}</span>
        <!-- svelte-ignore a11y-autofocus -->
        <input type="text" bind:value={confirmName} autofocus spellcheck="false" />
      </label>
    {:else}
      <label class="field">
        <span>{$t('folders.nameLabel')}</span>
        <!-- svelte-ignore a11y-autofocus -->
        <input
          type="text"
          bind:value={name}
          autofocus
          spellcheck="false"
          placeholder={$t('folders.namePlaceholder')}
        />
      </label>
    {/if}

    <div class="actions">
      <button type="button" class="cancel" on:click={close}>{$t('folders.cancel')}</button>
      <button
        type="button"
        class="go"
        class:danger={request.mode === 'delete'}
        disabled={!canSubmit || busy}
        on:click={submit}
      >
        {request.mode === 'delete' ? $t('action.delete') : $t('folders.save')}
      </button>
    </div>
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
    padding: var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
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

  .warn {
    display: flex;
    gap: var(--space-3);
    padding: var(--space-3);
    border-radius: var(--radius-control);
    background: var(--danger-bg);
  }

  .warn-icon {
    display: inline-flex;
    flex-shrink: 0;
    color: var(--danger);
  }

  .warn p {
    margin: 0;
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
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

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
  }

  .cancel {
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
    color: var(--text-secondary);
    background: transparent;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    cursor: pointer;
  }
  .cancel:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .go {
    padding: var(--space-2) var(--space-4);
    font-size: var(--fz-label);
    font-weight: var(--fw-medium);
    color: var(--accent-fg);
    background: var(--accent);
    border: none;
    border-radius: var(--radius-control);
    cursor: pointer;
  }
  .go.danger {
    color: var(--text-inverse);
    background: var(--danger);
  }
  .go:disabled {
    opacity: 0.5;
    cursor: default;
  }
</style>
