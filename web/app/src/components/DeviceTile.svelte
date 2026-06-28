<script lang="ts">
  import type { Device } from '../lib/types'
  import { CATEGORY_COLOR, modelName } from '../lib/types'

  let {
    device,
    selected = false,
    onselect,
  }: {
    device: Device
    selected?: boolean
    onselect?: (d: Device) => void
  } = $props()
</script>

<button
  class="tile"
  class:selected
  style="--cat: {CATEGORY_COLOR[device.category]}"
  onclick={() => onselect?.(device)}
  title={device.id}
>
  <span class="band"></span>
  <span class="body">
    <span class="model mono">{modelName(device.id)}</span>
    <span class="vendor">{device.vendor}</span>
    <span class="mem mono">{device.memoryGB}<i>GB</i></span>
    <span class="cost mono">${device.costPerHr.toFixed(2)}</span>
  </span>
</button>

<style>
  .tile {
    --cat: var(--gpu);
    position: relative;
    width: 96px;
    padding: 0;
    border: 1px solid var(--line);
    border-radius: var(--r-sm);
    background: var(--panel);
    color: var(--ink);
    text-align: left;
    cursor: pointer;
    overflow: hidden;
    transition:
      transform 0.09s ease,
      border-color 0.12s ease,
      box-shadow 0.12s ease;
  }
  .tile:hover {
    transform: translateY(-1px);
    border-color: var(--line-strong);
  }
  .tile.selected {
    border-color: var(--ink);
    box-shadow: 0 0 0 1px var(--ink);
  }

  .band {
    display: block;
    height: 3px;
    background: var(--cat);
  }

  .body {
    display: block;
    padding: 7px 8px 8px;
  }

  .model {
    display: block;
    font-size: 12px;
    font-weight: 600;
    line-height: 1.05;
    letter-spacing: 0.01em;
  }
  .vendor {
    display: block;
    font-size: 8px;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--ink-faint);
    margin-top: 1px;
  }
  .mem {
    display: block;
    font-size: 15px;
    font-weight: 500;
    margin-top: 7px;
  }
  .mem i {
    font-style: normal;
    font-size: 9px;
    color: var(--ink-faint);
    margin-left: 1px;
  }
  .cost {
    display: block;
    font-size: 10px;
    color: var(--ink-dim);
    margin-top: 1px;
  }
</style>
