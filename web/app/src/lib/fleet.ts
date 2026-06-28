import type { Device, Interconnect } from './types'
import { NO_ISLAND } from './types'

// A pod is a set of devices that share an interconnect island within one node
// (or a standalone group for island === NO_ISLAND).
export interface Pod {
  island: number
  interconnect: Interconnect
  devices: Device[]
}

export interface NodeGroup {
  node: number
  pods: Pod[]
}

/** Group a flat fleet into nodes → island pods, preserving fleet order within a
 *  pod. Islands sort ascending (standalone, island -1, sorts first). */
export function groupFleet(devices: Device[]): NodeGroup[] {
  const byNode = new Map<number, Device[]>()
  for (const d of devices) {
    const arr = byNode.get(d.node) ?? []
    arr.push(d)
    byNode.set(d.node, arr)
  }

  const nodes: NodeGroup[] = []
  for (const node of [...byNode.keys()].sort((a, b) => a - b)) {
    const byIsland = new Map<number, Device[]>()
    for (const d of byNode.get(node)!) {
      const arr = byIsland.get(d.island) ?? []
      arr.push(d)
      byIsland.set(d.island, arr)
    }
    const pods: Pod[] = [...byIsland.keys()]
      .sort((a, b) => a - b)
      .map((island) => ({
        island,
        interconnect: byIsland.get(island)![0].interconnect,
        devices: byIsland.get(island)!,
      }))
    nodes.push({ node, pods })
  }
  return nodes
}

export const isStandalone = (island: number): boolean => island === NO_ISLAND
