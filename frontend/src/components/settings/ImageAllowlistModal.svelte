<script lang="ts">
  // "Manage whitelist" modal: lists every sender and domain the user has trusted
  // for remote content (images), lets them open an example message to remember
  // who it is, and revoke any entry. Reached from the Privacy settings.
  import { createEventDispatcher, onMount } from 'svelte'
  import { fade, scale } from 'svelte/transition'
  import { IconX, IconTrash, IconMail, IconUser, IconWorld } from '@tabler/icons-svelte'
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

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      dispatch('close')
    }
  }
</script>

<svelte:window on:keydown={onKeydown} />

<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
<div class="backdrop" transition:fade={{ duration: 120 }} on:click={() => dispatch('close')}></div>
<div
  class="dialog"
  role="dialog"
  aria-modal="true"
  aria-label={$t('imageAllowlist.title')}
  transition:scale={{ duration: 150, start: 0.94 }}
>
  <header>
    <h2>{$t('imageAllowlist.title')}</h2>
    <button type="button" class="close" aria-label={$t('detail.attachments.close')} on:click={() => dispatch('close')}>
      <IconX size={16} stroke={1.8} />
    </button>
  </header>

  <p class="hint">{$t('imageAllowlist.hint')}</p>

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
</div>

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
    width: min(440px, calc(100vw - 2 * var(--space-5)));
    max-height: 72vh;
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
    font-size: var(--fz-heading);
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

  .hint {
    margin: 0;
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    line-height: 1.5;
  }

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
    cursor: pointer;
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
    cursor: pointer;
    padding: var(--space-2);
    border-radius: var(--radius-control);
  }
  .remove:hover {
    background: var(--danger-bg, var(--surface-hover));
    color: var(--danger);
  }
</style>
