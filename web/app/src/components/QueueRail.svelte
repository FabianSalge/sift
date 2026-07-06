<script lang="ts">
  import type { ClusterJob } from '../lib/types'

  let { queue, onselect }: { queue: ClusterJob[]; onselect?: (j: ClusterJob) => void } = $props()
  const shown = $derived(queue.slice(0, 8))
</script>

<div class="rail">
  <span class="label">queue</span>
  {#each shown as j (j.id)}
    <button class="job mono" class:gang={j.workload.deviceCount > 1} onclick={() => onselect?.(j)}>
      {j.workload.name}{j.workload.deviceCount > 1 ? ` ×${j.workload.deviceCount}` : ''}
    </button>
  {/each}
  {#if queue.length > 8}<span class="more mono">+{queue.length - 8}</span>{/if}
  {#if queue.length === 0}<span class="empty mono">empty — cluster keeping up</span>{/if}
</div>

<style>
  .rail {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    min-height: 30px;
  }
  .job {
    appearance: none;
    font-size: 11px;
    padding: 3px 9px;
    border-radius: 6px;
    background: var(--panel);
    border: 1px solid var(--line);
    color: var(--ink-dim);
    cursor: pointer;
  }
  .job:hover { color: var(--ink); border-color: var(--line-strong); }
  .job.gang {
    border-color: color-mix(in oklab, var(--train) 45%, transparent);
    color: var(--ink);
  }
  .more, .empty { font-size: 11px; color: var(--ink-faint); }
</style>
