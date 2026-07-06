// Deterministic PRNG so ambient traffic replays exactly per ?seed=.
export type Rng = () => number

export function mulberry32(seed: number): Rng {
  let a = seed >>> 0
  return () => {
    a = (a + 0x6d2b79f5) | 0
    let t = Math.imul(a ^ (a >>> 15), 1 | a)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

/** Exponential inter-arrival gap (sim-seconds) for a rate of perMin per minute. */
export const expGap = (rng: Rng, perMin: number): number => -Math.log(1 - rng()) * (60 / perMin)

/** ±20% duration jitter so the cluster never looks metronomic. */
export const jitter = (rng: Rng, base: number): number => base * (0.8 + 0.4 * rng())
