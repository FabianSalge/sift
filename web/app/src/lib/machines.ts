import type { Device } from './types'
import { NO_ISLAND } from './types'

// Machine catalog — data, not logic (CLAUDE.md). Specs mirror
// scenarios/realistic-2026.yaml exactly; if you tune one, mirror the other.
export interface MachineTemplate {
  id: string // catalog key and device-id prefix
  label: string
  count: number
  islanded: boolean // false → standalone devices (island -1)
  device: Omit<Device, 'id' | 'node' | 'island'>
}

export const MACHINE_CATALOG: MachineTemplate[] = [
  {
    id: 'h100', label: '4× H100 · NVLink island', count: 4, islanded: true,
    device: { vendor: 'nvidia', category: 'gpu', memoryGB: 80, precisions: ['bf16', 'fp16', 'fp8'], interconnect: 'nvlink', costPerHr: 2.5, trainable: true },
  },
  {
    id: 'mi300x', label: '4× MI300X · Infinity Fabric', count: 4, islanded: true,
    device: { vendor: 'amd', category: 'gpu', memoryGB: 192, precisions: ['bf16', 'fp16', 'fp8'], interconnect: 'infinity-fabric', costPerHr: 1.9, trainable: true },
  },
  {
    id: 'b200', label: '2× B200 · NVLink island', count: 2, islanded: true,
    device: { vendor: 'nvidia', category: 'gpu', memoryGB: 192, precisions: ['bf16', 'fp16', 'fp8', 'fp4'], interconnect: 'nvlink', costPerHr: 6.0, trainable: true },
  },
  {
    id: 'inferentia2', label: '4× Inferentia2 · standalone', count: 4, islanded: false,
    device: { vendor: 'aws', category: 'infer-asic', memoryGB: 32, precisions: ['fp16', 'int8'], interconnect: 'neuronlink', costPerHr: 0.75, trainable: false },
  },
]

// Keep the canvas legible (and the queueing story visible — a huge fleet
// never backs up).
export const MAX_DEVICES = 64

/** Concrete devices for one new node. The caller supplies fresh node/island
 *  numbers and a per-model serial so ids stay unique for the session. */
export function buildNode(t: MachineTemplate, node: number, island: number, serialStart: number): Device[] {
  return Array.from({ length: t.count }, (_, i) => ({
    ...t.device,
    precisions: [...t.device.precisions],
    id: `${t.id}-${serialStart + i}`,
    node,
    island: t.islanded ? island : NO_ISLAND,
  }))
}
