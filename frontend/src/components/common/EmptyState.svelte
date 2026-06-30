<script lang="ts">
  // a neutral placeholder for empty panes (empty folder, no message selected, no
  // accounts yet). icon goes in the default slot; an optional action goes in the
  // "action" slot.
  export let title: string
  export let detail: string = ''
</script>

<div class="empty">
  <div class="icon" aria-hidden="true"><slot /></div>
  <p class="title">{title}</p>
  {#if detail}
    <p class="detail">{detail}</p>
  {/if}
  <slot name="action" />
</div>

<style>
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    height: 100%;
    padding: var(--space-6);
    text-align: center;
    color: var(--text-tertiary);
  }

  .icon {
    color: var(--text-tertiary);
    opacity: 0.7;
  }

  .title {
    margin: 0;
    font-size: var(--fz-body);
    font-weight: var(--fw-medium);
    color: var(--text-secondary);
  }

  .detail {
    margin: 0;
    max-width: 32ch;
    font-size: var(--fz-label);
    line-height: 1.5;
  }
</style>
