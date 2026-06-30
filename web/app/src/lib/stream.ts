import type { SchedulerSim } from './types'

export interface DeviceState {
  workload: string
  useful: boolean
}

export interface ClusterState {
  busy: Map<string, DeviceState> // deviceID → what's running on it at time t
  busyCount: number
  wastedCount: number // busy devices doing no useful work (legacy's bad holds)
  queue: number // arrivals admitted but not yet running/done at t
  usefulDone: number // useful jobs completed by t
  cost: number // cumulative device-cost accrued by t
}

// Derive a scheduler's cluster state at simulation time t from its per-arrival
// timeline — the precomputed result the UI scrubs over.
export function clusterState(sim: SchedulerSim, t: number): ClusterState {
  const busy = new Map<string, DeviceState>()
  let busyCount = 0
  let wastedCount = 0
  let queue = 0
  let usefulDone = 0
  let cost = 0

  for (const a of sim.arrivals) {
    const running = a.placedAt >= 0 && a.placedAt <= t && t < a.end
    if (running) {
      const ids = a.deviceIDs ?? []
      for (const id of ids) busy.set(id, { workload: a.workload, useful: a.useful })
      busyCount += ids.length
      if (!a.useful) wastedCount += ids.length
    }
    if (a.arrivedAt <= t && (a.placedAt < 0 || a.placedAt > t)) queue++
    if (a.useful && a.end >= 0 && a.end <= t) usefulDone++
    if (a.placedAt >= 0 && a.placedAt <= t) {
      cost += a.costPerHr * (Math.min(t, a.end) - a.placedAt)
    }
  }

  return { busy, busyCount, wastedCount, queue, usefulDone, cost }
}
