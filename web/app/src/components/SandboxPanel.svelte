<script lang="ts">
  import type { Workload, Trace, Precision } from '../lib/types'
  import { modelName } from '../lib/types'

  let { workload = $bindable(), trace }: { workload: Workload; trace: Trace | null } = $props()

  const PRECISIONS: Precision[] = ['bf16', 'fp16', 'fp8', 'fp4', 'int8']

  function togglePrecision(p: Precision) {
    workload.requiredPrecisions = workload.requiredPrecisions.includes(p)
      ? workload.requiredPrecisions.filter((x) => x !== p)
      : [...workload.requiredPrecisions, p]
  }

  const feasible = $derived(trace ? trace.verdicts.filter((v) => v.feasible).length : 0)
  const rejected = $derived(trace ? trace.verdicts.filter((v) => !v.feasible).length : 0)
  const boundLabel = $derived(
    !trace?.bound || trace.bound.length === 0
      ? '—'
      : trace.bound.length > 1
        ? `${modelName(trace.bound[0])} ×${trace.bound.length}`
        : modelName(trace.bound[0]),
  )
</script>

<aside class="panel">
  <div class="label">author a workload</div>

  <div class="field">
    <span class="k">kind</span>
    <div class="seg">
      <button class:on={workload.kind === 'train'} onclick={() => (workload.kind = 'train')}>train</button>
      <button class:on={workload.kind === 'infer'} onclick={() => (workload.kind = 'infer')}>infer</button>
    </div>
  </div>

  <div class="field col">
    <span class="k">min memory <b class="mono">{workload.minMemoryGB} GB</b></span>
    <input type="range" min="0" max="256" step="8" bind:value={workload.minMemoryGB} />
  </div>

  <div class="field col">
    <span class="k">required precisions</span>
    <div class="chips">
      {#each PRECISIONS as p (p)}
        <button class="pchip mono" class:on={workload.requiredPrecisions.includes(p)} onclick={() => togglePrecision(p)}>{p}</button>
      {/each}
    </div>
  </div>

  <div class="field col">
    <span class="k">device count <b class="mono">×{workload.deviceCount}</b></span>
    <input type="range" min="1" max="8" step="1" bind:value={workload.deviceCount} />
  </div>

  {#if workload.deviceCount > 1}
    <label class="field check">
      <input type="checkbox" bind:checked={workload.sameIsland} />
      <span class="k">same island (gang stays whole)</span>
    </label>
  {/if}

  <div class="field col">
    <span class="k">cost weight <b class="mono">{workload.costWeight.toFixed(1)}</b></span>
    <input type="range" min="0" max="1" step="0.1" bind:value={workload.costWeight} />
  </div>

  <div class="result">
    {#if trace?.err}
      <span class="mk reject">no fit</span>
      <span class="rtext">{trace.err}</span>
    {:else if trace}
      <span class="mk bound">sift binds</span>
      <span class="rtext mono">{boundLabel}{trace.island >= 0 ? ` · island ${trace.island}` : ''}</span>
    {/if}
  </div>
  {#if trace}
    <div class="counts mono">{feasible} feasible · {rejected} rejected of {trace.verdicts.length}</div>
  {/if}
</aside>

<style>
  .panel {
    width: 244px;
    flex: none;
    background: var(--panel);
    border: 1px solid var(--line);
    border-radius: var(--r-lg);
    padding: 14px;
    position: sticky;
    top: 16px;
  }
  .label {
    margin-bottom: 13px;
  }

  .field {
    margin-bottom: 13px;
  }
  .field.col {
    display: flex;
    flex-direction: column;
    gap: 7px;
  }
  .field:not(.col) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }
  .k {
    font-size: 11.5px;
    color: var(--ink-dim);
  }
  .k b {
    color: var(--ink);
    font-weight: 500;
  }

  .seg {
    display: inline-flex;
    padding: 2px;
    gap: 2px;
    border: 1px solid var(--line);
    border-radius: 7px;
  }
  .seg button {
    appearance: none;
    border: none;
    background: none;
    color: var(--ink-dim);
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 4px 12px;
    border-radius: 5px;
    cursor: pointer;
  }
  .seg button.on {
    background: var(--panel-2);
    color: var(--ink);
    box-shadow: inset 0 0 0 1px var(--line-strong);
  }

  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
  }
  .pchip {
    appearance: none;
    border: 1px solid var(--line);
    background: var(--panel);
    color: var(--ink-faint);
    font-size: 10px;
    padding: 4px 9px;
    border-radius: 6px;
    cursor: pointer;
  }
  .pchip.on {
    color: var(--ink);
    border-color: color-mix(in oklab, var(--gpu) 55%, transparent);
    background: color-mix(in oklab, var(--gpu) 16%, transparent);
  }

  input[type='range'] {
    width: 100%;
    accent-color: var(--gpu);
    cursor: pointer;
  }

  .check {
    cursor: pointer;
  }
  .check input {
    accent-color: var(--gpu);
  }

  .result {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 16px;
    padding-top: 13px;
    border-top: 1px solid var(--line);
  }
  .rtext {
    font-size: 12px;
    color: var(--ink);
  }
  .mk {
    font-size: 9px;
    letter-spacing: 0.04em;
    padding: 2px 7px;
    border-radius: 5px;
    text-transform: uppercase;
    white-space: nowrap;
  }
  .mk.bound {
    color: var(--bound);
    background: color-mix(in oklab, var(--bound) 15%, transparent);
  }
  .mk.reject {
    color: var(--reject);
    background: color-mix(in oklab, var(--reject) 15%, transparent);
  }
  .counts {
    margin-top: 8px;
    font-size: 10px;
    color: var(--ink-faint);
  }
</style>
