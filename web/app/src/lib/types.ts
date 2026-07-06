// Mirrors the camelCase DTOs the WASM engine emits (web/wasm/engine/dto.go).

export type Vendor = 'nvidia' | 'amd' | 'google' | 'aws'
export type Category = 'gpu' | 'train-asic' | 'infer-asic'
export type Precision = 'bf16' | 'fp16' | 'fp8' | 'fp4' | 'int8'
export type Interconnect = 'nvlink' | 'infinity-fabric' | 'ici' | 'neuronlink' | 'none'
export type Kind = 'train' | 'infer'

export const NO_ISLAND = -1

export interface Device {
  id: string
  node: number
  island: number
  vendor: Vendor
  category: Category
  memoryGB: number
  precisions: Precision[]
  interconnect: Interconnect
  costPerHr: number
  trainable: boolean
}

export interface Workload {
  name: string
  kind: Kind
  minMemoryGB: number
  requiredPrecisions: Precision[]
  deviceCount: number
  sameIsland: boolean
  gang: boolean
  latencySensitive: boolean
  costWeight: number
}

export interface Reason {
  code: string
  detail: string
}

export interface Score {
  costComponent: number
  memoryWaste: number
}

export interface Verdict {
  deviceID: string
  feasible: boolean
  allocated: boolean
  reasons: Reason[]
  score: Score
  rank: number
}

export interface Trace {
  workload: string
  verdicts: Verdict[]
  bound: string[] | null
  island: number
  err: string
}

// ── display helpers ──────────────────────────────────────────────────────
export const CATEGORY_COLOR: Record<Category, string> = {
  gpu: 'var(--gpu)',
  'train-asic': 'var(--train)',
  'infer-asic': 'var(--infer)',
}

export const CATEGORY_LABEL: Record<Category, string> = {
  gpu: 'GPU',
  'train-asic': 'train-ASIC',
  'infer-asic': 'infer-ASIC',
}

/** Model name from a device id: "h100-0" -> "H100", "mi300x-3" -> "MI300X". */
export const modelName = (id: string): string => id.replace(/-\d+$/, '').toUpperCase()

// ── per-device decoration applied over the fleet (placement / explanation) ──
export type Ring = 'bound' | 'wrong' | 'frag' | 'select'

export interface Deco {
  ring?: Ring // colored outline: bound=green, wrong=red, frag=amber, select=white
  tag?: string // bottom strip text, e.g. the bound workload name
  mark?: string // small corner badge, e.g. a rank "#1" or "✓"
  dim?: boolean // faded — filtered out (Explain mode)
  reason?: string // hover tooltip, e.g. why a device was rejected
  pulse?: boolean // briefly emphasize — the device just placed in the timeline
}

// ── live cluster session (mirrors web/wasm/engine ClusterSnapshotDTO) ───────
export interface ClusterDevice extends Device {
  jobID: number // -1 when idle
  draining: boolean
}

export interface ClusterJob {
  id: number
  workload: Workload
  duration: number
  arrivedAt: number
  placedAt: number // -1 while queued
  end: number // -1 while queued
  deviceIDs: string[] | null
  useful: boolean
  costPerHr: number
}

export interface ClusterEvent {
  kind: 'placed' | 'completed' | 'node-removed'
  at: number
  jobID: number
  node: number
  deviceIDs: string[] | null
}

export interface ShadowMetrics {
  busy: number
  wasted: number
  queue: number
  usefulDone: number
  cost: number
}

export interface ClusterSnapshot {
  clock: number
  devices: ClusterDevice[]
  queue: ClusterJob[]
  running: ClusterJob[]
  usefulDone: number
  cost: number
  events: ClusterEvent[] | null
  shadow: ShadowMetrics
}
