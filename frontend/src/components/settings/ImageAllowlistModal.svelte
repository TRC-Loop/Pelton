<script lang="ts">
  // "Manage whitelist" modal: lists every sender and domain the user has trusted
  // for remote content (images), lets them open an example message to remember
  // who it is, and revoke any entry. Reached from the Privacy settings.
  import { createEventDispatcher, onMount } from 'svelte'
  import Modal from '../common/Modal.svelte'
  import { IconTrash, IconMail, IconUser, IconWorld } from '@tabler/icons-svelte'
  import { listImageAllowlist, removeImageAllow, type ImageAllowEntry } from '../../lib/api'
  import { openMessage } from '../../stores/selection'
  import { errorMessage, toastError } from '../../stores/toast'
  import { t } from '../../lib/i18n'

  const dispatch = createEventDispatcher<{ close: void; openMessage: void }>()

  let entries: ImageAllowEntry[] = []
  let loading = true

  onMount(load)

  async function load(): Promise<void> {
    loading = true
    try {
      entries = await listImageAllowlist()
    } catch (err) {
      toastError(errorMessage(err))
    } finally {
      loading = false
    }
  }

  async function remove(entry: ImageAllowEntry): Promise<void> {
    try {
      await removeImageAllow(entry.kind as 'sender' | 'domain', entry.value)
      entries = entries.filter((e) => !(e.kind === entry.kind && e.value === entry.value))
    } catch (err) {
      toastError(errorMessage(err))
    }
  }

  // open the example message in the detail pane and leave settings so it shows.
  function show(entry: ImageAllowEntry): void {
    if (!entry.exampleMessageId) {
      return
    }
    openMessage(entry.exampleMessageId)
    dispatch('openMessage')
  }

</script>

<Modal
  title={$t('imageAllowlist.title')}
  hint={$t('imageAllowlist.hint')}
  size="small"
  on:close={() => dispatch('close')}
>
  {#if loading}
    <p class="empty">{$t('imageAllowlist.loading')}</p>
  {:else if entries.length === 0}
    <p class="empty">{$t('imageAllowlist.empty')}</p>
  {:else}
    <ul class="list">
      {#each entries as entry (entry.kind + ':' + entry.value)}
        <li>
          <span class="kind" title={entry.kind === 'domain' ? $t('imageAllowlist.kindDomain') : $t('imageAllowlist.kindSender')}>
            {#if entry.kind === 'domain'}
              <IconWorld size={15} stroke={1.7} />
            {:else}
              <IconUser size={15} stroke={1.7} />
            {/if}
          </span>
          <div class="detail">
            <span class="value">{entry.value}</span>
            {#if entry.exampleMessageId}
              <button type="button" class="example" on:click={() => show(entry)}>
                <IconMail size={12} stroke={1.7} />
                <span>{entry.exampleSubject || $t('messageList.noSubject')}</span>
              </button>
            {/if}
          </div>
          <button
            type="button"
            class="remove"
            aria-label={$t('imageAllowlist.remove')}
            title={$t('imageAllowlist.remove')}
            on:click={() => remove(entry)}
          >
            <IconTrash size={15} stroke={1.7} />
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</Modal>

<style>

  .empty {
    margin: var(--space-4) 0;
    text-align: center;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
  }

  li {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) 0;
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .kind {
    flex-shrink: 0;
    color: var(--text-tertiary);
    display: inline-flex;
  }

  .detail {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }

  .value {
    font-size: var(--fz-label);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .example {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    max-width: 100%;
    border: none;
    background: transparent;
    padding: 0;
    color: var(--text-tertiary);
    font-size: var(--fz-meta);
    cursor: var(--cursor-action);
    text-align: left;
  }
  .example:hover {
    color: var(--accent);
  }
  .example span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .remove {
    flex-shrink: 0;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: var(--cursor-action);
    padding: var(--space-2);
    border-radius: var(--radius-control);
  }
  .remove:hover {
    background: var(--danger-bg, var(--surface-hover));
    color: var(--danger);
  }
</style>
