<script lang="ts">
  // the message action toolbar: reply, reply-all, forward, archive, delete, flag.
  // it only dispatches intents; the detail container performs them. icon-only
  // buttons carry aria-labels via IconButton.
  import { createEventDispatcher } from 'svelte'
  import {
    IconArrowBackUp,
    IconArrowBackUpDouble,
    IconArrowForwardUp,
    IconArchive,
    IconTrash,
    IconFlag,
    IconFlagFilled,
    IconPrinter,
    IconInfoCircle,
  } from '@tabler/icons-svelte'
  import IconButton from '../common/IconButton.svelte'
  import { t } from '../../lib/i18n'

  export let flagged: boolean = false

  const dispatch = createEventDispatcher<{
    reply: void
    replyAll: void
    forward: void
    archive: void
    delete: void
    toggleFlag: void
    print: void
    info: void
  }>()

  $: flagLabel = flagged ? $t('detail.toolbar.unflag') : $t('detail.toolbar.flag')
</script>

<div class="toolbar" role="toolbar" aria-label={$t('detail.toolbar.ariaLabel')}>
  <IconButton label={$t('action.reply')} on:click={() => dispatch('reply')}>
    <IconArrowBackUp size={18} stroke={1.6} />
  </IconButton>
  <IconButton label={$t('detail.toolbar.replyAll')} on:click={() => dispatch('replyAll')}>
    <IconArrowBackUpDouble size={18} stroke={1.6} />
  </IconButton>
  <IconButton label={$t('action.forward')} on:click={() => dispatch('forward')}>
    <IconArrowForwardUp size={18} stroke={1.6} />
  </IconButton>

  <span class="divider" aria-hidden="true"></span>

  <IconButton label={$t('action.archive')} on:click={() => dispatch('archive')}>
    <IconArchive size={18} stroke={1.6} />
  </IconButton>
  <IconButton label={flagLabel} active={flagged} on:click={() => dispatch('toggleFlag')}>
    {#if flagged}
      <IconFlagFilled size={18} />
    {:else}
      <IconFlag size={18} stroke={1.6} />
    {/if}
  </IconButton>
  <IconButton label={$t('action.delete')} danger on:click={() => dispatch('delete')}>
    <IconTrash size={18} stroke={1.6} />
  </IconButton>

  <span class="divider" aria-hidden="true"></span>

  <IconButton label={$t('detail.toolbar.print')} on:click={() => dispatch('print')}>
    <IconPrinter size={18} stroke={1.6} />
  </IconButton>
  <IconButton label={$t('detail.toolbar.messageInfo')} on:click={() => dispatch('info')}>
    <IconInfoCircle size={18} stroke={1.6} />
  </IconButton>
</div>

<style>
  .toolbar {
    display: flex;
    align-items: center;
    gap: var(--space-1);
  }

  .divider {
    width: var(--hairline);
    height: 18px;
    margin: 0 var(--space-2);
    background: var(--border-default);
  }
</style>
