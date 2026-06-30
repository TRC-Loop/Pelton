<script lang="ts">
  // a tiny, non-interactive mockup of the three-pane app, built entirely from
  // theme tokens. it is used in the onboarding accent and density steps as a live
  // preview: the selected row and the link bar use the accent, and the list rows
  // use the density-driven row padding, so picking a theme, accent or density
  // updates this in real time exactly like the real app would.
</script>

<div class="mock" aria-hidden="true">
  <div class="sidebar">
    <span class="bar w70"></span>
    <span class="bar w50"></span>
    <span class="bar w60 mt"></span>
    <span class="bar w40"></span>
    <span class="bar w55"></span>
  </div>

  <div class="list">
    {#each [0, 1, 2, 3] as i}
      <div class="row" class:selected={i === 1}>
        <span class="avatar"></span>
        <span class="lines">
          <span class="bar w80"></span>
          <span class="bar w60"></span>
        </span>
      </div>
    {/each}
  </div>

  <div class="detail">
    <span class="bar w70 tall"></span>
    <span class="bar w40 mt"></span>
    <span class="bar w90 mt2"></span>
    <span class="bar w85"></span>
    <span class="bar w50"></span>
    <span class="bar link mt2 w60"></span>
  </div>
</div>

<style>
  .mock {
    display: grid;
    grid-template-columns: 28% 34% 1fr;
    height: 184px;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    overflow: hidden;
    background: var(--surface-raised);
  }

  .sidebar {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-3);
    background: var(--surface-base);
    border-right: var(--hairline) solid var(--border-subtle);
  }

  .list {
    display: flex;
    flex-direction: column;
    background: var(--surface-raised);
    border-right: var(--hairline) solid var(--border-subtle);
    overflow: hidden;
  }

  /* rows use the live density padding so the density step preview is honest. */
  .row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--row-pad-y) var(--row-pad-x);
    border-bottom: var(--hairline) solid var(--border-subtle);
  }

  .row.selected {
    background: var(--selection-bg);
  }

  .avatar {
    width: 18px;
    height: 18px;
    border-radius: 999px;
    background: var(--surface-sunken);
    flex-shrink: 0;
  }

  .lines {
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
    min-width: 0;
  }

  .detail {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-4);
  }

  .bar {
    height: 7px;
    border-radius: 999px;
    background: var(--text-tertiary);
    opacity: 0.32;
  }

  .bar.tall {
    height: 11px;
    opacity: 0.5;
  }

  .bar.link {
    background: var(--link);
    opacity: 0.85;
  }

  .mt {
    margin-top: var(--space-1);
  }

  .mt2 {
    margin-top: var(--space-3);
  }

  .w40 { width: 40%; }
  .w50 { width: 50%; }
  .w55 { width: 55%; }
  .w60 { width: 60%; }
  .w70 { width: 70%; }
  .w80 { width: 80%; }
  .w85 { width: 85%; }
  .w90 { width: 90%; }
</style>
