<script lang="ts">
  import type { Device, Deco } from '../lib/types'
  import { CATEGORY_COLOR, modelName } from '../lib/types'

  let {
    device,
    selected = false,
    deco,
    onselect,
  }: {
    device: Device
    selected?: boolean
    deco?: Deco
    onselect?: (d: Device) => void
  } = $props()

  const RING: Record<string, string> = {
    bound: 'var(--bound)',
    wrong: 'var(--reject)',
    frag: 'var(--accent)',
    select: 'var(--ink)',
  }

  const ring = $derived(selected ? 'select' : deco?.ring)
  const ringColor = $derived(ring ? RING[ring] : null)
</script>

<button
  class="tile"
  class:dim={deco?.dim}
  style="--cat: {CATEGORY_COLOR[device.category]}; {ringColor ? `--ring: ${ringColor}` : ''}"
  class:ringed={!!ringColor}
  onclick={() => onselect?.(device)}
  title={deco?.tag ? `${device.id} → ${deco.tag}` : device.id}
>
  {#if deco?.mark}
    <span class="mark mono" style="--ring: {ringColor ?? 'var(--gpu)'}">{deco.mark}</span>
  {/if}
  <span class="band"></span>
  <span class="body">
    <span class="model mono">{modelName(device.id)}</span>
    <span class="vendor">{device.vendor}</span>
    <span class="mem mono">{device.memoryGB}<i>GB</i></span>
    <span class="cost mono">${device.costPerHr.toFixed(2)}</span>
  </span>
  {#if deco?.tag}
    <span class="tag mono" style="--ring: {ringColor ?? 'var(--bound)'}">{deco.tag}</span>
  {/if}
</button>

<style>
  .tile {
    --cat: var(--gpu);
    --ring: transparent;
    position: relative;
    width: 96px;
    padding: 0;
    border: 1px solid var(--line);
    border-radius: var(--r-sm);
    background: var(--panel);
    color: var(--ink);
    text-align: left;
    cursor: pointer;
    overflow: visible;
    transition:
      transform 0.09s ease,
      border-color 0.12s ease,
      box-shadow 0.12s ease,
      opacity 0.12s ease;
  }
  .tile:hover {
    transform: translateY(-1px);
    border-color: var(--line-strong);
  }
  .tile.ringed {
    border-color: var(--ring);
    box-shadow: 0 0 0 1px var(--ring);
  }
  .tile.dim {
    opacity: 0.34;
  }

  .band {
    display: block;
    height: 3px;
    background: var(--cat);
    border-radius: var(--r-sm) var(--r-sm) 0 0;
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

  .tag {
    display: block;
    font-size: 8px;
    letter-spacing: 0.02em;
    padding: 2px 6px;
    color: #07120a;
    background: var(--ring);
    border-radius: 0 0 var(--r-sm) var(--r-sm);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .mark {
    position: absolute;
    top: -7px;
    left: -7px;
    z-index: 2;
    min-width: 17px;
    height: 17px;
    padding: 0 4px;
    border-radius: 9px;
    background: var(--ring);
    color: #07120a;
    font-size: 9px;
    font-weight: 600;
    display: flex;
    align-items: center;
    justify-content: center;
  }
</style>
