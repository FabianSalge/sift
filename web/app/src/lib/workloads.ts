import type { Workload } from './types'

export interface Preset {
  id: string
  label: string
  caption: string
  workloads: Workload[]
}

// Fill the unset Workload fields so presets stay terse.
const w = (
  o: Partial<Workload> & { name: string; kind: Workload['kind'] },
): Workload => ({
  minMemoryGB: 0,
  requiredPrecisions: [],
  deviceCount: 1,
  sameIsland: false,
  gang: false,
  latencySensitive: false,
  costWeight: 0.5,
  ...o,
})

// Curated mixes. The tour presets each isolate one failure mode; "three-modes"
// is the headline run that exercises all three at once (mirrors cmd/bench).
export const PRESETS: Preset[] = [
  {
    id: 'three-modes',
    label: 'Three failure modes',
    caption: 'Type-rejection, cost, and topology — all in one run.',
    workloads: [
      w({ name: 'train-llm', kind: 'train', minMemoryGB: 80, requiredPrecisions: ['bf16'], costWeight: 0.5 }),
      w({ name: 'train-big', kind: 'train', minMemoryGB: 150, requiredPrecisions: ['bf16'], costWeight: 0.5 }),
      w({ name: 'infer-fp8', kind: 'infer', minMemoryGB: 16, requiredPrecisions: ['fp8'], costWeight: 1 }),
      w({ name: 'infer-int8', kind: 'infer', minMemoryGB: 16, requiredPrecisions: ['int8'], costWeight: 1 }),
      w({ name: 'gang-train', kind: 'train', minMemoryGB: 80, requiredPrecisions: ['bf16'], deviceCount: 4, sameIsland: true, costWeight: 0.7 }),
    ],
  },
  {
    id: 'type',
    label: 'Type-rejection',
    caption: 'A bf16 training job. Legacy grabs the first free device — a non-trainable inference ASIC.',
    workloads: [w({ name: 'train-llm', kind: 'train', minMemoryGB: 80, requiredPrecisions: ['bf16'], costWeight: 0.5 })],
  },
  {
    id: 'cost',
    label: 'Cost',
    caption: 'A cost-sensitive inference job. Legacy takes the first fit; Sift takes the cheapest fit.',
    workloads: [w({ name: 'infer-int8', kind: 'infer', minMemoryGB: 16, requiredPrecisions: ['int8'], costWeight: 1 })],
  },
  {
    id: 'topology',
    label: 'Topology gang',
    caption: 'A 4-device same-island training gang. Legacy fragments it across islands; Sift keeps it whole.',
    workloads: [w({ name: 'gang-train', kind: 'train', minMemoryGB: 80, requiredPrecisions: ['bf16'], deviceCount: 4, sameIsland: true, costWeight: 0.7 })],
  },
]

// Explain mode walks one workload at a time; reuse the headline mix's five.
export const EXPLAIN_WORKLOADS: Workload[] = PRESETS[0].workloads
