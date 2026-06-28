<script lang="ts">
  import type { Trace, Workload, Verdict } from '../lib/types'
  import { modelName } from '../lib/types'
  import { STAGES, STAGE_LABEL, type Stage } from '../lib/explain'

  let {
    trace,
    workload,
    stage,
    onstage,
  }: {
    trace: Trace
    workload: Workload
    stage: Stage
    onstage: (s: Stage) => void
  } = $props()

  const feasible = $derived(
    trace.verdicts.filter((v) => v.feasible).sort((a, b) => a.rank - b.rank),
  )
  const rejected = $derived(trace.verdicts.filter((v) => !v.feasible))
  const bound = $derived(new Set(trace.bound ?? []))

  const boundLabel = $derived(
    !trace.bound || trace.bound.length === 0
      ? '—'
      : trace.bound.length > 1
        ? `${modelName(trace.bound[0])} ×${trace.bound.length}`
        : modelName(trace.bound[0]),
  )

  const reqs = $derived(
    [
      workload.kind,
      `≥ ${workload.minMemoryGB}GB`,
      ...workload.requiredPrecisions,
      workload.deviceCount > 1 ? `×${workload.deviceCount}` : null,
      workload.sameIsland ? 'same-island' : null,
      `cost-weight ${workload.costWeight}`,
    ].filter(Boolean) as string[],
  )

  const reasons = (v: Verdict): string => v.reasons.map((r) => r.detail).join('; ')
</script>

<aside class="panel">
  <div class="reqs">
    <span class="wl mono">{workload.name}</span>
    {#each reqs as r (r)}
      <span class="req">{r}</span>
    {/each}
  </div>

  <div class="stepper">
    {#each STAGES as s, i (s)}
      {#if i > 0}<span class="line" class:done={STAGES.indexOf(stage) >= i}></span>{/if}
      <button class="step" class:on={stage === s} onclick={() => onstage(s)}>
        <span class="n">{i + 1}</span>{STAGE_LABEL[s]}
      </button>
    {/each}
  </div>

  <div class="result">
    {#if trace.err}
      <span class="mk reject">pending</span><span class="rtext">{trace.err}</span>
    {:else}
      <span class="mk bound">bind</span>
      <span class="rtext mono">{boundLabel}{trace.island >= 0 ? ` · island ${trace.island}` : ''}</span>
    {/if}
  </div>

  <div class="label sec">feasible · by score rank</div>
  <ul class="rows">
    {#each feasible as v (v.deviceID)}
      <li class:bound={bound.has(v.deviceID)}>
        <span class="rank mono">#{v.rank}</span>
        <span class="id mono">{v.deviceID}</span>
        <span class="score mono">{v.score.costComponent.toFixed(2)}</span>
      </li>
    {/each}
  </ul>

  {#if rejected.length}
    <div class="label sec">rejected · {rejected.length}</div>
    <ul class="rows rej">
      {#each rejected as v (v.deviceID)}
        <li>
          <span class="id mono">{v.deviceID}</span>
          <span class="why">{reasons(v)}</span>
        </li>
      {/each}
    </ul>
  {/if}
</aside>

<style>
  .panel {
    width: 268px;
    flex: none;
    background: var(--panel);
    border: 1px solid var(--line);
    border-radius: var(--r-lg);
    padding: 14px;
    position: sticky;
    top: 16px;
  }

  .reqs {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
    align-items: center;
    margin-bottom: 14px;
  }
  .wl {
    font-size: 12px;
    font-weight: 600;
    margin-right: 3px;
  }
  .req {
    font-size: 9.5px;
    padding: 2px 7px;
    border-radius: 5px;
    background: var(--panel-2);
    color: var(--ink-dim);
  }

  .stepper {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 12px;
  }
  .step {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    appearance: none;
    border: 1px solid var(--line);
    background: none;
    color: var(--ink-faint);
    font-family: var(--font-sans);
    font-size: 11px;
    padding: 5px 9px;
    border-radius: 7px;
    cursor: pointer;
    transition:
      color 0.12s,
      border-color 0.12s,
      background 0.12s;
  }
  .step .n {
    width: 15px;
    height: 15px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.12);
    font-size: 9px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .step.on {
    color: var(--ink);
    border-color: color-mix(in oklab, var(--gpu) 55%, transparent);
    background: color-mix(in oklab, var(--gpu) 13%, transparent);
  }
  .step.on .n {
    background: var(--gpu);
    color: #fff;
  }
  .line {
    flex: 1;
    height: 1px;
    background: var(--line);
  }
  .line.done {
    background: color-mix(in oklab, var(--gpu) 55%, transparent);
  }

  .result {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 9px 0 12px;
    border-bottom: 1px solid var(--line);
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
  }
  .mk.bound {
    color: var(--bound);
    background: color-mix(in oklab, var(--bound) 15%, transparent);
  }
  .mk.reject {
    color: var(--reject);
    background: color-mix(in oklab, var(--reject) 15%, transparent);
  }

  .sec {
    margin: 13px 0 6px;
  }

  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .rows li {
    display: grid;
    grid-template-columns: 28px 1fr auto;
    gap: 8px;
    align-items: center;
    padding: 4px 0;
    font-size: 11px;
    border-bottom: 1px solid var(--line);
  }
  .rows li:last-child {
    border-bottom: none;
  }
  .rows li.bound {
    color: var(--bound);
  }
  .rank {
    color: var(--ink-faint);
  }
  .rows li.bound .rank {
    color: var(--bound);
  }
  .id {
    color: var(--ink-dim);
  }
  .rows li.bound .id {
    color: var(--bound);
  }
  .score {
    color: var(--ink-faint);
    text-align: right;
  }

  .rej li {
    grid-template-columns: auto 1fr;
  }
  .why {
    font-size: 9.5px;
    color: color-mix(in oklab, var(--reject) 80%, var(--ink));
    line-height: 1.35;
  }
</style>
