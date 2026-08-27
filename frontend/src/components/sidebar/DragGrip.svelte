<script lang="ts">
  // the grab handle that makes a sidebar row draggable. it appears on hover so
  // the sidebar stays quiet at rest. the reorder action finds it by attribute
  // and refuses to start a drag anywhere else, so clicking a row still selects.
  import { IconGripVertical } from '@tabler/icons-svelte'
  import { reorderHandleAttr } from '../../lib/reorder'
  import { t } from '../../lib/i18n'
</script>

<span class="grip" title={$t('sidebar.dragHandle')} aria-hidden="true" {...{ [reorderHandleAttr]: '' }}>
  <IconGripVertical size={13} stroke={1.7} />
</span>

<style>
  .grip {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    padding: 0 var(--space-1);
    color: var(--text-tertiary);
    cursor: grab;
    opacity: 0;
  }

  /* revealed by the row's own hover rule, which lives in the parent so the grip
     does not need to know what kind of row it sits in. */
  :global(.row:hover) .grip,
  :global(.account-row:hover) .grip {
    opacity: 1;
  }

  .grip:active {
    cursor: grabbing;
  }
</style>
