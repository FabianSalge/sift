import type { Trace, Deco } from './types'

export type Stage = 'filter' | 'score' | 'bind'
export const STAGES: Stage[] = ['filter', 'score', 'bind']
export const STAGE_LABEL: Record<Stage, string> = {
  filter: 'Filter',
  score: 'Score',
  bind: 'Bind',
}

// Decorate the fleet for one stage of a trace:
//   filter — rejected devices dim and carry a ✗ + their reason; survivors plain
//   score  — feasible devices show their rank badge; rejected stay dim
//   bind   — the bound devices glow green (tagged); everything else dims
export function explainDecos(trace: Trace, stage: Stage): Map<string, Deco> {
  const m = new Map<string, Deco>()
  const bound = new Set(trace.bound ?? [])

  for (const v of trace.verdicts) {
    if (!v.feasible) {
      m.set(v.deviceID, {
        dim: true,
        mark: stage === 'filter' ? '✗' : undefined,
        reason: v.reasons.map((r) => r.detail).join('; '),
      })
      continue
    }
    if (stage === 'filter') continue // feasible survivor — leave it plain
    if (stage === 'score') {
      m.set(v.deviceID, { mark: `#${v.rank}` })
      continue
    }
    // bind
    if (bound.has(v.deviceID)) m.set(v.deviceID, { ring: 'bound', tag: trace.workload })
    else m.set(v.deviceID, { dim: true })
  }
  return m
}
