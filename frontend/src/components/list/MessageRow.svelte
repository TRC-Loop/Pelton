<script lang="ts">
  // one row in the message list. its layout follows the chosen row template
  // (relaxed, comfortable, compact, single) with per-field overrides for the
  // avatar and snippet and a clamped preview-line count. unread shows by weight
  // and a dot (never accent). flagged rows highlight per the user's chosen style
  // (flag icon, a left bar, both icon+bar, or off); a color label paints the left
  // bar and a small dot. offline and snoozed states show little meta icons.
  // trackpad swipes (horizontal wheel) reveal a configurable action.
  import { createEventDispatcher } from 'svelte'
  import {
    IconPaperclip,
    IconFlagFilled,
    IconFlag,
    IconTrash,
    IconMailOpened,
    IconMailFilled,
    IconArchive,
    IconClock,
    IconDownload,
    IconClockPause,
  } from '@tabler/icons-svelte'
  import Avatar from '../common/Avatar.svelte'
  import { prefs } from '../../stores/prefs'
  import { formatListDate, displayName } from '../../lib/format'
  import { flagColorHex } from '../../theme/flagcolors'
  import type { MessageSummary, RowTemplate, SwipeAction } from '../../lib/types'
  import { t } from '../../lib/i18n'

  export let message: MessageSummary
  export let selected: boolean = false
  // checked reflects membership in the bulk-selection set (modifier-click).
  export let checked: boolean = false

  const dispatch = createEventDispatcher<{ swipe: 'left' | 'right' }>()

  $: unread = !message.seen
  $: template = $prefs.rowTemplate as RowTemplate

  $: showAvatar = $prefs.rowShowAvatar && template !== 'single'
  $: showSnippet =
    $prefs.rowShowSnippet && (template === 'relaxed' || template === 'comfortable') && !!message.snippet
  $: avatarSize = template === 'relaxed' ? 36 : 30
  $: previewLines = $prefs.previewLines

  // flag highlight styles. "both" shows the flag icon and a left edge bar.
  $: flagStyle = $prefs.flagHighlight
  $: showFlagIcon = message.flagged && (flagStyle === 'flag' || flagStyle === 'both')
  $: barLeft = message.flagged && (flagStyle === 'left' || flagStyle === 'both')

  // a color label paints the left bar (taking precedence over the warning bar)
  // and a small dot in the meta row.
  $: colorHex = flagColorHex(message.flagColor)
  $: showBar = !!colorHex || barLeft
  $: barColor = colorHex || 'var(--warning)'
  $: snoozed = !!message.snoozeUntil
  $: showOffline = message.offline && $prefs.showOfflineIndicator

  // swipe gesture state. horizontal wheel deltas (trackpad) accumulate into a
  // translation that reveals the action behind the row, apple-mail style: the
  // reveal grows as you swipe and the action only fires once you have dragged
  // most of the way across. a short idle timer ends the gesture. requiring a
  // large, clearly-horizontal drag also stops a vertical scroll from triggering.
  let wrapWidth = 320
  let offset = 0
  let settling = false
  let engaged = false
  let committed = false
  // pre-engage accumulation: we only start a swipe once there is a coherent burst
  // of horizontal movement, so ordinary vertical scrolling is never hijacked.
  let preDX = 0
  let idleTimer: ReturnType<typeof setTimeout> | undefined

  // commit once dragged past ~55% of the row; the reveal can grow to the full row.
  $: commitDistance = wrapWidth * 0.55
  $: maxReveal = wrapWidth

  $: leftAction = $prefs.swipeLeftAction as SwipeAction
  $: rightAction = $prefs.swipeRightAction as SwipeAction

  $: actionMeta = {
    none: { label: '', color: 'var(--surface-hover)', icon: IconFlag },
    delete: { label: $t('action.delete'), color: 'var(--danger)', icon: IconTrash },
    read: { label: $t('messageList.swipe.read'), color: 'var(--accent)', icon: IconMailOpened },
    unread: { label: $t('messageList.swipe.unread'), color: 'var(--accent)', icon: IconMailFilled },
    flag: { label: $t('messageList.flag'), color: 'var(--warning)', icon: IconFlagFilled },
    archive: { label: $t('action.archive'), color: 'var(--text-tertiary)', icon: IconArchive },
    snooze: { label: $t('shortcut.snooze'), color: 'var(--accent)', icon: IconClockPause },
  } as Record<SwipeAction, { label: string; color: string; icon: typeof IconTrash }>

  // revealed is the action about to fire given the current drag direction. a
  // positive offset (row dragged right) reveals the right-swipe action on the
  // left edge; negative reveals the left-swipe action on the right edge.
  $: revealAction = offset > 0 ? rightAction : leftAction
  $: revealMeta = actionMeta[revealAction] ?? actionMeta.none

  function onWheel(event: WheelEvent): void {
    if (!$prefs.swipeEnabled || (leftAction === 'none' && rightAction === 'none')) {
      return
    }
    const dx = event.deltaX
    const dy = event.deltaY
    if (!engaged) {
      // a clearly vertical event: let it scroll and reset any partial buffer, so
      // stray horizontal jitter during scrolling never accumulates into a swipe.
      if (Math.abs(dy) > Math.abs(dx) * 1.5) {
        preDX = 0
        return
      }
      // accumulate horizontal intent; engage only after a deliberate burst.
      preDX += dx
      if (Math.abs(preDX) < 26) {
        return
      }
      engaged = true
      committed = false
      offset = 0
    }
    event.preventDefault()
    settling = false
    offset = clamp(offset - dx, -maxReveal, maxReveal)
    if (idleTimer) {
      clearTimeout(idleTimer)
    }
    idleTimer = setTimeout(endSwipe, 130)
  }

  function endSwipe(): void {
    if (!engaged) {
      return
    }
    const dir = offset > 0 ? 'right' : 'left'
    const action = offset > 0 ? rightAction : leftAction
    const passed = Math.abs(offset) >= commitDistance
    engaged = false
    preDX = 0
    settling = true
    if (passed && action !== 'none' && !committed) {
      // slide the rest of the way so the whole row shows the action, then fire.
      committed = true
      offset = dir === 'right' ? maxReveal : -maxReveal
      const fireDir = dir
      setTimeout(() => {
        offset = 0
        committed = false
        dispatch('swipe', fireDir)
      }, 150)
    } else {
      offset = 0
    }
  }

  function clamp(v: number, lo: number, hi: number): number {
    return Math.max(lo, Math.min(hi, v))
  }
