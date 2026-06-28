<script lang="ts">
  import { onMount } from 'svelte'
  import { loadScenario, run, explain } from './lib/engine'
  import type { Device, Workload, Report, Trace } from './lib/types'
  import { PRESETS, EXPLAIN_WORKLOADS } from './lib/workloads'
  import { contrastDecos } from './lib/contrast'
  import { explainDecos, type Stage } from './lib/explain'
  import Fleet from './components/Fleet.svelte'
  import ModeSwitch, { type Mode } from './components/ModeSwitch.svelte'
  import ContrastPanel from './components/ContrastPanel.svelte'
  import ExplainPanel from './components/ExplainPanel.svelte'
  import SandboxPanel from './components/SandboxPanel.svelte'

  let devices = $state<Device[]>([])
  let error = $state<string | null>(null)
  let loading = $state(true)

  // Initial view is deep-linkable: ?mode=explain&wl=train-llm&stage=score
  const params = new URLSearchParams(location.search)
  const pick = <T extends string>(key: string, allowed: readonly T[], fallback: T): T => {
    const v = params.get(key) as T | null
    return v && allowed.includes(v) ? v : fallback
  }

  let mode = $state<Mode>(pick('mode', ['contrast', 'explain', 'sandbox'] as const, 'contrast'))
  let presetId = $state(pick('preset', PRESETS.map((p) => p.id), PRESETS[0].id))
  let active = $state<'sift' | 'legacy'>(pick('show', ['sift', 'legacy'] as const, 'sift'))
  let report = $state<Report | null>(null)
  let selectedID = $state<string | null>(null)

  let explainName = $state(pick('wl', EXPLAIN_WORKLOADS.map((w) => w.name), EXPLAIN_WORKLOADS[0].name))
  let stage = $state<Stage>(pick('stage', ['filter', 'score', 'bind'] as const, 'filter'))
  let trace = $state<Trace | null>(null)

  let sandboxWorkload = $state<Workload>({
    name: 'custom',
    kind: 'train',
    minMemoryGB: 80,
    requiredPrecisions: ['bf16'],
    deviceCount: 1,
    sameIsland: false,
    gang: false,
    latencySensitive: false,
    costWeight: 0.5,
  })
  let sandboxTrace = $state<Trace | null>(null)

  const base = import.meta.env.BASE_URL
  const preset = $derived(PRESETS.find((p) => p.id === presetId) ?? PRESETS[0])
  const explainWorkload = $derived(
    EXPLAIN_WORKLOADS.find((w) => w.name === explainName) ?? EXPLAIN_WORKLOADS[0],
  )

  const decorations = $derived.by(() => {
    if (mode === 'contrast' && report) return contrastDecos(active === 'sift' ? report.sift : report.legacy)
    if (mode === 'explain' && trace) return explainDecos(trace, stage)
    if (mode === 'sandbox' && sandboxTrace) return explainDecos(sandboxTrace, 'bind')
    return undefined
  })

  onMount(async () => {
    try {
      const yaml = await (await fetch(`${base}scenarios/realistic-2026.yaml`)).text()
      devices = await loadScenario(yaml)
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  })

  // Re-run the contrast when the fleet or workload mix changes.
  $effect(() => {
    const wl = preset.workloads
    if (!devices.length) return
    run(devices, wl)
      .then((r) => (report = r))
      .catch((e) => (error = String(e)))
  })

  // Re-trace when the explained workload changes.
  $effect(() => {
    const w = explainWorkload
    if (!devices.length) return
    explain(devices, w, null)
      .then((t) => (trace = t))
      .catch((e) => (error = String(e)))
  })

  // Live-trace the sandbox workload on any field edit (stringify reads all fields).
  $effect(() => {
    const payload = JSON.stringify(sandboxWorkload)
    if (!devices.length) return
    explain(devices, JSON.parse(payload) as Workload, null)
      .then((t) => (sandboxTrace = t))
      .catch((e) => (error = String(e)))
  })
</script>

<div class="frame">
  <header class="masthead">
    <div class="row">
      <div class="brand">
        <span class="mark" aria-hidden="true"></span>
        <span class="word">sift</span>
      </div>
      {#if devices.length}
        <span class="count mono">{devices.length} devices · realistic-2026</span>
      {/if}
    </div>
    <p class="tag">
      capability- &amp; topology-aware accelerator scheduling — <span class="dim"
        >place each workload on the right device, in the right place, at the right cost</span
      >
    </p>
  </header>

  {#if loading}
    <div class="status"><span class="label">booting</span><p>loading the scheduler engine…</p></div>
  {:else if error}
    <div class="status err"><span class="label">error</span><p class="mono">{error}</p></div>
  {:else}
    <div class="controls">
      <ModeSwitch {mode} onchange={(m) => (mode = m)} />
      {#if mode === 'contrast'}
        <div class="chips">
          {#each PRESETS as p (p.id)}
            <button class="chip" class:on={p.id === presetId} onclick={() => (presetId = p.id)}>{p.label}</button>
          {/each}
        </div>
      {:else if mode === 'explain'}
        <div class="chips">
          {#each EXPLAIN_WORKLOADS as w (w.name)}
            <button class="chip" class:on={w.name === explainName} onclick={() => (explainName = w.name)}>{w.name}</button>
          {/each}
        </div>
      {/if}
    </div>

    <main class="canvas">
      <Fleet {devices} {selectedID} {decorations} onselect={(d) => (selectedID = d.id)} />

      {#if mode === 'contrast'}
        {#if report}
          <ContrastPanel {report} {active} caption={preset.caption} ontoggle={(s) => (active = s)} />
        {/if}
      {:else if mode === 'explain'}
        {#if trace}
          <ExplainPanel {trace} workload={explainWorkload} {stage} onstage={(s) => (stage = s)} />
        {/if}
      {:else if mode === 'sandbox'}
        <SandboxPanel bind:workload={sandboxWorkload} trace={sandboxTrace} />
      {/if}
    </main>
  {/if}
</div>

<style>
  .frame {
    max-width: 1180px;
    margin: 0 auto;
    padding: 36px 28px 80px;
  }

  .masthead {
    border-bottom: 1px solid var(--line);
    padding-bottom: 20px;
  }
  .masthead .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 11px;
  }
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
  .count {
    font-size: 11px;
    color: var(--ink-faint);
  }
  .tag {
    margin: 12px 0 0;
    font-size: 13.5px;
    color: var(--ink-dim);
    max-width: 70ch;
  }
  .tag .dim {
    color: var(--ink-faint);
  }

  .controls {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
    margin: 24px 0 18px;
  }
  .chips {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }
  .chip {
    appearance: none;
    border: 1px solid var(--line);
    background: var(--panel);
    color: var(--ink-dim);
    font-family: var(--font-sans);
    font-size: 12px;
    padding: 6px 12px;
    border-radius: 999px;
    cursor: pointer;
    transition:
      color 0.12s,
      border-color 0.12s,
      background 0.12s;
  }
  .chip:hover {
    color: var(--ink);
    border-color: var(--line-strong);
  }
  .chip.on {
    color: var(--ink);
    border-color: color-mix(in oklab, var(--gpu) 55%, transparent);
    background: color-mix(in oklab, var(--gpu) 14%, transparent);
  }

  .canvas {
    display: flex;
    gap: 16px;
    align-items: flex-start;
  }
  .canvas :global(.fleet) {
    flex: 1;
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
  .status p {
    margin: 10px 0 0;
    font-size: 13px;
  }
  .status.err {
    border-color: color-mix(in oklab, var(--reject) 50%, transparent);
  }
  .status.err p {
    color: var(--reject);
  }
</style>
