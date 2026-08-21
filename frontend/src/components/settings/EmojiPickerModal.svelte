<script lang="ts">
  // the icon picker behind the profile form: a grid of emoji grouped by kind,
  // over the form rather than inside it, so a long grid never pushes the rest
  // of the form off the screen.
  import { createEventDispatcher } from 'svelte'
  import { IconX } from '@tabler/icons-svelte'
  import Modal from '../common/Modal.svelte'
  import { t } from '../../lib/i18n'

  /** The icon currently chosen, so the grid can mark it. */
  export let value = ''

  const dispatch = createEventDispatcher<{ select: string; close: void }>()

  // a label you recognise at a glance in the status bar, so it is a fixed set
  // rather than a text field to paste anything into.
  const groups: { key: string; icons: string[] }[] = [
    {
      key: 'people',
      icons: ['🧑', '👤', '👥', '🧑‍💻', '🧑‍🎓', '🧑‍🏫', '🧑‍⚕️', '🧑‍🔧', '🧑‍🍳', '🧑‍🎨'],
    },
    {
      key: 'places',
      icons: ['🏠', '🏡', '🏢', '🏫', '🏥', '🏦', '🏭', '🏛️', '⛺', '🌆'],
    },
    {
      key: 'work',
      icons: [
        '💼', '📁', '🗂️', '📋', '📌', '📎', '🖇️', '📐', '📏', '✏️',
        '🖊️', '🖋️', '📝', '📚', '📖', '🎓', '🧾', '💳', '💰', '📊',
        '📈', '📉', '🗓️', '⏰', '⌛', '🔔', '📣', '🗣️', '🤝', '🧠',
      ],
    },
    {
      key: 'mail',
      icons: ['✉️', '📮', '📨', '📬', '📭', '📦', '📡', '☎️', '📱', '💬'],
    },
    {
      key: 'tools',
      icons: [
        '🛠️', '🔧', '🔩', '⚙️', '🧰', '🔬', '🧪', '⚗️', '🧬', '🔭',
        '💡', '🔌', '🔋', '🖥️', '💻', '⌨️', '🖱️', '🖨️', '💾', '🗄️',
      ],
    },
    {
      key: 'arts',
      icons: [
        '🎨', '🖌️', '🎭', '🎬', '📷', '📹', '🎧', '🎵', '🎸', '🎹',
        '🥁', '🎤', '🎮', '🕹️', '🎲', '♟️', '🧩', '🎯', '🎳', '🏆',
      ],
    },
    {
      key: 'sport',
      icons: [
        '⚽', '🏀', '🏈', '🎾', '🏐', '🏓', '🥊', '🚴', '🏃', '🧗',
        '🏊', '⛷️', '🏕️', '🥾', '🚵', '🛹', '🛼', '🪁', '🎣', '🧘',
      ],
    },
    {
      key: 'travel',
      icons: [
        '✈️', '🚀', '🚗', '🚌', '🚂', '🚢', '⛵', '🛵', '🚲', '🗺️',
        '🧳', '🏝️', '🏔️', '🌋', '🗽', '🎡', '🎢', '🌉', '⛱️', '🧭',
      ],
    },
    {
      key: 'nature',
      icons: [
        '🌍', '🌙', '⭐', '🌟', '☀️', '⛅', '🌈', '❄️', '🔥', '💧',
        '🌊', '🌱', '🌳', '🌵', '🍀', '🌸', '🌻', '🍁', '🍄', '🪴',
      ],
    },
    {
      key: 'animals',
      icons: [
        '🐢', '🐝', '🐞', '🦊', '🐻', '🐼', '🐨', '🐧', '🦉', '🦅',
        '🐬', '🐙', '🦀', '🐈', '🐕', '🐎', '🦋', '🐳', '🦔', '🦕',
      ],
    },
    {
      key: 'food',
      icons: [
        '🥔', '☕', '🍵', '🍺', '🍷', '🥂', '🍎', '🍌', '🍓', '🍇',
        '🍕', '🍔', '🌮', '🍜', '🍣', '🥐', '🥗', '🍪', '🎂', '🍫',
      ],
    },
    {
      key: 'symbols',
      icons: [
        '❤️', '💙', '💚', '💛', '💜', '🖤', '🤍', '🧡', '💫', '✨',
        '🔒', '🔑', '🛡️', '⚡', '♻️', '⚓', '🧿', '🔮', '🎁', '🏁',
      ],
    },
  ]

  function pick(icon: string): void {
    dispatch('select', icon)
    dispatch('close')
  }
</script>

<Modal title={$t('iconPicker.title')} size="medium" on:close={() => dispatch('close')}>
  <button type="button" class="clear" class:on={value === ''} on:click={() => pick('')}>
    <IconX size={14} stroke={2} />
    {$t('profiles.icon.none')}
  </button>

  {#each groups as group (group.key)}
    <h4>{$t(`iconPicker.group.${group.key}`)}</h4>
    <div class="grid" role="radiogroup" aria-label={$t(`iconPicker.group.${group.key}`)}>
      {#each group.icons as icon (icon)}
        <button
          type="button"
          class="choice"
          class:on={value === icon}
          role="radio"
          aria-checked={value === icon}
          aria-label={icon}
          on:click={() => pick(icon)}
        >
          {icon}
        </button>
      {/each}
    </div>
  {/each}
</Modal>

<style>
  .clear {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-secondary);
    font-size: var(--fz-label);
    cursor: var(--cursor-action);
  }
  .clear:hover {
    background: var(--surface-hover);
    color: var(--text-primary);
  }
  .clear.on {
    border-color: var(--accent);
    color: var(--text-primary);
  }

  h4 {
    margin: var(--space-4) 0 var(--space-2);
    font-size: var(--fz-label);
    font-weight: var(--fw-semibold);
    color: var(--text-tertiary);
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(34px, 1fr));
    gap: var(--space-1);
  }

  .choice {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 34px;
    border: var(--hairline) solid transparent;
    border-radius: var(--radius-control);
    background: transparent;
    font-size: var(--fz-body);
    line-height: 1;
    cursor: var(--cursor-action);
  }
  .choice:hover {
    background: var(--surface-hover);
  }
  .choice.on {
    border-color: var(--accent);
    background: var(--selection-bg);
  }
</style>