</script>

<!-- keyboard navigation for the list is handled by the listbox container
     (arrow keys + enter), so the row itself only needs the click affordance. -->
<div class="swipe-wrap" class:swiping={offset !== 0} bind:clientWidth={wrapWidth}>
  {#if offset !== 0}
    <div
      class="reveal"
      class:right-edge={offset < 0}
      style={`--reveal-color:${revealMeta.color}`}
      aria-hidden="true"
    >
      <svelte:component this={revealMeta.icon} size={18} stroke={1.7} />
      <span>{revealMeta.label}</span>
    </div>
  {/if}
  <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
  <div
    class="row {template}"
    class:selected
    class:unread
    class:checked
    class:bar-left={showBar}
    class:settling
    style={`transform:translateX(${offset}px);--bar-color:${barColor}`}
    role="option"
    tabindex="-1"
    aria-selected={selected}
    on:click
    on:dblclick
    on:contextmenu
    on:wheel={onWheel}
  >
    {#if showAvatar}
      <Avatar name={message.fromName} email={message.fromAddress} size={avatarSize} />
    {/if}

    <div class="content">
      <div class="line top">
        <span class="dot" class:show={unread} aria-hidden="true"></span>
        <span class="sender">{displayName(message.fromName, message.fromAddress)}</span>
        {#if template === 'single'}
          <span class="subject inline">{message.subject || $t('messageList.noSubject')}</span>
        {/if}
        <span class="meta">
          {#if message.flagColor > 0}
            <span class="color-dot" style={`background:${colorHex}`} aria-hidden="true"></span>
          {/if}
          {#if snoozed}
            <IconClock size={12} stroke={1.7} class="snooze" />
          {/if}
          {#if showOffline}
            <IconDownload size={12} stroke={1.7} class="offline" />
          {/if}
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
        <div class="subject">{message.subject || $t('messageList.noSubject')}</div>
      {/if}
      {#if showSnippet}
        <div class="snippet" style={`--preview-lines:${previewLines}`}>{message.snippet}</div>
      {/if}
    </div>
  </div>
</div>

<style>
  .swipe-wrap {
    position: relative;
    overflow: hidden;
  }

  /* the action revealed behind the row while swiping. */
  .reveal {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0 var(--space-4);
    color: #fff;
    font-size: var(--fz-meta);
    font-weight: var(--fw-semibold);
    background: var(--reveal-color);
  }
  .reveal.right-edge {
    justify-content: flex-end;
  }

  .row {
    position: relative;
    display: flex;
    gap: var(--space-3);
    padding: var(--row-pad-y) var(--row-pad-x);
    border-bottom: var(--hairline) solid var(--border-subtle);
    cursor: pointer;
    background: var(--surface-raised);
  }

  /* settle back smoothly after a gesture ends; live dragging has no transition. */
  .row.settling {
    transition: transform 0.18s cubic-bezier(0.22, 1, 0.36, 1);
  }

  /* tighter vertical rhythm as the templates get denser. relaxed keeps the roomy
     default; comfortable is a step tighter (but still shows the snippet, unlike
     compact) so the two read differently. */
  .row.comfortable {
    padding-top: var(--space-2);
    padding-bottom: var(--space-2);
  }
  .row.compact,
  .row.single {
    padding-top: var(--space-2);
    padding-bottom: var(--space-2);
  }
  .row.comfortable .content {
    gap: 1px;
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

  .row.checked {
    background: var(--selection-bg);
  }

  /* the left edge bar: warning color for a flag, or the label color when set. */
  .row.bar-left::before {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    width: 3px;
    background: var(--bar-color);
  }

  .content {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: var(--row-gap);
  }

  /* center alignment keeps the sender and the meta (time, flag, icons) on a stable
     baseline: with baseline alignment, adding the flag icon nudged the date up. */
  .line.top {
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
    align-self: center;
  }

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

  .meta :global(.offline) {
    color: var(--success, var(--accent));
  }

  .meta :global(.snooze) {
    color: var(--accent);
  }

  .color-dot {
    width: 8px;
    height: 8px;
    border-radius: 999px;
    flex-shrink: 0;
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
    display: -webkit-box;
    -webkit-line-clamp: var(--preview-lines, 1);
    line-clamp: var(--preview-lines, 1);
    -webkit-box-orient: vertical;
  }
</style>
