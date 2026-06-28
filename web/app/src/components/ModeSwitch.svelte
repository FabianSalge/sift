<script lang="ts">
  export type Mode = 'contrast' | 'explain' | 'sandbox'

  let { mode, onchange }: { mode: Mode; onchange: (m: Mode) => void } = $props()

  const modes: { id: Mode; label: string }[] = [
    { id: 'contrast', label: 'Contrast' },
    { id: 'explain', label: 'Explain' },
    { id: 'sandbox', label: 'Sandbox' },
  ]
</script>

<div class="seg" role="tablist">
  {#each modes as m (m.id)}
    <button class="opt" class:on={mode === m.id} role="tab" aria-selected={mode === m.id} onclick={() => onchange(m.id)}>
      {m.label}
    </button>
  {/each}
</div>

<style>
  .seg {
    display: inline-flex;
    padding: 3px;
    gap: 2px;
    border: 1px solid var(--line);
    border-radius: var(--r-md);
    background: var(--panel);
  }
  .opt {
    appearance: none;
    border: none;
    background: none;
    color: var(--ink-dim);
    font-family: var(--font-sans);
    font-size: 12.5px;
    font-weight: 500;
    padding: 6px 16px;
    border-radius: 7px;
    cursor: pointer;
    transition:
      background 0.12s ease,
      color 0.12s ease;
  }
  .opt:hover {
    color: var(--ink);
  }
  .opt.on {
    color: #fff;
    background: linear-gradient(180deg, color-mix(in oklab, var(--gpu) 92%, white), var(--gpu));
    box-shadow: 0 1px 0 rgba(0, 0, 0, 0.3);
  }
</style>
