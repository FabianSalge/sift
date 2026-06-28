<script lang="ts">
  import type { Device } from '../lib/types'
  import { CATEGORY_COLOR, CATEGORY_LABEL, modelName, NO_ISLAND } from '../lib/types'

  let { device }: { device: Device | null } = $props()
</script>

<aside class="rail">
  {#if device}
    <div class="hd">
      <span class="sw" style="background: {CATEGORY_COLOR[device.category]}"></span>
      <div>
        <h3 class="mono">{modelName(device.id)}</h3>
        <div class="id mono">{device.id} · node {device.node}</div>
      </div>
    </div>
    <dl>
      <div><dt>vendor</dt><dd>{device.vendor}</dd></div>
      <div><dt>category</dt><dd>{CATEGORY_LABEL[device.category]}</dd></div>
      <div><dt>memory</dt><dd class="mono">{device.memoryGB} GB</dd></div>
      <div><dt>precisions</dt><dd class="mono">{device.precisions.join(' · ')}</dd></div>
      <div><dt>interconnect</dt><dd>{device.interconnect}</dd></div>
      <div><dt>island</dt><dd class="mono">{device.island === NO_ISLAND ? '—' : device.island}</dd></div>
      <div><dt>cost</dt><dd class="mono">${device.costPerHr.toFixed(2)}/hr</dd></div>
      <div><dt>trainable</dt><dd>{device.trainable ? 'yes' : 'no'}</dd></div>
    </dl>
  {:else}
    <div class="empty">
      <span class="label">device</span>
      <p>Select a device to inspect its capabilities.</p>
    </div>
  {/if}
</aside>

<style>
  .rail {
    width: 184px;
    flex: none;
    background: var(--panel);
    border: 1px solid var(--line);
    border-radius: var(--r-lg);
    padding: 14px;
    position: sticky;
    top: 16px;
  }

  .hd {
    display: flex;
    gap: 9px;
    align-items: flex-start;
    margin-bottom: 12px;
  }
  .sw {
    width: 10px;
    height: 10px;
    border-radius: 3px;
    margin-top: 3px;
    flex: none;
  }
  h3 {
    font-size: 14px;
  }
  .id {
    font-size: 9px;
    color: var(--ink-faint);
    margin-top: 2px;
  }

  dl {
    margin: 0;
  }
  dl > div {
    display: flex;
    justify-content: space-between;
    gap: 10px;
    padding: 5px 0;
    border-bottom: 1px solid var(--line);
    font-size: 11px;
  }
  dl > div:last-child {
    border-bottom: none;
  }
  dt {
    color: var(--ink-faint);
  }
  dd {
    margin: 0;
    text-align: right;
    color: var(--ink);
  }

  .empty {
    color: var(--ink-faint);
  }
  .empty p {
    margin: 8px 0 0;
    font-size: 12px;
    line-height: 1.5;
  }
</style>
