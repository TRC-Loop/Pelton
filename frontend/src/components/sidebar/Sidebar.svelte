<script lang="ts">
  // column 1: the sidebar. a compose action at the top, then the unified views,
  // then every account's full folder tree, with a settings entry pinned to the
  // bottom. it renders explicit loading, error and empty (no accounts) states.
  import { createEventDispatcher } from 'svelte'
  import { IconPencil, IconRefresh, IconMailbox, IconPlus } from '@tabler/icons-svelte'
  import UnifiedViews from './UnifiedViews.svelte'
  import AccountTree from './AccountTree.svelte'
  import Spinner from '../common/Spinner.svelte'
  import ErrorState from '../common/ErrorState.svelte'
  import EmptyState from '../common/EmptyState.svelte'
  import { sidebar, loadSidebar } from '../../stores/accounts'
  import { syncing } from '../../stores/outbox'
  import { t } from '../../lib/i18n'

  const dispatch = createEventDispatcher<{ compose: void; sync: void; addMailbox: void }>()
</script>

<aside class="sidebar">
  <div class="top">
    <button type="button" class="compose-btn" on:click={() => dispatch('compose')}>
      <IconPencil size={16} stroke={1.8} />
      <span>{$t('action.compose')}</span>
    </button>
    <button
      type="button"
      class="sync-btn"
      class:spinning={$syncing}
      aria-label={$t('shortcut.sync')}
      title={$t('shortcut.sync')}
      on:click={() => dispatch('sync')}
    >
      <IconRefresh size={16} stroke={1.8} />
    </button>
  </div>

  <div class="scroll">
    {#if $sidebar.status === 'loading' && !$sidebar.data}
      <Spinner label={$t('sidebar.loading.mailboxes')} />
    {:else if $sidebar.status === 'error'}
      <ErrorState message={$sidebar.error} onRetry={loadSidebar} />
    {:else if $sidebar.data}
      {#if $sidebar.data.accounts.length === 0}
        <EmptyState
          title={$t('sidebar.empty.title')}
          detail={$t('sidebar.empty.detail')}
        >
          <IconMailbox size={28} stroke={1.4} />
          <button slot="action" type="button" class="add-mailbox" on:click={() => dispatch('addMailbox')}>
            <IconPlus size={15} stroke={1.8} />
            {$t('addMailbox.cta')}
          </button>
        </EmptyState>
      {:else}
        <UnifiedViews views={$sidebar.data.views} />
        {#each $sidebar.data.accounts as account (account.id)}
          <AccountTree {account} folders={$sidebar.data.foldersByAccount[account.id] ?? []} />
        {/each}
      {/if}
    {/if}
  </div>
</aside>

<style>
  .sidebar {
    display: grid;
    grid-template-rows: auto 1fr;
    height: 100%;
    background: var(--surface-base);
    border-right: var(--hairline) solid var(--border-default);
    min-width: 0;
  }

  .top {
    display: flex;
    gap: var(--space-2);
    padding: var(--space-3);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .compose-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-2);
    flex: 1;
    height: var(--control-height);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-weight: var(--fw-medium);
    cursor: pointer;
  }

  .compose-btn:hover {
    background: var(--surface-hover);
  }

  .sync-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: var(--control-height);
    height: var(--control-height);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-secondary);
    cursor: pointer;
  }

  .sync-btn:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .sync-btn.spinning :global(svg) {
    animation: spin 0.8s linear infinite;
  }

  /* spin counter-clockwise so the refresh arrows read as rewinding/reloading,
     which is the direction that looked right. */
  @keyframes spin {
    to {
      transform: rotate(-360deg);
    }
  }

  .scroll {
    /* min-height:0 lets the 1fr middle track scroll rather than push the footer
       off the bottom when there are many folders. */
    min-height: 0;
    overflow-y: auto;
    padding: var(--space-3) var(--space-2);
  }

  .add-mailbox {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    margin-top: var(--space-2);
    padding: var(--space-2) var(--space-4);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-label);
    cursor: pointer;
  }

  .add-mailbox:hover {
    background: var(--surface-hover);
  }
</style>
