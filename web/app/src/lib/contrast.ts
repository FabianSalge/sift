import type { Summary, Deco } from './types'

// Map a scheduler's placement onto the fleet: each bound device gets a colored
// ring (green = good, red = mis-typed, amber = fragmented gang) tagged with the
// workload it was bound to. `limit` reveals only the first N workloads of the
// queue — the placement timeline animates by raising it; the most-recently
// placed workload pulses while the animation is mid-flight.
export function contrastDecos(summary: Summary, limit = Number.POSITIVE_INFINITY): Map<string, Deco> {
  const m = new Map<string, Deco>()
  const shown = summary.outcomes.slice(0, limit)
  const animating = limit < summary.outcomes.length

  shown.forEach((o, i) => {
    if (o.pending || !o.deviceIDs) return
    const ring: Deco['ring'] = !o.feasible ? 'wrong' : !o.sameIslandOK ? 'frag' : 'bound'
    const pulse = animating && i === shown.length - 1
    for (const id of o.deviceIDs) {
      m.set(id, { ring, tag: o.workload, pulse })
    }
  })
  return m
}
