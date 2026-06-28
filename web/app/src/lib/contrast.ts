import type { Summary, Deco } from './types'

// Map a scheduler's placement onto the fleet: each bound device gets a colored
// ring (green = good, red = mis-typed, amber = fragmented gang) tagged with the
// workload it was bound to.
export function contrastDecos(summary: Summary): Map<string, Deco> {
  const m = new Map<string, Deco>()
  for (const o of summary.outcomes) {
    if (o.pending || !o.deviceIDs) continue
    const ring: Deco['ring'] = !o.feasible ? 'wrong' : !o.sameIslandOK ? 'frag' : 'bound'
    for (const id of o.deviceIDs) {
      m.set(id, { ring, tag: o.workload })
    }
  }
  return m
}
