<script lang="ts">
  import { onMount, untrack } from 'svelte'
  import { simulate } from '../lib/engine'
  import { STREAM } from '../lib/streams'
  import { clusterState, type ClusterState } from '../lib/stream'
  import type { Device, SimResult, SchedulerSim } from '../lib/types'
  import StreamFleet from './StreamFleet.svelte'

  let {
    devices,
    initialT = 0,
    initialSpeed = 4,
  }: { devices: Device[]; initialT?: number; initialSpeed?: number } = $props()

  let sim = $state<SimResult | null>(null)
  let t = $state(untrack(() => initialT))
  let playing = $state(false)
  let speed = $state(untrack(() => initialSpeed)) // sim-time units per real second

  const horizon = $derived(sim ? sim.horizon : 0)
  const upcoming = $derived(STREAM.filter((a) => a.at > t).slice(0, 7))

  const columns = $derived(
    sim
      ? ([
          { name: 'Sift', sched: sim.sift, win: true },
          { name: 'Legacy', sched: sim.legacy, win: false },
        ] as { name: string; sched: SchedulerSim; win: boolean }[])
      : [],
  )
  const stateOf = (s: SchedulerSim): ClusterState => clusterState(s, t)

  onMount(async () => {
    sim = await simulate(devices, STREAM)
    if (initialT <= 0) playing = true // autoplay from the start; a deep-linked t stays paused
  })

  $effect(() => {
    if (!playing || !sim) return
    const id = setInterval(() => {
      t = Math.min(t + speed * 0.05, horizon)
      if (t >= horizon) playing = false
    }, 50)
    return () => clearInterval(id)
  })

  function toggle() {
    if (t >= horizon) t = 0
    playing = !playing
  }
  function setT(v: number) {
    playing = false
    t = v
  }
  const pct = (n: number) => `${Math.round((n / Math.max(1, devices.length)) * 100)}%`
</script>

<div class="stream">
  <div class="transport">
    <button class="play" aria-label={playing ? 'pause' : 'play'} onclick={toggle}>
      {#if playing}<span class="ico">❚❚</span>{:else}<span class="ico tri">▶</span>{/if}
    </button>
    <input class="scrub" type="range" min="0" max={horizon} step="0.1" value={t} oninput={(e) => setT(+e.currentTarget.value)} />
    <span class="clock mono">t = {t.toFixed(1)}</span>
    <label class="speed">
      <span class="k">speed</span>
      <input type="range" min="1" max="10" step="1" bind:value={speed} />
    </label>
  </div>

  <div class="cols">
    {#each columns as col (col.name)}
      {@const cs = stateOf(col.sched)}
      <section class="col" class:win={col.win}>
        <header>
          <span class="name mono">{col.name}</span>
          <span class="done"><b class="mono">{cs.usefulDone}</b> useful</span>
        </header>

        <StreamFleet {devices} state={cs} />

        <div class="util">
          <div class="bar">
            <span class="seg useful" style="width: {pct(cs.busyCount - cs.wastedCount)}"></span>
            <span class="seg wasted" style="width: {pct(cs.wastedCount)}"></span>
          </div>
        </div>

        <dl class="metrics mono">
          <div><dt>running</dt><dd>{cs.busyCount - cs.wastedCount}</dd></div>
          <div class:bad={cs.wastedCount > 0}><dt>wasted</dt><dd>{cs.wastedCount}</dd></div>
          <div class:bad={cs.queue > 0}><dt>queue</dt><dd>{cs.queue}</dd></div>
          <div><dt>cost</dt><dd>${cs.cost.toFixed(0)}</dd></div>
        </dl>
      </section>
    {/each}
  </div>

  <div class="incoming">
    <span class="label">incoming</span>
    <div class="rail">
      {#each upcoming as a (a.workload.name)}
        <span class="job mono" class:gang={a.workload.deviceCount > 1}>
          {a.workload.name}{a.workload.deviceCount > 1 ? ` ×${a.workload.deviceCount}` : ''}
        </span>
      {/each}
      {#if upcoming.length === 0}<span class="job done mono">stream complete</span>{/if}
    </div>
  </div>

  <p class="note">
    Illustrative stream — Sift decides every placement; the sim models that a job on an
    unfit device does no useful work. Not a production queueing system.
  </p>
</div>

<style>
  .stream {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .transport {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 9px 14px;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: var(--panel);
    max-width: 720px;
  }
  .play {
    flex: none;
    width: 30px;
    height: 30px;
    border-radius: 50%;
    border: none;
    cursor: pointer;
    background: linear-gradient(180deg, color-mix(in oklab, var(--gpu) 92%, white), var(--gpu));
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .ico {
    font-size: 10px;
  }
  .ico.tri {
    font-size: 11px;
    margin-left: 2px;
  }
  .scrub {
    flex: 1;
    min-width: 140px;
    accent-color: var(--gpu);
    cursor: pointer;
  }
  .clock {
    flex: none;
    font-size: 12px;
    color: var(--ink-dim);
    min-width: 64px;
  }
  .speed {
    display: flex;
    align-items: center;
    gap: 7px;
  }
  .speed .k {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--ink-faint);
  }
  .speed input {
    width: 80px;
    accent-color: var(--gpu);
  }

  .cols {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }
  .col {
    border: 1px solid var(--line);
    border-radius: var(--r-lg);
    padding: 14px;
    background: var(--bg-2);
  }
  .col.win {
    border-color: color-mix(in oklab, var(--bound) 35%, transparent);
  }
  .col header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 12px;
  }
  .name {
    font-size: 14px;
    font-weight: 600;
  }
  .done {
    font-size: 12px;
    color: var(--ink-dim);
  }
  .done b {
    font-size: 17px;
    color: var(--bound);
    font-weight: 600;
  }

  .util {
    margin: 13px 0 10px;
  }
  .bar {
    display: flex;
    height: 7px;
    border-radius: 4px;
    overflow: hidden;
    background: rgba(255, 255, 255, 0.06);
  }
  .seg.useful {
    background: var(--bound);
  }
  .seg.wasted {
    background: var(--reject);
  }

  .metrics {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 8px;
  }
  .metrics > div {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .metrics dt {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--ink-faint);
  }
  .metrics dd {
    margin: 0;
    font-size: 15px;
    font-weight: 500;
  }
  .metrics > div.bad dd {
    color: var(--reject);
  }

  .incoming {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .rail {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }
  .job {
    font-size: 11px;
    padding: 3px 9px;
    border-radius: 6px;
    background: var(--panel);
    border: 1px solid var(--line);
    color: var(--ink-dim);
  }
  .job.gang {
    border-color: color-mix(in oklab, var(--train) 45%, transparent);
    color: var(--ink);
  }
  .job.done {
    color: var(--ink-faint);
    border-style: dashed;
  }

  .note {
    margin: 4px 0 0;
    font-size: 11px;
    line-height: 1.5;
    color: var(--ink-faint);
    max-width: 72ch;
  }
</style>
