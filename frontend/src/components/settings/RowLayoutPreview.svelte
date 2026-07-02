<script lang="ts">
  // a small, static preview of the message-list row layout, reflecting the live
  // row settings (template, avatar, snippet, preview lines, flag highlight). it
  // mirrors MessageRow's structure with sample data so the user sees the effect
  // of each control without opening the list.
  import { prefs } from '../../stores/prefs'
  import type { RowTemplate } from '../../lib/types'

  const samples = [
    { from: 'Ada Lovelace', subject: 'Notes on the Analytical Engine', snippet: 'I have been thinking about how the machine might weave algebraic patterns, and…', unread: true, flagged: true },
    { from: 'potato@pelton.email', subject: 'Lunch on Friday?', snippet: 'Are we still on for that place near the park? They have the good fries.', unread: false, flagged: false },
  ]

  $: template = $prefs.rowTemplate as RowTemplate
  $: showAvatar = $prefs.rowShowAvatar && template !== 'single'
  $: showSnippet = $prefs.rowShowSnippet && (template === 'relaxed' || template === 'comfortable')
  $: flag = $prefs.flagHighlight
</script>

<div class="preview" aria-hidden="true">
  {#each samples as s}
    <div class="row {template}" class:unread={s.unread} class:bar={s.flagged && (flag === 'left' || flag === 'both')}>
      {#if showAvatar}
        <span class="avatar">{s.from.slice(0, 1).toUpperCase()}</span>
      {/if}
      <div class="content">
        <div class="top">
          <span class="dot" class:show={s.unread}></span>
          <span class="from">{s.from}</span>
          {#if template === 'single'}
            <span class="subj inline">{s.subject}</span>
          {/if}
          <span class="meta">
            {#if s.flagged && (flag === 'flag' || flag === 'both')}<span class="flag">⚑</span>{/if}
            <span class="time">9:24</span>
          </span>
        </div>
        {#if template !== 'single'}
          <div class="subj">{s.subject}</div>
        {/if}
        {#if showSnippet}
          <div class="snippet" style={`--lines:${$prefs.previewLines}`}>{s.snippet}</div>
        {/if}
      </div>
    </div>
  {/each}
</div>

<style>
  .preview {
    margin-top: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    overflow: hidden;
    background: var(--surface-raised);
  }

  .row {
    position: relative;
    display: flex;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }
  .row:last-child {
    border-bottom: none;
  }
  .row.comfortable,
  .row.compact,
  .row.single {
    padding-top: var(--space-2);
    padding-bottom: var(--space-2);
  }
  .row.comfortable .content {
    gap: 0;
  }
  .row.single {
    align-items: center;
  }
  /* relaxed carries a larger avatar than the denser templates. */
  .row.comfortable .avatar,
  .row.compact .avatar {
    width: 28px;
    height: 28px;
  }
  .row.bar::before {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    width: 3px;
    background: var(--warning);
  }

  .avatar {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    flex-shrink: 0;
    border-radius: 999px;
    background: var(--surface-sunken);
    color: var(--text-secondary);
    font-size: var(--fz-label);
  }

  .content {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .top {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .dot {
    width: 7px;
    height: 7px;
    border-radius: 999px;
    background: transparent;
    flex-shrink: 0;
  }
  .dot.show {
    background: var(--text-primary);
  }
  .from {
    font-size: var(--fz-list);
    color: var(--text-secondary);
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .unread .from {
    color: var(--text-primary);
    font-weight: var(--fw-semibold);
  }
  .meta {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    margin-left: auto;
    color: var(--text-tertiary);
    flex-shrink: 0;
  }
  .flag {
    color: var(--warning);
  }
  .time {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }
  .subj {
    font-size: var(--fz-list);
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .subj.inline {
    flex: 1;
    color: var(--text-secondary);
  }
  .unread .subj {
    font-weight: var(--fw-medium);
  }
  .snippet {
    font-size: var(--fz-label);
    color: var(--text-tertiary);
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: var(--lines, 1);
    line-clamp: var(--lines, 1);
    -webkit-box-orient: vertical;
  }
</style>
