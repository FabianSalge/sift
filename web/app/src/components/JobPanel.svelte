<script lang="ts">
  import type { ClusterJob, Trace, Verdict } from '../lib/types'
  import { modelName } from '../lib/types'

  let {
    job,
    trace,
    clock,
    onclose,
  }: { job: ClusterJob; trace: Trace | null; clock: number; onclose: () => void } = $props()

  const reqs = $derived(
    [
      job.workload.kind,
      `≥ ${job.workload.minMemoryGB}GB`,
      ...job.workload.requiredPrecisions,
      job.workload.deviceCount > 1 ? `×${job.workload.deviceCount}` : null,
      job.workload.sameIsland ? 'same-island' : null,
      `cost-weight ${job.workload.costWeight}`,
    ].filter(Boolean) as string[],
  )

  const feasible = $derived(trace ? trace.verdicts.filter((v) => v.feasible).sort((a, b) => a.rank - b.rank) : [])
  const rejected = $derived(trace ? trace.verdicts.filter((v) => !v.feasible) : [])
  const bound = $derived(new Set(job.deviceIDs ?? []))
  const reasons = (v: Verdict): string => v.reasons.map((r) => r.detail).join('; ')
  const remaining = $derived(job.end >= 0 ? Math.max(0, job.end - clock) : 0)
</script>

<section class="panel">
  <header>
    <span class="wl mono">{job.workload.name}</span>
    <button class="x" title="close" onclick={onclose}>×</button>
  </header>

  <div class="reqs">
    {#each reqs as r (r)}<span class="req">{r}</span>{/each}
  </div>

  <div class="result">
    {#if job.placedAt < 0}
      <span class="mk reject">queued</span>
      <span class="rtext">waiting {Math.max(0, clock - job.arrivedAt).toFixed(0)}s — nothing fits right now</span>
    {:else}
      <span class="mk bound">running</span>
      <span class="rtext mono">
        {(job.deviceIDs ?? []).map(modelName).join(', ')} · {remaining.toFixed(0)}s left · ${job.costPerHr.toFixed(2)}/h
      </span>
    {/if}
  </div>

  {#if trace}
    <div class="label sec">{job.placedAt < 0 ? 'current fits' : 'why here'} — feasible by score</div>
    <ul class="rows">
      {#each feasible as v (v.deviceID)}
        <li class:bound={bound.has(v.deviceID)}>
          <span class="rank mono">#{v.rank}</span>
          <span class="id mono">{v.deviceID}</span>
          <span class="score mono">{v.score.costComponent.toFixed(2)}</span>
        </li>
      {/each}
      {#if feasible.length === 0}<li class="none">none free & capable</li>{/if}
    </ul>
    {#if rejected.length}
      <div class="label sec">rejected · {rejected.length}</div>
      <ul class="rows rej">
        {#each rejected as v (v.deviceID)}
          <li><span class="id mono">{v.deviceID}</span><span class="why">{reasons(v)}</span></li>
        {/each}
      </ul>
    {/if}
  {/if}
</section>

<style>
  .panel {
    border: 1px solid color-mix(in oklab, var(--gpu) 40%, transparent);
    border-radius: var(--r-lg);
    background: var(--panel);
    padding: 14px;
  }
  header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
  .wl { font-size: 12px; font-weight: 600; }
  .x {
    appearance: none;
    border: none;
    background: none;
    color: var(--ink-faint);
    font-size: 15px;
    cursor: pointer;
    line-height: 1;
  }
  .x:hover { color: var(--ink); }

  .reqs { display: flex; flex-wrap: wrap; gap: 5px; margin-bottom: 12px; }
  .req {
    font-size: 9.5px;
    padding: 2px 7px;
    border-radius: 5px;
    background: var(--panel-2);
    color: var(--ink-dim);
  }

  .result {
    display: flex;
    align-items: center;
    gap: 8px;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--line);
  }
  .rtext { font-size: 11.5px; color: var(--ink); }
  .mk {
    font-size: 9px;
    letter-spacing: 0.04em;
    padding: 2px 7px;
    border-radius: 5px;
    text-transform: uppercase;
    white-space: nowrap;
  }
  .mk.bound { color: var(--bound); background: color-mix(in oklab, var(--bound) 15%, transparent); }
  .mk.reject { color: var(--reject); background: color-mix(in oklab, var(--reject) 15%, transparent); }

  .sec { margin: 13px 0 6px; }
  .rows { list-style: none; margin: 0; padding: 0; max-height: 200px; overflow-y: auto; }
  .rows li {
    display: grid;
    grid-template-columns: 28px 1fr auto;
    gap: 8px;
    align-items: center;
    padding: 4px 0;
    font-size: 11px;
    border-bottom: 1px solid var(--line);
  }
  .rows li:last-child { border-bottom: none; }
  .rows li.bound, .rows li.bound .rank, .rows li.bound .id { color: var(--bound); }
  .rows li.none { grid-template-columns: 1fr; color: var(--ink-faint); font-size: 10.5px; }
  .rank { color: var(--ink-faint); }
  .id { color: var(--ink-dim); }
  .score { color: var(--ink-faint); text-align: right; }
  .rej li { grid-template-columns: auto 1fr; }
  .why { font-size: 9.5px; color: color-mix(in oklab, var(--reject) 80%, var(--ink)); line-height: 1.35; }
</style>
