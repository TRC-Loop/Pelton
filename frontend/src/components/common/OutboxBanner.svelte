<script lang="ts">
  // a slim status bar surfacing the outbox: how many messages are queued or
  // sending, and any that failed. it makes the durable send queue and its
  // failures visible instead of silent. hidden when the outbox is empty.
  import { IconSend, IconAlertTriangle } from '@tabler/icons-svelte'
  import { outbox } from '../../stores/outbox'

  // outbox states mirror the backend outbox package constants.
  $: pending = $outbox.filter((r) => r.state === 'queued' || r.state === 'sending')
  $: failed = $outbox.filter((r) => r.state === 'failed')
</script>

{#if pending.length > 0 || failed.length > 0}
  <div class="banner" class:has-failed={failed.length > 0}>
    {#if pending.length > 0}
      <span class="item">
        <IconSend size={13} stroke={1.7} />
        {pending.length} sending
      </span>
    {/if}
    {#if failed.length > 0}
      <span class="item failed" title={failed[0].lastError}>
        <IconAlertTriangle size={13} stroke={1.7} />
        {failed.length} failed
      </span>
    {/if}
  </div>
{/if}

<style>
  .banner {
    display: flex;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-1) var(--space-4);
    background: var(--surface-sunken);
    border-top: var(--hairline) solid var(--border-subtle);
    font-size: var(--fz-meta);
    color: var(--text-secondary);
  }

  .banner.has-failed {
    background: var(--danger-bg);
  }

  .item {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
  }

  .item.failed {
    color: var(--danger);
  }
</style>
