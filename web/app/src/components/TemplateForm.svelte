<script lang="ts">
  import type { Precision } from '../lib/types'
  import type { WorkloadTemplate } from '../lib/templates'
  import { blankTemplate } from '../lib/templates'

  let {
    existing,
    onadd,
    oncancel,
  }: { existing: string[]; onadd: (t: WorkloadTemplate) => void; oncancel: () => void } = $props()

  const PRECISIONS: Precision[] = ['bf16', 'fp16', 'fp8', 'fp4', 'int8']
  let t = $state<WorkloadTemplate>(blankTemplate('my-job'))

  const idTaken = $derived(existing.includes(t.id))
  const valid = $derived(t.id.trim().length > 0 && !idTaken)

  function togglePrecision(p: Precision) {
    const w = t.workload
    w.requiredPrecisions = w.requiredPrecisions.includes(p)
      ? w.requiredPrecisions.filter((x) => x !== p)
      : [...w.requiredPrecisions, p]
  }
</script>

<div class="form">
  <div class="field col">
    <span class="k">name {#if idTaken}<b class="bad">taken</b>{/if}</span>
    <input class="mono name" type="text" bind:value={t.id} maxlength="16" spellcheck="false" />
  </div>

  <div class="field">
    <span class="k">kind</span>
    <div class="seg">
      <button class:on={t.workload.kind === 'train'} onclick={() => (t.workload.kind = 'train')}>train</button>
      <button class:on={t.workload.kind === 'infer'} onclick={() => (t.workload.kind = 'infer')}>infer</button>
    </div>
  </div>

  <div class="field col">
    <span class="k">min memory <b class="mono">{t.workload.minMemoryGB} GB</b></span>
    <input type="range" min="0" max="256" step="8" bind:value={t.workload.minMemoryGB} />
  </div>

  <div class="field col">
    <span class="k">required precisions</span>
    <div class="chips">
      {#each PRECISIONS as p (p)}
        <button class="pchip mono" class:on={t.workload.requiredPrecisions.includes(p)} onclick={() => togglePrecision(p)}>{p}</button>
      {/each}
    </div>
  </div>

  <div class="field col">
    <span class="k">device count <b class="mono">×{t.workload.deviceCount}</b></span>
    <input type="range" min="1" max="8" step="1" bind:value={t.workload.deviceCount} />
  </div>

  {#if t.workload.deviceCount > 1}
    <label class="field check">
      <input type="checkbox" bind:checked={t.workload.sameIsland} />
      <span class="k">same island (gang stays whole)</span>
    </label>
  {/if}

  <div class="field col">
    <span class="k">cost weight <b class="mono">{t.workload.costWeight.toFixed(1)}</b></span>
    <input type="range" min="0" max="1" step="0.1" bind:value={t.workload.costWeight} />
  </div>

  <div class="field col">
    <span class="k">duration <b class="mono">{t.durationS}s</b></span>
    <input type="range" min="10" max="240" step="5" bind:value={t.durationS} />
  </div>

  <div class="actions">
    <button class="ok" disabled={!valid} onclick={() => onadd($state.snapshot(t))}>add template</button>
    <button class="no" onclick={oncancel}>cancel</button>
  </div>
</div>

<style>
  .form {
    border: 1px dashed var(--line-strong);
    border-radius: var(--r-md);
    padding: 12px;
  }
  .field { margin-bottom: 12px; }
  .field.col { display: flex; flex-direction: column; gap: 7px; }
  .field:not(.col) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }
  .k { font-size: 11.5px; color: var(--ink-dim); }
  .k b { color: var(--ink); font-weight: 500; }
  .k .bad { color: var(--reject); }

  input.name {
    background: var(--panel);
    border: 1px solid var(--line);
    border-radius: 6px;
    color: var(--ink);
    font-size: 12px;
    padding: 6px 9px;
  }

  .seg { display: inline-flex; padding: 2px; gap: 2px; border: 1px solid var(--line); border-radius: 7px; }
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
  .seg button.on { background: var(--panel-2); color: var(--ink); box-shadow: inset 0 0 0 1px var(--line-strong); }

  .chips { display: flex; flex-wrap: wrap; gap: 5px; }
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

  input[type='range'] { width: 100%; accent-color: var(--gpu); cursor: pointer; }
  .check { cursor: pointer; }
  .check input { accent-color: var(--gpu); }

  .actions { display: flex; gap: 8px; margin-top: 4px; }
  .actions button {
    appearance: none;
    border-radius: 7px;
    font-size: 11px;
    padding: 6px 12px;
    cursor: pointer;
  }
  .ok {
    border: 1px solid color-mix(in oklab, var(--bound) 50%, transparent);
    background: color-mix(in oklab, var(--bound) 14%, transparent);
    color: var(--bound);
  }
  .ok:disabled { opacity: 0.4; cursor: not-allowed; }
  .no { border: 1px solid var(--line); background: none; color: var(--ink-faint); }
</style>
