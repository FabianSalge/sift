// Package allocator is the pure decision core of Sift: given a heterogeneous
// pool of accelerators, it places each workload on the right kind of device, in
// the right place, at the right cost. It has no Kubernetes or YAML dependencies,
// so the same logic backs the DRA driver, the WASM demo, and the benchmark.
//
// # Model
//
// A [Device] is a static capability advertisement — the analogue of a DRA
// ResourceSlice: vendor, category, accelerator memory, supported precisions,
// interconnect, island, hourly cost, and trainability. It carries no allocation
// state. A [Workload] is a request — the analogue of a ResourceClaim: how many
// devices, the minimum memory and required precisions, and topology/cost
// preferences.
//
// # Scheduling: filter, score, bind
//
// [SiftScheduler] places a workload in three steps:
//
//   - Filter (hard constraints): keep only free devices that can run the job —
//     trainable for a training job, enough memory, and required precisions a
//     subset of the device's set.
//   - Score (soft preferences): rank the survivors by the lexicographic key
//     (CostWeight*CostPerHr, memory waste, ID) — cheapest fitting when the job is
//     cost-sensitive, otherwise best-fit packing, with a deterministic ID
//     tiebreak.
//   - Bind: take the best and mark it allocated so it cannot be reused. When
//     nothing fits, the workload is Pending ([ErrNoFeasibleDevice]).
//
// # Multiple devices and topology
//
// A workload may need several devices (DeviceCount); the scheduler binds all of
// them or none — there is no partial placement. When SameIsland is set, the whole
// gang must come from one interconnect island: the scheduler picks the cheapest
// island that can hold the entire request, then the best devices within it.
// Otherwise it takes the best devices globally.
//
// # The legacy contrast
//
// [LegacyScheduler] models the integer device-plugin (nvidia.com/gpu: N): it
// hands out the first free devices in fleet order, blind to type, memory,
// precision, cost, and island. That blindness is the point — it mis-types,
// overspends, and fragments exactly where Sift does not, which is what the
// benchmark and the failure-mode tests demonstrate.
package allocator
