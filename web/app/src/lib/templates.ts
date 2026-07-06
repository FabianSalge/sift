import type { Workload } from './types'

// A workload template is the user's unit of creation: a reusable job spec
// plus a duration and an ambient arrival rate (0 = manual bursts only).
export interface WorkloadTemplate {
  id: string // template name and job-name prefix
  workload: Omit<Workload, 'name'>
  durationS: number
  ratePerMin: number
}

const spec = (
  o: Partial<Omit<Workload, 'name'>> & { kind: Workload['kind'] },
): Omit<Workload, 'name'> => ({
  minMemoryGB: 0,
  requiredPrecisions: [],
  deviceCount: 1,
  sameIsland: false,
  gang: false,
  latencySensitive: false,
  costWeight: 0.5,
  ...o,
})

// Boot presets: the three failure-mode archetypes plus a big-memory trainer,
// at rates that keep roughly half the realistic-2026 fleet busy, with bursts.
export const DEFAULT_TEMPLATES: WorkloadTemplate[] = [
  { id: 'train-llm', workload: spec({ kind: 'train', minMemoryGB: 80, requiredPrecisions: ['bf16'] }), durationS: 75, ratePerMin: 2 },
  { id: 'big-train', workload: spec({ kind: 'train', minMemoryGB: 150, requiredPrecisions: ['bf16'] }), durationS: 100, ratePerMin: 0.6 },
  { id: 'infer-int8', workload: spec({ kind: 'infer', minMemoryGB: 16, requiredPrecisions: ['int8'], costWeight: 1 }), durationS: 20, ratePerMin: 6 },
  { id: 'gang-train', workload: spec({ kind: 'train', minMemoryGB: 80, requiredPrecisions: ['bf16'], deviceCount: 4, sameIsland: true, costWeight: 0.7 }), durationS: 120, ratePerMin: 0.4 },
]

export const blankTemplate = (id: string): WorkloadTemplate => ({
  id,
  workload: spec({ kind: 'train', minMemoryGB: 80, requiredPrecisions: ['bf16'] }),
  durationS: 60,
  ratePerMin: 0,
})
