<script lang="ts">
  // one row in the message list. its layout follows the chosen row template
  // (relaxed, comfortable, compact, single) with per-field overrides for the
  // avatar and snippet and a clamped preview-line count. unread shows by weight
  // and a dot (never accent). flagged rows highlight per the user's chosen style
  // (flag icon, a left bar, both icon+bar, or off). multi-selection is driven by
  // modifier-clicks from the list; the technical-info badges live in the detail.
  import { IconPaperclip, IconFlagFilled } from '@tabler/icons-svelte'
  import Avatar from '../common/Avatar.svelte'
  import { prefs } from '../../stores/prefs'
  import { formatListDate, displayName } from '../../lib/format'
  import type { MessageSummary, RowTemplate } from '../../lib/types'

  export let message: MessageSummary
  export let selected: boolean = false
  // checked reflects membership in the bulk-selection set (modifier-click).
  export let checked: boolean = false

  $: unread = !message.seen
  $: template = $prefs.rowTemplate as RowTemplate

  // the avatar shows on the spacious templates when the user keeps it on; single
  // line has no room for it.
  $: showAvatar = $prefs.rowShowAvatar && template !== 'single'
  // the snippet only fits on the multi-line templates.
  $: showSnippet =
    $prefs.rowShowSnippet && (template === 'relaxed' || template === 'comfortable') && !!message.snippet
  $: avatarSize = template === 'relaxed' ? 36 : 30
  $: previewLines = $prefs.previewLines

  // flag highlight styles. "both" shows the flag icon and a left edge bar.
  $: flagStyle = $prefs.flagHighlight
  $: showFlagIcon = message.flagged && (flagStyle === 'flag' || flagStyle === 'both')
  $: barLeft = message.flagged && (flagStyle === 'left' || flagStyle === 'both')
</script>

<!-- keyboard navigation for the list is handled by the listbox container
     (arrow keys + enter), so the row itself only needs the click affordance. -->
<!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
<div
  class="row {template}"
  class:selected
  class:unread
  class:checked
  class:bar-left={barLeft}
  role="option"
  tabindex="-1"
  aria-selected={selected}
  on:click
  on:dblclick
  on:contextmenu
>
  {#if showAvatar}
    <Avatar name={message.fromName} email={message.fromAddress} size={avatarSize} />
  {/if}

  <div class="content">
    <div class="line top">
      <span class="dot" class:show={unread} aria-hidden="true"></span>
      <span class="sender">{displayName(message.fromName, message.fromAddress)}</span>
      {#if template === 'single'}
        <span class="subject inline">{message.subject || '(no subject)'}</span>
      {/if}
      <span class="meta">
        {#if showFlagIcon}
          <IconFlagFilled size={12} class="flag" />
        {/if}
        {#if message.hasAttachments}
          <IconPaperclip size={12} stroke={1.6} />
        {/if}
        {#if $prefs.showDateTime}
          <time class="time">{formatListDate(message.date)}</time>
        {/if}
      </span>
    </div>

    {#if template !== 'single'}
      <div class="subject">{message.subject || '(no subject)'}</div>
    {/if}
    {#if showSnippet}
      <div class="snippet" style={`--preview-lines:${previewLines}`}>{message.snippet}</div>
    {/if}
  </div>
</div>

<style>
  .row {
    position: relative;
    display: flex;
    gap: var(--space-3);
    padding: var(--row-pad-y) var(--row-pad-x);
    border-bottom: var(--hairline) solid var(--border-subtle);
    cursor: pointer;
  }

  /* tighter vertical rhythm as the templates get denser. */
  .row.compact,
  .row.single {
    padding-top: var(--space-2);
    padding-bottom: var(--space-2);
  }

  .row.single {
    align-items: center;
  }

  .row:hover {
    background: var(--surface-hover);
  }

  .row.selected {
    background: var(--selection-bg);
  }

  /* an active multi-selection reads stronger than a plain hover. */
  .row.checked {
    background: var(--selection-bg);
  }

  /* the flag left edge bar uses the warning color, matching the flag icon, and
     sits inside the row so it does not shift the layout. */
  .row.bar-left::before {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    width: 3px;
    background: var(--warning);
  }

  .content {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: var(--row-gap);
  }

  .line.top {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
  }

  .dot {
    width: 7px;
    height: 7px;
    border-radius: 999px;
    background: transparent;
    flex-shrink: 0;
    align-self: center;
  }

  /* unread indicator: a neutral dot, not the accent. */
  .dot.show {
    background: var(--text-primary);
  }

  .sender {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: var(--fz-list);
    color: var(--text-secondary);
    flex: 0 1 auto;
  }

  /* on the multi-line templates the sender takes the remaining width; on a single
     line it shrinks so the inline subject can share the row. */
  .row:not(.single) .sender {
    flex: 1;
  }

  .unread .sender {
    color: var(--text-primary);
    font-weight: var(--fw-semibold);
  }

  .meta {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    flex-shrink: 0;
    margin-left: auto;
    color: var(--text-tertiary);
  }

  .meta :global(.flag) {
    color: var(--warning);
  }

  .time {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .subject {
    font-size: var(--fz-list);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* the inline subject sits next to the sender on the single-line template. */
  .subject.inline {
    flex: 1;
    min-width: 0;
    color: var(--text-secondary);
  }

  .unread .subject {
    font-weight: var(--fw-medium);
  }

  .snippet {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    overflow: hidden;
    /* clamp to the chosen preview-line count. */
    display: -webkit-box;
    -webkit-line-clamp: var(--preview-lines, 1);
    line-clamp: var(--preview-lines, 1);
    -webkit-box-orient: vertical;
  }
</style>
