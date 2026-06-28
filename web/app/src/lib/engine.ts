// Loads the Go WASM engine and wraps its JS globals in typed, promise-returning
// calls. All scheduling logic lives in the wasm (the pure allocator core); this
// module only loads it and marshals strings. See web/wasm + ADR-0003.

import type { Device, Workload, Report, Trace } from './types'

const BASE = import.meta.env.BASE_URL

interface Envelope {
  ok: boolean
  data?: string
  error?: string
}

declare global {
  // Set by wasm_exec.js and the Go wasm main respectively.
  // eslint-disable-next-line no-var
  var Go: { new (): { importObject: WebAssembly.Imports; run(i: WebAssembly.Instance): void } }
  function siftLoadScenario(yaml: string): Envelope
  function siftRun(fleetJSON: string, workloadsJSON: string): Envelope
  function siftExplain(fleetJSON: string, workloadJSON: string, allocatedJSON: string): Envelope
}

let ready: Promise<void> | null = null

function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const s = document.createElement('script')
    s.src = src
    s.onload = () => resolve()
    s.onerror = () => reject(new Error(`failed to load ${src}`))
    document.head.appendChild(s)
  })
}

async function instantiate(go: InstanceType<typeof globalThis.Go>): Promise<WebAssembly.Instance> {
  const url = `${BASE}wasm/app.wasm`
  try {
    const res = await WebAssembly.instantiateStreaming(fetch(url), go.importObject)
    return res.instance
  } catch {
    // Fallback if the server doesn't send application/wasm.
    const buf = await (await fetch(url)).arrayBuffer()
    const res = await WebAssembly.instantiate(buf, go.importObject)
    return res.instance
  }
}

/** Idempotent: loads wasm_exec.js + app.wasm once and starts the Go runtime. */
export function initEngine(): Promise<void> {
  if (!ready) {
    ready = (async () => {
      await loadScript(`${BASE}wasm/wasm_exec.js`)
      const go = new globalThis.Go()
      const instance = await instantiate(go)
      go.run(instance) // do not await — main() registers the globals then parks
    })()
  }
  return ready
}

function call(fn: Envelope, label: string): unknown {
  if (!fn.ok) throw new Error(`${label}: ${fn.error ?? 'unknown error'}`)
  return JSON.parse(fn.data ?? 'null')
}

export async function loadScenario(yaml: string): Promise<Device[]> {
  await initEngine()
  return call(siftLoadScenario(yaml), 'loadScenario') as Device[]
}

export async function run(fleet: Device[], workloads: Workload[]): Promise<Report> {
  await initEngine()
  return call(siftRun(JSON.stringify(fleet), JSON.stringify(workloads)), 'run') as Report
}

export async function explain(
  fleet: Device[],
  workload: Workload,
  allocated: Record<string, boolean> | null = null,
): Promise<Trace> {
  await initEngine()
  return call(
    siftExplain(JSON.stringify(fleet), JSON.stringify(workload), JSON.stringify(allocated)),
    'explain',
  ) as Trace
}
