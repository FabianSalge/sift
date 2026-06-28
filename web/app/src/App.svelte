<script lang="ts">
  import { onMount } from 'svelte'
  import { loadScenario } from './lib/engine'
  import type { Device } from './lib/types'
  import Fleet from './components/Fleet.svelte'
  import DetailRail from './components/DetailRail.svelte'

  let devices = $state<Device[]>([])
  let selectedID = $state<string | null>(null)
  let error = $state<string | null>(null)
  let loading = $state(true)

  const selected = $derived(devices.find((d) => d.id === selectedID) ?? null)
  const base = import.meta.env.BASE_URL

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

  <main class="stage">
    {#if loading}
      <div class="status"><span class="label">booting</span><p>loading the scheduler engine…</p></div>
    {:else if error}
      <div class="status err"><span class="label">error</span><p class="mono">{error}</p></div>
    {:else}
      <div class="canvas">
        <Fleet {devices} {selectedID} onselect={(d) => (selectedID = d.id)} />
        <DetailRail device={selected} />
      </div>
    {/if}
  </main>
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

  .stage {
    margin-top: 26px;
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
