<script lang="ts">
  import { onMount } from 'svelte'
  import { loadScenario, clusterInit, clusterSubmit, clusterAdvance, clusterAddNode, clusterDrainNode, clusterExplain } from './lib/engine'
  import type { ClusterSnapshot, Deco, ClusterJob, Trace, Device } from './lib/types'
  import { DEFAULT_TEMPLATES, type WorkloadTemplate } from './lib/templates'
  import { mulberry32, jitter } from './lib/rng'
  import { Generator } from './lib/traffic'
  import { MAX_DEVICES, buildNode, type MachineTemplate } from './lib/machines'
  import Fleet from './components/Fleet.svelte'
  import QueueRail from './components/QueueRail.svelte'
  import ShadowStrip from './components/ShadowStrip.svelte'
  import Transport from './components/Transport.svelte'
  import WorkloadDock from './components/WorkloadDock.svelte'
  import MachineDock from './components/MachineDock.svelte'
  import JobPanel from './components/JobPanel.svelte'

  const params = new URLSearchParams(location.search)
  const seed = Math.abs(Number(params.get('seed') ?? 0)) || 4212
  let speed = $state(Math.min(20, Math.max(1, Number(params.get('speed') ?? 0) || 4)))
  let paused = $state(false)

  let snap = $state<ClusterSnapshot | null>(null)
  let error = $state<string | null>(null)
  let loading = $state(true)
  let templates = $state<WorkloadTemplate[]>(structuredClone(DEFAULT_TEMPLATES))
  let selected = $state<{ job: ClusterJob; trace: Trace | null } | null>(null)

  const rng = mulberry32(seed)
  const gen = new Generator(rng)
  let simNow = 0
  let ticking = false
  const jobCounts = new Map<string, number>()
  let pulses = $state<Set<string>>(new Set())

  const base = import.meta.env.BASE_URL

  const ratePerHr = $derived(snap ? snap.running.reduce((s, j) => s + j.costPerHr, 0) : 0)

  // Running jobs ring their devices; draining devices dim.
  const decorations = $derived.by(() => {
    if (!snap) return undefined
    const m = new Map<string, Deco>()
    for (const j of snap.running) {
      for (const id of j.deviceIDs ?? []) {
        m.set(id, { ring: 'bound', tag: j.workload.name, pulse: pulses.has(id) })
      }
    }
    for (const d of snap.devices) {
      if (d.draining) m.set(d.id, { ...m.get(d.id), dim: true, reason: `${d.id} · draining` })
    }
    return m
  })

  function submitFrom(t: WorkloadTemplate, duration: number): Promise<number> {
    const n = (jobCounts.get(t.id) ?? 0) + 1
    jobCounts.set(t.id, n)
    return clusterSubmit({ ...t.workload, name: `${t.id}-${n}` }, duration)
  }

  async function burst(t: WorkloadTemplate, n: number) {
    for (let i = 0; i < n; i++) await submitFrom(t, jitter(rng, t.durationS))
  }

  let nextNode = 0
  let nextIsland = 0
  const serials = new Map<string, number>()

  async function addMachine(t: MachineTemplate) {
    if (!snap || snap.devices.length + t.count > MAX_DEVICES) return
    const serial = serials.get(t.id) ?? 0
    serials.set(t.id, serial + t.count)
    const devs = buildNode(t, nextNode++, t.islanded ? nextIsland++ : -1, serial)
    await clusterAddNode(devs)
  }

  async function drainNode(node: number) {
    await clusterDrainNode(node)
  }

  async function selectJob(j: ClusterJob) {
    selected = { job: j, trace: null }
    const trace = await clusterExplain(j.id)
    if (selected?.job.id === j.id) selected = { job: j, trace }
  }

  function selectDevice(d: Device) {
    const cd = snap?.devices.find((x) => x.id === d.id)
    if (!cd || cd.jobID < 0) return
    const j = snap?.running.find((x) => x.id === cd.jobID)
    if (j) selectJob(j)
  }

  // A node is "draining" when every one of its remaining devices is.
  const drainingNodes = $derived.by(() => {
    const all = new Map<number, boolean>()
    for (const d of snap?.devices ?? []) all.set(d.node, (all.get(d.node) ?? true) && d.draining)
    return new Set([...all].filter(([, v]) => v).map(([n]) => n))
  })

  async function tick() {
    if (paused || ticking || loading || error) return
    ticking = true
    try {
      simNow += 0.1 * speed
      for (const due of gen.due(templates, simNow)) await submitFrom(due.template, due.duration)
      const s = await clusterAdvance(simNow)
      const placedIDs: string[] = []
      for (const e of s.events ?? []) {
        if (e.kind === 'placed') for (const id of e.deviceIDs ?? []) placedIDs.push(id)
      }
      if (placedIDs.length) {
        const next = new Set(pulses)
        for (const id of placedIDs) next.add(id)
        pulses = next
        // Each batch clears only its own ids, so overlapping placements keep
        // their full 700ms flash.
        setTimeout(() => {
          const cleared = new Set(pulses)
          for (const id of placedIDs) cleared.delete(id)
          pulses = cleared
        }, 700)
      }
      snap = s
      if (selected) {
        const cur = [...s.running, ...s.queue].find((j) => j.id === selected!.job.id)
        if (!cur) {
          selected = null
        } else {
          const justPlaced = cur.placedAt >= 0 && selected.job.placedAt < 0
          selected = { job: cur, trace: selected.trace }
          // A queued job that just placed has a real "why here" now — refetch
          // the trace instead of keeping the click-time no-fit ranking.
          if (justPlaced) selectJob(cur)
        }
      }
    } catch (e) {
      error = String(e)
    } finally {
      ticking = false
    }
  }

  onMount(async () => {
    try {
      const yaml = await (await fetch(`${base}scenarios/realistic-2026.yaml`)).text()
      const fleet = await loadScenario(yaml)
      for (const d of fleet) {
        nextNode = Math.max(nextNode, d.node + 1)
        nextIsland = Math.max(nextIsland, d.island + 1)
        const model = d.id.replace(/-\d+$/, '')
        serials.set(model, (serials.get(model) ?? 0) + 1)
      }
      await clusterInit(fleet)
      snap = await clusterAdvance(0)
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  })

  $effect(() => {
    if (loading || error) return
    const id = setInterval(tick, 100)
    return () => clearInterval(id)
  })
</script>

<div class="frame">
  <header class="masthead">
    <div class="row">
      <div class="brand">
        <span class="mark" aria-hidden="true"></span>
        <span class="word">sift</span>
        <span class="live mono">live cluster</span>
      </div>
      {#if snap}
        <div class="stats mono">
          <span><b>{snap.usefulDone}</b> useful</span>
          <span class:bad={snap.queue.length > 0}><b>{snap.queue.length}</b> queued</span>
          <span><b>${ratePerHr.toFixed(2)}</b>/h</span>
          <span>{snap.devices.length} devices</span>
        </div>
      {/if}
    </div>
    <p class="tag">
      capability- &amp; topology-aware accelerator scheduling —
      <span class="dim">a cluster you operate: invent workloads, grow the fleet, and watch what legacy would waste</span>
    </p>
  </header>

  {#if loading}
    <div class="status"><span class="label">booting</span><p>loading the scheduler engine…</p></div>
  {:else if error}
    <div class="status err"><span class="label">error</span><p class="mono">{error}</p></div>
  {:else if snap}
    <ShadowStrip
      shadow={snap.shadow}
      sift={{ usefulDone: snap.usefulDone, queue: snap.queue.length, cost: snap.cost }}
      deviceCount={snap.devices.length}
    />

    <div class="bar">
      <Transport {paused} {speed} clock={snap.clock} {seed} ontoggle={() => (paused = !paused)} onspeed={(v) => (speed = v)} />
      <QueueRail queue={snap.queue} onselect={selectJob} />
    </div>

    <main class="canvas">
      <div class="left">
        <Fleet devices={snap.devices} {decorations} ondrain={drainNode} {drainingNodes} onselect={selectDevice} />
      </div>
      <aside class="dock">
        {#if selected}
          <JobPanel job={selected.job} trace={selected.trace} clock={snap.clock} onclose={() => (selected = null)} />
        {/if}
        <WorkloadDock bind:templates onburst={burst} />
        <MachineDock deviceCount={snap.devices.length} onadd={addMachine} />
      </aside>
    </main>
  {/if}
</div>

<style>
  .frame {
    max-width: 1280px;
    margin: 0 auto;
    padding: 36px 28px 80px;
  }

  .masthead { border-bottom: 1px solid var(--line); padding-bottom: 20px; }
  .masthead .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }
  .brand { display: flex; align-items: center; gap: 11px; }
  .mark {
    width: 13px;
    height: 13px;
    border-radius: 3px;
    background: linear-gradient(135deg, var(--gpu), var(--infer));
    box-shadow: 0 0 18px rgba(99, 102, 241, 0.5);
  }
  .word {
    font-family: var(--font-mono);
    font-weight: 600;
    font-size: 22px;
    letter-spacing: 0.02em;
  }
  .live {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.09em;
    color: var(--ink-faint);
    border: 1px solid var(--line);
    border-radius: 999px;
    padding: 3px 9px;
  }
  .stats { display: flex; gap: 16px; font-size: 12px; color: var(--ink-dim); }
  .stats b { font-size: 15px; color: var(--ink); font-weight: 600; }
  .stats .bad b { color: var(--accent); }
  .tag { margin: 12px 0 0; font-size: 13.5px; color: var(--ink-dim); max-width: 78ch; }
  .tag .dim { color: var(--ink-faint); }

  .bar {
    display: flex;
    align-items: center;
    gap: 18px;
    flex-wrap: wrap;
    margin-bottom: 18px;
  }

  .canvas { display: flex; gap: 16px; align-items: flex-start; }
  .left { flex: 1; min-width: 0; }
  .dock {
    width: 264px;
    flex: none;
    display: flex;
    flex-direction: column;
    gap: 14px;
    position: sticky;
    top: 16px;
  }

  .status {
    border: 1px dashed var(--line-strong);
    border-radius: var(--r-lg);
    padding: 56px 28px;
    text-align: center;
    color: var(--ink-faint);
    background: var(--bg-2);
    margin-top: 26px;
  }
  .status p { margin: 10px 0 0; font-size: 13px; }
  .status.err { border-color: color-mix(in oklab, var(--reject) 50%, transparent); }
  .status.err p { color: var(--reject); }
</style>
