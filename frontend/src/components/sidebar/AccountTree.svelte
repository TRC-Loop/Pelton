<script lang="ts">
  // one account's section in the sidebar: a header with the address and the full
  // folder tree below it. the header chevron collapses the whole account. root
  // folders are those without a parent; the rest nest via FolderNode recursion.
  import { IconChevronRight } from '@tabler/icons-svelte'
  import FolderNode from './FolderNode.svelte'
  import { prefs } from '../../stores/prefs'
  import { collapsedAccounts, toggleAccountCollapsed } from '../../stores/sidebarstate'
  import type { Account, Folder } from '../../lib/types'
  import { t } from '../../lib/i18n'

  export let account: Account
  export let folders: Folder[]

  // the account section is expanded by default; the collapsed set (persisted) is
  // the source of truth so the choice survives restarts.
  $: expanded = !$collapsedAccounts.has(account.id)

  // roots have no parent. loose equality catches both null and undefined.
  $: roots = folders.filter((f) => f.parentId == null)
  // the header shows the display name, or the email when the user prefers it (or
  // when there is no display name to show).
  $: label = $prefs.showAccountEmail ? account.email : account.displayName || account.email
</script>

<section class="account">
  <button
    type="button"
    class="account-head"
    class:open={expanded}
    title={account.email}
    aria-expanded={expanded}
    on:click={() => toggleAccountCollapsed(account.id)}
  >
    <span class="account-caret" aria-hidden="true"><IconChevronRight size={13} stroke={1.9} /></span>
    <span class="account-name">{label}</span>
  </button>
  {#if expanded}
    {#if roots.length === 0}
      <p class="empty-note">{$t('sidebar.account.noFolders')}</p>
    {:else}
      {#each roots as folder (folder.id)}
        <FolderNode {folder} {folders} depth={0} />
      {/each}
    {/if}
  {/if}
</section>

<style>
  .account {
    margin-top: var(--space-4);
  }

  .account-head {
    display: flex;
    align-items: center;
    gap: var(--space-1);
    width: 100%;
    padding: var(--space-2) var(--space-3);
    border: none;
    background: transparent;
    cursor: pointer;
    text-align: left;
    border-radius: var(--radius-control);
  }

  .account-head:hover {
    background: var(--surface-hover);
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
