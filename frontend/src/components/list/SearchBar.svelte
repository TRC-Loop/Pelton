<script lang="ts">
  // the search field at the top of the message list. besides free text it accepts
  // typed keyword chips (from:/sender:, to:, subject:, has:attachment) and date
  // chips (before:/after:). a keyword token is committed to a chip on space or
  // enter; an autocomplete dropdown suggests keywords as you type; the calendar
  // button inserts before/after date chips. everything is emitted as free text
  // plus a structured SearchFilter so the list re-runs the ranked search.
  import { createEventDispatcher, tick } from 'svelte'
  import { IconSearch, IconX, IconCalendar } from '@tabler/icons-svelte'
  import { prefs } from '../../stores/prefs'
  import { shortcutLabel, t } from '../../lib/i18n'
  import { emptyFilter, type SearchFilter } from '../../stores/messages'
  import { selection } from '../../stores/selection'
  import DateTimePicker from '../common/DateTimePicker.svelte'

  export let value: string = ''

  // switching folder/view clears the query upstream (searchQuery), so drop any
  // typed text and chips here to match, keyed off the selection changing.
  let lastSelection = $selection
  $: if ($selection !== lastSelection) {
    lastSelection = $selection
    text = ''
    chips = []
  }

  const searchHint = shortcutLabel('mod+f')
  const dispatch = createEventDispatcher<{ search: string; filter: SearchFilter }>()

  // chip fields and the aliases that produce them ("sender:" -> from).
  type ChipField = 'from' | 'to' | 'subject' | 'has' | 'before' | 'after'
  const alias: Record<string, ChipField> = {
    from: 'from',
    sender: 'from',
    to: 'to',
    subject: 'subject',
    has: 'has',
    before: 'before',
    after: 'after',
  }
  const keywordList = ['from', 'sender', 'to', 'subject', 'has', 'before', 'after']

  interface Chip {
    field: ChipField
    value: string
  }

  let chips: Chip[] = []
  // free text (and the keyword currently being typed) live in the input.
  let text = value
  let inputEl: HTMLInputElement
  let timer: ReturnType<typeof setTimeout> | undefined
  let showDate = false
  let afterDate = ''
  let beforeDate = ''

  const day = 86400

  // fieldLabel is the human label shown on a chip.
  function fieldLabel(field: ChipField): string {
    return $t(`messageList.search.chip.${field}`)
  }

  // chipText renders a chip's value; has:attachment has no free value.
  function chipText(chip: Chip): string {
    if (chip.field === 'has') {
      return $t('messageList.search.chip.hasAttachment')
    }
    return `${fieldLabel(chip.field)}: ${chip.value}`
  }

  // parseToken turns a "keyword:value" token into a chip, or null if it is not a
  // recognized, valid keyword token.
  function parseToken(token: string): Chip | null {
    const at = token.indexOf(':')
    if (at <= 0) {
      return null
    }
    const field = alias[token.slice(0, at).toLowerCase()]
    const raw = token.slice(at + 1).trim()
    if (!field || raw === '') {
      return null
    }
    if (field === 'has') {
      return raw.toLowerCase().startsWith('attach') ? { field, value: 'attachment' } : null
    }
    if ((field === 'before' || field === 'after') && Number.isNaN(Date.parse(raw))) {
      return null
    }
    return { field, value: raw }
  }

  // addChip stores a chip, replacing any existing chip of the same field so each
  // constraint appears once.
  function addChip(chip: Chip): void {
    chips = [...chips.filter((c) => c.field !== chip.field), chip]
    emit()
  }

  function removeChip(index: number): void {
    chips = chips.filter((_, i) => i !== index)
    emit()
  }

  // buildFilter maps the chips onto the structured SearchFilter.
  function buildFilter(): SearchFilter {
    const f: SearchFilter = { ...emptyFilter }
    for (const c of chips) {
      if (c.field === 'from') {
        f.from = c.value
      } else if (c.field === 'to') {
        f.to = c.value
      } else if (c.field === 'subject') {
        f.subject = c.value
      } else if (c.field === 'has') {
        f.hasAttachment = true
      } else if (c.field === 'after') {
        f.afterUnix = Math.floor(Date.parse(c.value) / 1000) || 0
      } else if (c.field === 'before') {
        // include the whole "before" day.
        f.beforeUnix = Math.floor(Date.parse(c.value) / 1000) + day - 1 || 0
      }
    }
    return f
  }

  // emit debounces and dispatches the current free text and chip filter.
  function emit(): void {
    clearTimeout(timer)
    timer = setTimeout(() => {
      dispatch('search', text.trim())
      dispatch('filter', buildFilter())
    }, 180)
  }

  // the token currently being typed (after the last space), used for autocomplete.
  $: partial = text.slice(text.lastIndexOf(' ') + 1)
  $: suggestions =
    partial.length > 0 && !partial.includes(':')
      ? keywordList.filter((k) => k.startsWith(partial.toLowerCase()) && k !== partial.toLowerCase())
      : []

  function onInput(event: Event): void {
    text = (event.currentTarget as HTMLInputElement).value
    // a trailing space commits a completed keyword token immediately.
    if (text.endsWith(' ')) {
      commitTrailingToken()
    }
    emit()
  }

  // commitTrailingToken pulls a finished "keyword:value" token out of the input
  // into a chip, leaving the remaining free text behind.
  function commitTrailingToken(): void {
    const parts = text.split(/\s+/).filter(Boolean)
    if (parts.length === 0) {
      return
    }
    const last = parts[parts.length - 1]
    const chip = parseToken(last)
    if (chip) {
      addChip(chip)
      text = parts.slice(0, -1).join(' ')
      text = text ? `${text} ` : ''
    }
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter') {
      const chip = parseToken(partial)
      if (chip) {
        event.preventDefault()
        addChip(chip)
        text = text.slice(0, text.lastIndexOf(' ') + 1).replace(/\s*$/, '')
        text = text ? `${text} ` : ''
        emit()
      }
      return
    }
    if (event.key === 'Tab' && suggestions.length > 0) {
      event.preventDefault()
      applySuggestion(suggestions[0])
      return
    }
    if (event.key === 'Backspace' && text === '' && chips.length > 0) {
      removeChip(chips.length - 1)
    }
  }

  // applySuggestion replaces the partial keyword with "keyword:" and refocuses.
  async function applySuggestion(keyword: string): Promise<void> {
    text = `${text.slice(0, text.lastIndexOf(' ') + 1)}${keyword}:`
    await tick()
    inputEl?.focus()
  }

  function clearAll(): void {
    clearTimeout(timer)
    text = ''
    chips = []
    dispatch('search', '')
    dispatch('filter', { ...emptyFilter })
  }

  // applyDates turns the popover's picked dates into before/after chips.
  function applyDates(): void {
    if (afterDate) {
      addChip({ field: 'after', value: afterDate })
    }
    if (beforeDate) {
      addChip({ field: 'before', value: beforeDate })
    }
    showDate = false
  }

  $: hasContent = text !== '' || chips.length > 0
