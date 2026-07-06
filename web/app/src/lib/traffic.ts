import type { WorkloadTemplate } from './templates'
import type { Rng } from './rng'
import { expGap, jitter } from './rng'

export interface Due {
  template: WorkloadTemplate
  duration: number
}

// Schedules ambient arrivals per template off one seeded rng. Deterministic
// given the same rng, templates, and sequence of now-values.
export class Generator {
  private sched = new Map<string, { at: number; rate: number }>()
  constructor(private rng: Rng) {}

  /** Arrivals due at or before now. Schedules lazily; redraws on rate change. */
  due(templates: WorkloadTemplate[], now: number): Due[] {
    const out: Due[] = []
    const seen = new Set<string>()
    for (const t of templates) {
      seen.add(t.id)
      if (t.ratePerMin <= 0) {
        this.sched.delete(t.id)
        continue
      }
      let s = this.sched.get(t.id)
      if (!s || s.rate !== t.ratePerMin) {
        s = { at: now + expGap(this.rng, t.ratePerMin), rate: t.ratePerMin }
        this.sched.set(t.id, s)
      }
      while (s.at <= now) {
        out.push({ template: t, duration: jitter(this.rng, t.durationS) })
        s.at += expGap(this.rng, t.ratePerMin)
      }
    }
    for (const id of [...this.sched.keys()]) if (!seen.has(id)) this.sched.delete(id)
    return out
  }
}
