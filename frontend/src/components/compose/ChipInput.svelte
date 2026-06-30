<script lang="ts">
  // a recipient field that renders addresses as removable chips. it keeps the
  // store contract simple by emitting a comma-separated string (which the send
  // path already parses); chips are just the display layer. typing a comma,
  // semicolon, Enter or Tab commits the current token; Backspace on an empty
  // input removes the last chip.
  import { createEventDispatcher } from 'svelte'
  import { IconX } from '@tabler/icons-svelte'

  export let value = ''
  export let label: string
  export let id: string

  const dispatch = createEventDispatcher<{ change: string }>()

  let draft = ''

  $: chips = value
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)

  function emit(next: string[]): void {
    dispatch('change', next.join(', '))
  }

  function commit(): void {
    const token = draft.trim().replace(/[,;]+$/, '').trim()
    if (token) {
      emit([...chips, token])
    }
    draft = ''
  }

  function removeAt(index: number): void {
    emit(chips.filter((_, i) => i !== index))
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' || event.key === ',' || event.key === ';' || event.key === 'Tab') {
      if (draft.trim()) {
        event.preventDefault()
        commit()
      }
    } else if (event.key === 'Backspace' && draft === '' && chips.length > 0) {
      event.preventDefault()
      removeAt(chips.length - 1)
    }
  }

  // a light validity hint: a chip without an @ is probably incomplete.
  function looksValid(addr: string): boolean {
    return addr.includes('@')
  }
</script>

<div class="chips" role="group" aria-label={label}>
  {#each chips as chip, index (chip + index)}
    <span class="chip" class:invalid={!looksValid(chip)}>
      <span class="chip-text">{chip}</span>
      <button type="button" class="chip-x" aria-label={`Remove ${chip}`} on:click={() => removeAt(index)}>
        <IconX size={11} stroke={2} />
      </button>
    </span>
  {/each}
  <input
    {id}
    type="text"
    bind:value={draft}
    placeholder={chips.length === 0 ? 'name@example.com, …' : ''}
    on:keydown={onKeydown}
    on:blur={commit}
  />
</div>

<style>
  .chips {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    flex: 1;
    min-width: 0;
  }

  .chip {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    padding: 1px var(--space-1) 1px var(--space-2);
    border-radius: 999px;
    background: var(--surface-sunken);
    border: var(--hairline) solid var(--border-subtle);
    font-size: var(--fz-meta);
    max-width: 100%;
  }

  .chip.invalid {
    border-color: var(--warning);
  }

  .chip-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .chip-x {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    color: var(--text-tertiary);
    cursor: pointer;
    padding: 1px;
    border-radius: 999px;
    flex-shrink: 0;
  }

  .chip-x:hover {
    color: var(--danger);
  }

  input {
    flex: 1;
    min-width: 80px;
    border: none;
    background: transparent;
    outline: none;
    font-size: var(--fz-body);
  }
</style>
