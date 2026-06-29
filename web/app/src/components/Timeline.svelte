<script lang="ts">
  let {
    step,
    total,
    playing,
    label,
    ontoggle,
    onstep,
  }: {
    step: number
    total: number
    playing: boolean
    label: string
    ontoggle: () => void
    onstep: (s: number) => void
  } = $props()

  const atEnd = $derived(step >= total)
</script>

<div class="timeline">
  <button class="play" class:playing aria-label={playing ? 'pause' : 'play'} onclick={ontoggle}>
    {#if playing}
      <span class="ico">❚❚</span>
    {:else}
      <span class="ico tri">▶</span>
    {/if}
  </button>

  <input
    class="scrub"
    type="range"
    min="0"
    max={total}
    step="1"
    value={step}
    oninput={(e) => onstep(+e.currentTarget.value)}
  />

  <span class="readout mono">
    {#if atEnd}all placed{:else}placing {step}/{total}{/if}
  </span>

  <span class="cur mono">{label}</span>
</div>

<style>
  .timeline {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 9px 12px;
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
    box-shadow: 0 1px 0 rgba(0, 0, 0, 0.35);
    transition: transform 0.1s ease;
  }
  .play:hover {
    transform: scale(1.06);
  }
  .ico {
    font-size: 10px;
    line-height: 1;
  }
  .ico.tri {
    font-size: 11px;
    margin-left: 2px;
  }

  .scrub {
    flex: 1;
    min-width: 120px;
    accent-color: var(--gpu);
    cursor: pointer;
  }

  .readout {
    flex: none;
    font-size: 11px;
    color: var(--ink-dim);
    min-width: 78px;
  }

  .cur {
    flex: none;
    font-size: 11px;
    color: var(--ink-faint);
    min-width: 72px;
    text-align: right;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
