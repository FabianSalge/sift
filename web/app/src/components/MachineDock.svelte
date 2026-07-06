<script lang="ts">
  import { MACHINE_CATALOG, MAX_DEVICES, type MachineTemplate } from '../lib/machines'

  let { deviceCount, onadd }: { deviceCount: number; onadd: (t: MachineTemplate) => void } = $props()
</script>

<section class="sec">
  <div class="label">machines</div>
  {#each MACHINE_CATALOG as t (t.id)}
    {@const full = deviceCount + t.count > MAX_DEVICES}
    <div class="mcard">
      <div class="mi">
        <span class="name mono">{t.label}</span>
        <span class="sub mono">${t.device.costPerHr.toFixed(2)}/h each</span>
      </div>
      <button
        class="add"
        disabled={full}
        title={full ? `fleet cap ${MAX_DEVICES} devices` : 'add node'}
        onclick={() => onadd(t)}>+</button
      >
    </div>
  {/each}
  <p class="hint">drain a node from its card in the canvas — it finishes its jobs, then leaves.</p>
</section>

<style>
  .sec { display: flex; flex-direction: column; gap: 8px; }

  .mcard {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    border: 1px solid var(--line);
    border-radius: var(--r-md);
    background: var(--panel);
    padding: 9px 12px;
  }
  .mi { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
  .name { font-size: 11px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .sub { font-size: 9.5px; color: var(--ink-faint); }

  .add {
    appearance: none;
    flex: none;
    width: 26px;
    height: 26px;
    border-radius: 7px;
    border: 1px solid color-mix(in oklab, var(--bound) 45%, transparent);
    background: color-mix(in oklab, var(--bound) 12%, transparent);
    color: var(--bound);
    font-size: 15px;
    line-height: 1;
    cursor: pointer;
  }
  .add:hover:not(:disabled) { background: color-mix(in oklab, var(--bound) 22%, transparent); }
  .add:disabled { opacity: 0.35; cursor: not-allowed; }

  .hint { margin: 2px 0 0; font-size: 10px; line-height: 1.5; color: var(--ink-faint); }
</style>
