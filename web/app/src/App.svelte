<script lang="ts">
  import { onMount } from 'svelte'
  import { loadScenario, run } from './lib/engine'
  import type { Device, Report } from './lib/types'
  import { PRESETS } from './lib/workloads'
  import { contrastDecos } from './lib/contrast'
  import Fleet from './components/Fleet.svelte'
  import DetailRail from './components/DetailRail.svelte'
  import ModeSwitch, { type Mode } from './components/ModeSwitch.svelte'
  import ContrastPanel from './components/ContrastPanel.svelte'

  let devices = $state<Device[]>([])
  let error = $state<string | null>(null)
  let loading = $state(true)

  let mode = $state<Mode>('contrast')
  let presetId = $state(PRESETS[0].id)
  let active = $state<'sift' | 'legacy'>('sift')
  let report = $state<Report | null>(null)
  let selectedID = $state<string | null>(null)

  const base = import.meta.env.BASE_URL
  const preset = $derived(PRESETS.find((p) => p.id === presetId) ?? PRESETS[0])
  const selected = $derived(devices.find((d) => d.id === selectedID) ?? null)

  const decorations = $derived(
    mode === 'contrast' && report
      ? contrastDecos(active === 'sift' ? report.sift : report.legacy)
      : undefined,
  )

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

  // Re-run the contrast whenever the fleet or the chosen workload mix changes.
  $effect(() => {
    const wl = preset.workloads
    if (!devices.length) return
    run(devices, wl)
      .then((r) => (report = r))
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
        <div class="presets">
          {#each PRESETS as p (p.id)}
            <button class="chip" class:on={p.id === presetId} onclick={() => (presetId = p.id)}>{p.label}</button>
          {/each}
        </div>
      {/if}
    </div>

    <main class="canvas">
      <Fleet {devices} {selectedID} {decorations} onselect={(d) => (selectedID = d.id)} />

      {#if mode === 'contrast' && report}
        <ContrastPanel {report} {active} caption={preset.caption} ontoggle={(s) => (active = s)} />
      {:else}
        <div class="side">
          <div class="soon"><span class="label">{mode}</span><p>This mode is coming in the next increment.</p></div>
          <DetailRail device={selected} />
        </div>
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
  .presets {
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

  .side {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .soon {
    width: 184px;
    border: 1px dashed var(--line-strong);
    border-radius: var(--r-lg);
    padding: 16px;
    color: var(--ink-faint);
  }
  .soon p {
    margin: 8px 0 0;
    font-size: 12px;
    line-height: 1.5;
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
