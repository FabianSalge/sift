<script lang="ts">
  import type { WorkloadTemplate } from '../lib/templates'
  import TemplateForm from './TemplateForm.svelte'

  let {
    templates = $bindable(),
    onburst,
  }: { templates: WorkloadTemplate[]; onburst: (t: WorkloadTemplate, n: number) => void } = $props()

  let adding = $state(false)

  const summary = (t: WorkloadTemplate): string =>
    [
      t.workload.kind,
      `≥${t.workload.minMemoryGB}GB`,
      ...t.workload.requiredPrecisions,
      t.workload.deviceCount > 1 ? `×${t.workload.deviceCount}` : null,
      t.workload.sameIsland ? 'same-island' : null,
      `~${t.durationS}s`,
    ]
      .filter(Boolean)
      .join(' · ')

  function add(t: WorkloadTemplate) {
    templates.push(t)
    adding = false
  }
  function remove(id: string) {
    templates = templates.filter((t) => t.id !== id)
  }
</script>

<section class="sec">
  <div class="label">workloads</div>
  {#each templates as t, i (t.id)}
    <div class="card">
      <div class="hd">
        <span class="name mono">{t.id}</span>
        <button class="x" title="delete template" onclick={() => remove(t.id)}>×</button>
      </div>
      <div class="sum">{summary(t)}</div>
      <label class="rate">
        <span class="k">ambient {t.ratePerMin > 0 ? `${t.ratePerMin.toFixed(1)}/min` : 'off'}</span>
        <input type="range" min="0" max="10" step="0.2" bind:value={templates[i].ratePerMin} />
      </label>
      <div class="bursts">
        <button onclick={() => onburst(t, 1)}>burst ×1</button>
        <button onclick={() => onburst(t, 5)}>burst ×5</button>
      </div>
    </div>
  {/each}

  {#if adding}
    <TemplateForm existing={templates.map((t) => t.id)} onadd={add} oncancel={() => (adding = false)} />
  {:else}
    <button class="new" onclick={() => (adding = true)}>+ new workload</button>
  {/if}
</section>

<style>
  .sec { display: flex; flex-direction: column; gap: 10px; }

  .card {
    border: 1px solid var(--line);
    border-radius: var(--r-md);
    background: var(--panel);
    padding: 10px 12px;
  }
  .hd { display: flex; align-items: center; justify-content: space-between; }
  .name { font-size: 12px; font-weight: 600; }
  .x {
    appearance: none;
    border: none;
    background: none;
    color: var(--ink-faint);
    font-size: 14px;
    cursor: pointer;
    line-height: 1;
    padding: 2px;
  }
  .x:hover { color: var(--reject); }

  .sum { font-size: 10px; color: var(--ink-faint); margin: 3px 0 9px; }

  .rate { display: flex; flex-direction: column; gap: 5px; margin-bottom: 9px; }
  .rate .k { font-size: 10px; color: var(--ink-dim); }
  .rate input { width: 100%; accent-color: var(--gpu); cursor: pointer; }

  .bursts { display: flex; gap: 6px; }
  .bursts button {
    appearance: none;
    flex: 1;
    border: 1px solid color-mix(in oklab, var(--gpu) 40%, transparent);
    background: color-mix(in oklab, var(--gpu) 10%, transparent);
    color: var(--ink);
    font-size: 10.5px;
    padding: 5px 0;
    border-radius: 6px;
    cursor: pointer;
  }
  .bursts button:hover { background: color-mix(in oklab, var(--gpu) 20%, transparent); }

  .new {
    appearance: none;
    border: 1px dashed var(--line-strong);
    background: none;
    color: var(--ink-dim);
    font-size: 11.5px;
    padding: 9px;
    border-radius: var(--r-md);
    cursor: pointer;
  }
  .new:hover { color: var(--ink); border-color: var(--ink-faint); }
</style>
