import type { Arrival, Workload } from './types'

const w = (o: Partial<Workload> & { name: string; kind: Workload['kind'] }): Workload => ({
  minMemoryGB: 0,
  requiredPrecisions: [],
  deviceCount: 1,
  sameIsland: false,
  gang: false,
  latencySensitive: false,
  costWeight: 0.5,
  ...o,
})

const train = (name: string): Workload => w({ name, kind: 'train', minMemoryGB: 80, requiredPrecisions: ['bf16'], costWeight: 0.5 })
const big = (name: string): Workload => w({ name, kind: 'train', minMemoryGB: 150, requiredPrecisions: ['bf16'], costWeight: 0.5 })
const infer = (name: string): Workload => w({ name, kind: 'infer', minMemoryGB: 16, requiredPrecisions: ['int8'], costWeight: 1 })
const gang = (name: string): Workload => w({ name, kind: 'train', minMemoryGB: 80, requiredPrecisions: ['bf16'], deviceCount: 4, sameIsland: true, costWeight: 0.7 })

// A curated, deterministic stream over realistic-2026. Training jobs arrive first,
// so legacy's first-fit dumps them onto the (non-trainable) Inferentia2s at the
// head of the fleet — wasted capacity — and fragments the gangs across islands,
// while Sift places them on capable GPUs and keeps the gangs whole. Bursty enough
// that legacy's wasted holds keep real work queued behind them. (Tour-trap
// discipline, ADR-0014.)
export const STREAM: Arrival[] = [
  { at: 0, workload: train('train-1'), duration: 14 },
  { at: 1, workload: train('train-2'), duration: 14 },
  { at: 2, workload: gang('gang-a'), duration: 18 },
  { at: 3, workload: infer('infer-1'), duration: 7 },
  { at: 4, workload: train('train-3'), duration: 14 },
  { at: 5, workload: big('big-1'), duration: 12 },
  { at: 6, workload: infer('infer-2'), duration: 7 },
  { at: 7, workload: train('train-4'), duration: 14 },
  { at: 8, workload: gang('gang-b'), duration: 18 },
  { at: 9, workload: train('train-5'), duration: 14 },
  { at: 10, workload: infer('infer-3'), duration: 7 },
  { at: 11, workload: big('big-2'), duration: 12 },
  { at: 12, workload: train('train-6'), duration: 14 },
  { at: 13, workload: train('train-7'), duration: 14 },
]
