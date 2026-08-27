<script lang="ts">
  // the approval queue for mail an agent wants to send (#127).
  //
  // Nothing here has been sent and nothing has reached a server. An agent that
  // reads a message can be talked into proposing anything the message asks for,
  // so the recipient and the body are shown in full and the decision stays with
  // the reader.
  import { IconSend, IconTrash, IconRobot } from '@tabler/icons-svelte'
  import Modal from './Modal.svelte'
  import { proposals, answering, approve, discard } from '../../stores/agent'
  import { t } from '../../lib/i18n'

  let open = false

  // the oldest proposal is the one on show; answering it brings up the next.
  $: current = $proposals[0]
  $: waiting = $proposals.length
</script>

{#if waiting > 0}
  <button type="button" class="chip" on:click={() => (open = true)}>
    <IconRobot size={13} stroke={1.8} />
    {$t('agent.waiting').replace('{n}', String(waiting))}
  </button>
{/if}

{#if open && current}
  <Modal
    title={$t('agent.reviewTitle')}
    hint={$t('agent.reviewHint')}
    size="medium"
    on:close={() => (open = false)}
  >
    <dl class="fields">
      <dt>{$t('compose.field.to')}</dt>
      <dd>{current.to}</dd>
      {#if current.cc}
        <dt>{$t('compose.field.cc')}</dt>
        <dd>{current.cc}</dd>
      {/if}
      {#if current.bcc}
        <dt>{$t('compose.field.bcc')}</dt>
        <dd>{current.bcc}</dd>
      {/if}
      <dt>{$t('compose.field.subject')}</dt>
      <dd>{current.subject}</dd>
    </dl>

    <!-- plain text, never rendered: an agent's body is as untrusted as the mail
         that may have prompted it. -->
    <pre class="body">{current.body}</pre>

    {#if waiting > 1}
      <p class="more">{$t('agent.more').replace('{n}', String(waiting - 1))}</p>
    {/if}

    <svelte:fragment slot="footer">
      <button
        type="button"
        class="ghost"
        disabled={$answering}
        on:click={() => void discard(current.id)}
      >
        <IconTrash size={14} stroke={1.8} />
        {$t('agent.discard')}
      </button>
      <button
        type="button"
        class="primary"
        disabled={$answering}
        on:click={() => void approve(current.id)}
      >
        <IconSend size={14} stroke={1.8} />
        {$t('agent.send')}
      </button>
    </svelte:fragment>
  </Modal>
{/if}

<style>
  .chip {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: 0 var(--space-2);
    height: 100%;
    border: none;
    border-radius: var(--radius-control);
    background: transparent;
    color: var(--warning);
    font-size: var(--fz-meta);
    cursor: var(--cursor-action);
  }
  .chip:hover {
    background: var(--surface-hover);
  }

  .fields {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: var(--space-1) var(--space-3);
    margin: 0 0 var(--space-3);
  }
  dt {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }
  dd {
    margin: 0;
    font-size: var(--fz-label);
    color: var(--text-primary);
    overflow-wrap: anywhere;
  }

  .body {
    margin: 0;
    padding: var(--space-3);
    max-height: 40vh;
    overflow: auto;
    border: var(--hairline) solid var(--border-subtle);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-secondary);
    font-family: var(--font-mono);
    font-size: var(--fz-label);
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .more {
    margin: var(--space-3) 0 0;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .ghost,
  .primary {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-4);
    border-radius: var(--radius-control);
    font-size: var(--fz-label);
    cursor: var(--cursor-action);
  }
  .ghost {
    border: var(--hairline) solid var(--border-default);
    background: transparent;
    color: var(--text-secondary);
  }
  .ghost:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }
  .primary {
    border: var(--hairline) solid transparent;
    background: var(--accent);
    color: var(--accent-fg);
  }
  .primary:hover {
    filter: brightness(1.05);
  }
  .ghost:disabled,
  .primary:disabled {
    opacity: 0.6;
    cursor: default;
  }
</style>