</script>

<div class="bar">
  <div class="search">
    <IconSearch size={15} stroke={1.6} class="search-icon" />
    <div class="field">
      {#each chips as chip, i (chip.field + chip.value)}
        <span class="chip">
          <span class="chip-text">{chipText(chip)}</span>
          <button type="button" class="chip-x" aria-label={$t('messageList.search.removeChip')} on:click={() => removeChip(i)}>
            <IconX size={11} stroke={2} />
          </button>
        </span>
      {/each}
      <input
        type="text"
        bind:this={inputEl}
        placeholder={chips.length === 0 ? $t('messageList.search.placeholder') : ''}
        aria-label={$t('messageList.search.placeholder')}
        value={text}
        on:input={onInput}
        on:keydown={onKeydown}
      />
    </div>
    {#if hasContent}
      <button type="button" class="clear" aria-label={$t('messageList.search.clearSearch')} on:click={clearAll}>
        <IconX size={14} stroke={1.8} />
      </button>
    {:else if $prefs.showShortcutHints}
      <kbd class="hint">{searchHint}</kbd>
    {/if}

    {#if suggestions.length > 0}
      <div class="autocomplete" role="listbox">
        {#each suggestions as s (s)}
          <button type="button" class="ac-opt" role="option" aria-selected="false" on:click={() => applySuggestion(s)}>
            <span class="ac-key">{s}:</span>
            <span class="ac-desc">{$t(`messageList.search.suggest.${s}`)}</span>
          </button>
        {/each}
      </div>
    {/if}
  </div>

  <div class="filter-wrap">
    <button
      type="button"
      class="filter-btn"
      aria-label={$t('messageList.search.filterByDate')}
      aria-expanded={showDate}
      title={$t('messageList.search.filterByDate')}
      on:click={() => (showDate = !showDate)}
    >
      <IconCalendar size={16} stroke={1.7} />
    </button>

    {#if showDate}
      <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
      <div class="scrim" on:click={() => (showDate = false)}></div>
      <div class="menu" role="menu">
        <span class="menu-label">{$t('messageList.search.dateRange')}</span>
        <div class="date">
          <span>{$t('messageList.search.chip.after')}</span>
          <div class="date-picker">
            <DateTimePicker mode="date" bind:value={afterDate} />
          </div>
        </div>
        <div class="date">
          <span>{$t('messageList.search.chip.before')}</span>
          <div class="date-picker">
            <DateTimePicker mode="date" bind:value={beforeDate} />
          </div>
        </div>
        <div class="menu-actions">
          <button type="button" class="primary" on:click={applyDates}>{$t('messageList.search.apply')}</button>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .bar {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .search {
    position: relative;
    flex: 1;
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0 var(--space-3);
    min-height: var(--control-height);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-tertiary);
  }

  .search:focus-within {
    border-color: var(--accent);
  }

  .field {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-1);
    padding: 3px 0;
  }

  .chip {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: 1px var(--space-1) 1px var(--space-2);
    border-radius: var(--radius-control);
    background: var(--selection-bg);
    color: var(--text-primary);
    font-size: var(--fz-meta);
    white-space: nowrap;
  }

  .chip-x {
    display: inline-flex;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: 1px;
    border-radius: var(--radius-control);
  }
  .chip-x:hover {
    color: var(--text-primary);
  }

  input[type='text'] {
    flex: 1;
    min-width: 80px;
    border: none;
    background: transparent;
    outline: none;
    font-size: var(--fz-list);
    color: var(--text-primary);
  }

  .clear {
    display: inline-flex;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: 2px;
    border-radius: var(--radius-control);
  }
  .clear:hover {
    color: var(--text-primary);
  }

  .hint {
    flex-shrink: 0;
    padding: 1px var(--space-2);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-tertiary);
    font-family: var(--font-mono);
    font-size: var(--fz-meta);
    line-height: 1.4;
  }

  .autocomplete {
    position: absolute;
    top: calc(100% + var(--space-1));
    left: 0;
    right: 0;
    z-index: 41;
    padding: var(--space-1);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
  }

  .ac-opt {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    width: 100%;
    border: none;
    background: transparent;
    cursor: pointer;
    text-align: left;
    padding: var(--space-2);
    border-radius: var(--radius-control);
  }
  .ac-opt:hover {
    background: var(--surface-hover);
  }
  .ac-key {
    font-family: var(--font-mono);
    font-size: var(--fz-label);
    color: var(--accent);
  }
  .ac-desc {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .filter-wrap {
    position: relative;
    flex-shrink: 0;
  }

  .filter-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: var(--control-height);
    width: var(--control-height);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-secondary);
    cursor: pointer;
  }
  .filter-btn:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }

  .scrim {
    position: fixed;
    inset: 0;
    z-index: 40;
  }

  .menu {
    position: absolute;
    top: calc(100% + var(--space-1));
    right: 0;
    z-index: 41;
    width: 232px;
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
  }

  .menu-label {
    display: block;
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
    margin: 0 0 var(--space-2);
  }

  .date {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-2);
    padding-bottom: var(--space-2);
  }
  .date span {
    font-size: var(--fz-label);
    color: var(--text-secondary);
  }
  .date-picker {
    width: 132px;
  }

  .menu-actions {
    display: flex;
    justify-content: flex-end;
    padding-top: var(--space-1);
  }

  .primary {
    padding: var(--space-1) var(--space-3);
    border-radius: var(--radius-control);
    border: none;
    background: var(--accent);
    color: var(--accent-fg);
    font-size: var(--fz-label);
    cursor: pointer;
  }
</style>
