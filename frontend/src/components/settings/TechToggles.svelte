<script lang="ts">
  // toggles for the per-row technical-info display. each switch independently
  // shows or hides one piece of info in the list and detail header. the auth
  // toggle controls a placeholder badge: spf/dkim/dmarc data is not parsed yet
  // (documented backend follow-up), so the badge shows a neutral "not available"
  // state when on.
  import { prefs, setToggle } from '../../stores/prefs'
  import ToggleSwitch from '../common/ToggleSwitch.svelte'

  const items: { key: 'showMailboxBadge' | 'showDateTime' | 'showPgp' | 'showAuth'; label: string; note: string }[] = [
    { key: 'showMailboxBadge', label: 'Mailbox badge', note: 'Account and folder on each message' },
    { key: 'showDateTime', label: 'Date and time', note: 'Timestamp on rows and in the header' },
    { key: 'showPgp', label: 'PGP status', note: 'Signed or encrypted indicator' },
    { key: 'showAuth', label: 'Auth status', note: 'SPF/DKIM/DMARC (not available yet)' },
  ]
</script>

<div class="toggles">
  <span class="group-label">Technical info</span>
  {#each items as item (item.key)}
    <div class="toggle">
      <span class="text">
        <span class="name">{item.label}</span>
        <span class="note">{item.note}</span>
      </span>
      <ToggleSwitch
        checked={$prefs[item.key]}
        label={item.label}
        on:change={(e) => setToggle(item.key, e.detail)}
      />
    </div>
  {/each}
</div>

<style>
  .toggles {
    padding: var(--space-3) 0;
  }

  .group-label {
    display: block;
    font-size: var(--fz-body);
    color: var(--text-primary);
    margin-bottom: var(--space-2);
  }

  .toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    padding: var(--space-2) 0;
    cursor: pointer;
  }

  .text {
    display: flex;
    flex-direction: column;
  }

  .name {
    font-size: var(--fz-label);
    color: var(--text-primary);
  }

  .note {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }
</style>
