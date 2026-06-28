<script lang="ts">
  import type { Report, Summary, Outcome } from '../lib/types'
  import { modelName } from '../lib/types'

  let {
    report,
    active,
    caption,
    ontoggle,
  }: {
    report: Report
    active: 'sift' | 'legacy'
    caption: string
    ontoggle: (s: 'sift' | 'legacy') => void
  } = $props()

  const summary = $derived<Summary>(active === 'sift' ? report.sift : report.legacy)

  const mark = (o: Outcome): { text: string; tone: string } => {
    if (o.pending) return { text: 'pending', tone: 'reject' }
    if (!o.feasible) return { text: 'wrong-type', tone: 'reject' }
    if (!o.sameIslandOK) return { text: 'fragmented', tone: 'frag' }
    return { text: 'ok', tone: 'ok' }
  }

  const devLabel = (o: Outcome): string => {
    if (!o.deviceIDs || o.deviceIDs.length === 0) return '—'
    const head = modelName(o.deviceIDs[0])
    return o.deviceIDs.length > 1 ? `${head} ×${o.deviceIDs.length}` : head
  }
</script>

<aside class="panel">
  <p class="caption">{caption}</p>

  <div class="scores">
    {#each [report.sift, report.legacy] as s (s.name)}
      <div class="col" class:win={s.name === 'Sift'}>
        <div class="col-hd">{s.name}</div>
        <div class="stat"><span class="mono big">${s.totalCost.toFixed(2)}</span><span class="unit">/hr</span></div>
        <dl>
          <div><dt>type-correct</dt><dd class="mono">{s.typeCorrect}/{report.workloads}</dd></div>
          <div><dt>fragmented</dt><dd class="mono">{s.fragmented}</dd></div>
          <div><dt>pending</dt><dd class="mono">{s.pending}</dd></div>
        </dl>
      </div>
    {/each}
  </div>

  <div class="toggle">
    <span class="label">show on fleet</span>
    <div class="seg">
      <button class:on={active === 'sift'} onclick={() => ontoggle('sift')}>Sift</button>
      <button class:on={active === 'legacy'} onclick={() => ontoggle('legacy')}>Legacy</button>
    </div>
  </div>

  <ul class="outcomes">
    {#each summary.outcomes as o (o.workload)}
      {@const m = mark(o)}
      <li>
        <span class="wl mono">{o.workload}</span>
        <span class="dev mono">{devLabel(o)}</span>
        <span class="mk {m.tone}">{m.text}</span>
      </li>
    {/each}
  </ul>
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

  .caption {
    margin: 0 0 13px;
    font-size: 12px;
    line-height: 1.5;
    color: var(--ink-dim);
  }

  .scores {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }
  .col {
    border: 1px solid var(--line);
    border-radius: var(--r-md);
    padding: 10px;
  }
  .col.win {
    border-color: color-mix(in oklab, var(--bound) 45%, transparent);
    background: color-mix(in oklab, var(--bound) 7%, transparent);
  }
  .col-hd {
    font-size: 11px;
    font-weight: 600;
    color: var(--ink);
    margin-bottom: 7px;
  }
  .stat {
    display: flex;
    align-items: baseline;
    gap: 3px;
  }
  .big {
    font-size: 19px;
    font-weight: 600;
  }
  .unit {
    font-size: 10px;
    color: var(--ink-faint);
  }
  dl {
    margin: 9px 0 0;
  }
  dl > div {
    display: flex;
    justify-content: space-between;
    font-size: 10.5px;
    padding: 3px 0;
  }
  dt {
    color: var(--ink-faint);
  }
  dd {
    margin: 0;
    color: var(--ink-dim);
  }

  .toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 14px 0 10px;
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
    padding: 3px 11px;
    border-radius: 5px;
    cursor: pointer;
  }
  .seg button.on {
    background: var(--panel-2);
    color: var(--ink);
    box-shadow: inset 0 0 0 1px var(--line-strong);
  }

  .outcomes {
    list-style: none;
    margin: 0;
    padding: 9px 0 0;
    border-top: 1px solid var(--line);
  }
  .outcomes li {
    display: grid;
    grid-template-columns: 1fr auto auto;
    gap: 8px;
    align-items: center;
    padding: 5px 0;
    font-size: 11px;
  }
  .wl {
    color: var(--ink);
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .dev {
    color: var(--ink-dim);
  }
  .mk {
    font-size: 9px;
    letter-spacing: 0.03em;
    padding: 2px 6px;
    border-radius: 5px;
  }
  .mk.ok {
    color: var(--bound);
    background: color-mix(in oklab, var(--bound) 14%, transparent);
  }
  .mk.reject {
    color: var(--reject);
    background: color-mix(in oklab, var(--reject) 14%, transparent);
  }
  .mk.frag {
    color: var(--accent);
    background: color-mix(in oklab, var(--accent) 14%, transparent);
  }
</style>
