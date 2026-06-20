# Capability-Aware Scheduling (vs the Integer Model)

## What it is
Two ways to schedule hardware:
- **Integer / device-plugin model** — hardware is an opaque count
  (`nvidia.com/gpu: N`). The scheduler can find *enough free things* but is blind
  to type, capability, cost, and location. No notion of "the right kind".
- **Capability-aware (DRA) model** — devices advertise attributes; workloads
  request by attribute. The scheduler matches *requirements to capabilities*.

The matching shape is **filter → score → bind**:
- **Filter** (hard constraints): keep only free devices that *can* run the job —
  trainable if it's a training job, enough memory, and required precisions a
  subset of what the device supports.
- **Score** (soft preferences): rank the survivors by the lexicographic key
  `(CostWeight·CostPerHr, memory waste, ID)` — cheapest fitting when the job is
  cost-sensitive, otherwise best-fit packing, with a stable ID tiebreak.
- **Bind**: take the best and mark it allocated. If the filter empties the pool,
  the job is **Pending** — the correct outcome, not a crash.

For a multi-device job the scheduler binds *all* requested devices or none. A
same-island gang must come from one interconnect island: it picks the cheapest
island that can hold the whole gang, then the best devices within it.

The integer model has, at best, a crude filter and **no score step** — which is
exactly why it mis-matches type, overspends, and fragments.

## Why it matters for Sift
This contrast is the entire project. Sift implements both schedulers; the
benchmark shows the integer model failing in three ways the capability-aware one
doesn't:
- **Type** — a training job lands on a non-trainable inference ASIC.
- **Cost** — a cost-sensitive job takes the first expensive device, not the
  cheapest one that fits.
- **Topology** — a same-island gang fragments across islands.

## How Sift implements it
`allocator` is the pure decision core. `SiftScheduler` does filter → score →
bind; `LegacyScheduler` is first-fit with no score step. Each failure mode has a
test asserting Sift catches it and legacy misses it (`allocator/*_test.go`). For
the precise algorithm see the package doc: `go doc ./allocator`.

## Gotcha / what confused me
- **Pending is the correct outcome, not a failure.** When nothing fits, returning
  "no feasible device" *is* the scheduler working — it refuses to mis-place rather
  than force a bad binding (which is exactly what legacy does).
- **Fleet order is load-bearing for the demo.** Legacy is first-fit, so placing a
  tempting wrong device *early* is what makes it visibly fail; reorder the fleet
  and the trap stops biting.

## See also
- `docs/concepts/dra.md`, `docs/concepts/heterogeneous-accelerators.md`,
  `docs/concepts/topology-islands.md`
- `allocator/` — the implementation and its tests.
