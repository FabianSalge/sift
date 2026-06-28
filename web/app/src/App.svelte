<script lang="ts">
  // Phase 3 scaffold — engine wiring + the fleet canvas land in the next increments.
  const version = '0.1.0'
</script>

<div class="frame">
  <header class="masthead">
    <div class="brand">
      <span class="mark" aria-hidden="true"></span>
      <span class="word">sift</span>
      <span class="ver mono">v{version}</span>
    </div>
    <p class="tag">
      capability- &amp; topology-aware accelerator scheduling — <span class="dim">a runnable model</span>
    </p>
  </header>

  <main class="stage">
    <div class="placeholder">
      <span class="label">canvas</span>
      <p>Fleet, contrast, and the filter&nbsp;→&nbsp;score&nbsp;→&nbsp;bind explainer mount here.</p>
    </div>
  </main>
</div>

<style>
  .frame {
    max-width: 1180px;
    margin: 0 auto;
    padding: 40px 28px 80px;
  }

  .masthead {
    border-bottom: 1px solid var(--line);
    padding-bottom: 22px;
  }

  .brand {
    display: flex;
    align-items: baseline;
    gap: 11px;
  }

  .mark {
    width: 13px;
    height: 13px;
    border-radius: 3px;
    align-self: center;
    background: linear-gradient(135deg, var(--gpu), var(--infer));
    box-shadow: 0 0 18px rgba(99, 102, 241, 0.5);
  }

  .word {
    font-family: var(--font-mono);
    font-weight: 600;
    font-size: 22px;
    letter-spacing: 0.02em;
  }

  .ver {
    font-size: 11px;
    color: var(--ink-faint);
  }

  .tag {
    margin: 12px 0 0;
    font-size: 14px;
    color: var(--ink-dim);
  }
  .tag .dim {
    color: var(--ink-faint);
  }

  .stage {
    margin-top: 28px;
  }

  .placeholder {
    border: 1px dashed var(--line-strong);
    border-radius: var(--r-lg);
    padding: 64px 28px;
    text-align: center;
    color: var(--ink-faint);
    background: var(--bg-2);
  }
  .placeholder p {
    margin: 10px 0 0;
    font-size: 13px;
  }
</style>
