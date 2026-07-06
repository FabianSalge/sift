<script lang="ts">
  let {
    paused,
    speed,
    clock,
    seed,
    ontoggle,
    onspeed,
  }: {
    paused: boolean
    speed: number
    clock: number
    seed: number
    ontoggle: () => void
    onspeed: (v: number) => void
  } = $props()
</script>

<div class="transport">
  <button class="play" aria-label={paused ? 'play' : 'pause'} onclick={ontoggle}>
    {#if paused}<span class="ico tri">▶</span>{:else}<span class="ico">❚❚</span>{/if}
  </button>
  <span class="clock mono">t = {clock.toFixed(1)}s</span>
  <label class="speed">
    <span class="k">speed ×{speed}</span>
    <input type="range" min="1" max="20" step="1" value={speed} oninput={(e) => onspeed(+e.currentTarget.value)} />
  </label>
  <span class="seed mono" title="ambient-traffic seed — add ?seed= to replay a run">seed {seed}</span>
</div>

<style>
  .transport {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 8px 14px;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: var(--panel);
  }
  .play {
    flex: none;
    width: 30px;
    height: 30px;
    border-radius: 50%;
    border: none;
    cursor: pointer;
    background: linear-gradient(180deg, color-mix(in oklab, var(--gpu) 92%, white), var(--gpu));
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .ico { font-size: 10px; }
  .ico.tri { font-size: 11px; margin-left: 2px; }
  .clock { font-size: 12px; color: var(--ink-dim); min-width: 74px; }
  .speed { display: flex; align-items: center; gap: 7px; }
  .speed .k {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--ink-faint);
    min-width: 64px;
  }
  .speed input { width: 90px; accent-color: var(--gpu); cursor: pointer; }
  .seed { font-size: 10px; color: var(--ink-faint); }
</style>
