<script lang="ts">
  import type { Device } from '../lib/types'
  import { CATEGORY_COLOR } from '../lib/types'
  import { groupFleet, isStandalone, type Pod } from '../lib/fleet'
  import DeviceTile from './DeviceTile.svelte'

  let {
    devices,
    selectedID = null,
    onselect,
  }: {
    devices: Device[]
    selectedID?: string | null
    onselect?: (d: Device) => void
  } = $props()

  const groups = $derived(groupFleet(devices))

  const podTint = (pod: Pod): string => CATEGORY_COLOR[pod.devices[0].category]
</script>

<div class="fleet">
  {#each groups as g (g.node)}
    <div class="node">
      <div class="node-hd label">node {g.node}</div>
      <div class="pods">
        {#each g.pods as pod (pod.island)}
          <div
            class="pod"
            class:standalone={isStandalone(pod.island)}
            style="--tint: {podTint(pod)}"
          >
            <div class="pod-tag mono">
              {#if isStandalone(pod.island)}
                standalone
              {:else}
                <span class="chain">⛓</span> island {pod.island} · {pod.interconnect}
              {/if}
            </div>
            <div class="row">
              {#each pod.devices as d (d.id)}
                <DeviceTile device={d} selected={d.id === selectedID} {onselect} />
              {/each}
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/each}
</div>

<style>
  .fleet {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    align-items: flex-start;
  }

  .node {
    border: 1px solid var(--line);
    border-radius: var(--r-lg);
    padding: 11px;
    background: var(--bg-2);
  }
  .node-hd {
    margin-bottom: 9px;
  }

  .pods {
    display: flex;
    flex-direction: column;
    gap: 9px;
  }

  .pod {
    --tint: var(--gpu);
    border-radius: var(--r-md);
    padding: 8px;
    border: 1px solid color-mix(in oklab, var(--tint) 38%, transparent);
    background: linear-gradient(
      180deg,
      color-mix(in oklab, var(--tint) 12%, transparent),
      color-mix(in oklab, var(--tint) 2%, transparent)
    );
  }
  .pod.standalone {
    border: 1px dashed var(--line-strong);
    background: none;
  }

  .pod-tag {
    font-size: 9px;
    letter-spacing: 0.03em;
    color: var(--ink-dim);
    margin-bottom: 7px;
  }
  .chain {
    font-size: 11px;
  }

  .row {
    display: flex;
    gap: 7px;
    flex-wrap: wrap;
  }
</style>
