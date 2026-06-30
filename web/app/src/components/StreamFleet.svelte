<script lang="ts">
  import type { Device } from '../lib/types'
  import type { ClusterState } from '../lib/stream'
  import { CATEGORY_COLOR } from '../lib/types'
  import { groupFleet, isStandalone } from '../lib/fleet'

  let { devices, state }: { devices: Device[]; state: ClusterState } = $props()

  const groups = $derived(groupFleet(devices))
</script>

<div class="sf">
  {#each groups as g (g.node)}
    <div class="node">
      {#each g.pods as pod (pod.island)}
        <div
          class="pod"
          class:isl={!isStandalone(pod.island)}
          style="--tint: {CATEGORY_COLOR[pod.devices[0].category]}"
        >
          {#each pod.devices as d (d.id)}
            {@const s = state.busy.get(d.id)}
            <span
              class="cell"
              class:useful={s?.useful}
              class:wasted={s && !s.useful}
              style="--cat: {CATEGORY_COLOR[d.category]}"
              title={s ? `${d.id} · ${s.workload}${s.useful ? '' : ' (wasted)'}` : `${d.id} · idle`}
            ></span>
          {/each}
        </div>
      {/each}
    </div>
  {/each}
</div>

<style>
  .sf {
    display: flex;
    flex-wrap: wrap;
    gap: 7px;
    align-content: flex-start;
  }
  .node {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 5px;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--bg-2);
  }
  .pod {
    --tint: var(--gpu);
    display: flex;
    gap: 3px;
    padding: 4px;
    border-radius: 6px;
    border: 1px solid transparent;
  }
  .pod.isl {
    border-color: color-mix(in oklab, var(--tint) 32%, transparent);
    background: color-mix(in oklab, var(--tint) 7%, transparent);
  }

  .cell {
    width: 15px;
    height: 15px;
    border-radius: 3px;
    /* idle: a faint outline in the device's category color */
    background: color-mix(in oklab, var(--cat) 14%, transparent);
    box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--cat) 30%, transparent);
    transition:
      background 0.15s ease,
      box-shadow 0.15s ease;
  }
  .cell.useful {
    background: var(--bound);
    box-shadow: 0 0 0 1px var(--bound);
  }
  .cell.wasted {
    background: var(--reject);
    box-shadow: 0 0 0 1px var(--reject);
  }
</style>
