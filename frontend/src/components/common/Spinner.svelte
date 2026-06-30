<script lang="ts">
  // a small, accessible loading indicator. label describes what is loading.
  export let label: string = 'Loading'
  export let inline: boolean = false
</script>

<div class="spinner-wrap" class:inline role="status" aria-live="polite">
  <span class="spinner" aria-hidden="true"></span>
  <span class="label">{label}</span>
</div>

<style>
  .spinner-wrap {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-3);
    padding: var(--space-6);
    color: var(--text-tertiary);
    font-size: var(--fz-label);
  }

  .spinner-wrap.inline {
    flex-direction: row;
    padding: var(--space-3);
    gap: var(--space-2);
  }

  .spinner {
    width: 18px;
    height: 18px;
    border: 2px solid var(--border-default);
    border-top-color: var(--accent);
    border-radius: 999px;
    animation: spin 0.7s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
