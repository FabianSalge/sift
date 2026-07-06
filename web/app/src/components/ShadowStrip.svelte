<script lang="ts">
  import type { ShadowMetrics } from '../lib/types'

  let {
    shadow,
    sift,
    deviceCount,
  }: {
    shadow: ShadowMetrics
    sift: { usefulDone: number; queue: number; cost: number }
    deviceCount: number
  } = $props()

  const wastedPct = $derived(deviceCount ? Math.round((shadow.wasted / deviceCount) * 100) : 0)
</script>

<div class="shadow">
  <span class="k">legacy shadow</span>
  <span class="txt">same traffic, first-fit `gpu: N`:</span>
  <span class="m mono" class:bad={shadow.wasted > 0}>{shadow.wasted} wasted ({wastedPct}%)</span>
  <span class="m mono" class:bad={shadow.queue > sift.queue}>queue {shadow.queue}</span>
  <span class="m mono">{shadow.usefulDone} useful <i>vs {sift.usefulDone}</i></span>
  <span class="m mono">${shadow.cost.toFixed(0)} spent <i>vs ${sift.cost.toFixed(0)}</i></span>
  <details class="why">
    <summary>what is this?</summary>
    <p>
      The legacy integer scheduler is simulated invisibly on the exact same arrivals and
      fleet edits. A job it places on an unfit device holds that capacity for the full
      duration and does no useful work. Illustrative stream — Sift decides every placement;
      not a production queueing system.
    </p>
  </details>
</div>

<style>
  .shadow {
    display: flex;
    align-items: baseline;
    gap: 12px;
    flex-wrap: wrap;
    padding: 8px 14px;
    margin: 16px 0 14px;
    border: 1px solid color-mix(in oklab, var(--reject) 30%, transparent);
    border-radius: var(--r-md);
    background: color-mix(in oklab, var(--reject) 5%, transparent);
    font-size: 12px;
  }
  .k {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: color-mix(in oklab, var(--reject) 80%, var(--ink));
    font-weight: 600;
  }
  .txt { color: var(--ink-dim); }
  .m { color: var(--ink-dim); }
  .m.bad { color: var(--reject); }
  .m i { font-style: normal; color: var(--ink-faint); font-size: 10px; }
  .why summary {
    cursor: pointer;
    font-size: 10px;
    color: var(--ink-faint);
    list-style: none;
  }
  .why[open] {
    flex-basis: 100%;
  }
  .why p {
    margin: 6px 0 0;
    font-size: 11px;
    line-height: 1.5;
    color: var(--ink-faint);
    max-width: 78ch;
  }
</style>
