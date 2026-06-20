# Capability-Aware Scheduling (vs the Integer Model)

## What it is
Two ways to schedule hardware:
- **Integer / device-plugin model** — hardware is an opaque count
  (`nvidia.com/gpu: N`). The scheduler can find *enough free things* but is blind
  to type, capability, cost, and location. No notion of "the right kind".
- **Capability-aware (DRA) model** — devices advertise attributes; workloads
  request by attribute. The scheduler matches *requirements to capabilities*.

The matching shape is **filter → score → bind**:
- **Filter** (hard constraints): keep only devices that *can* run the job —
  right category/trainability, enough memory, required precisions ⊆ device set,
  same-island feasible.
- **Score** (soft preferences): rank the survivors — cheapest first (weighted by
  `CostWeight`), then packing/locality.
- **Bind**: take the best; if the filter empties the pool, the job is Pending.

The integer model has, at best, a crude filter and **no score step** — which is
exactly why it mis-matches type, overspends, and fragments.

## Why it matters for Sift
This contrast is the entire project. Sift implements both schedulers; the
benchmark shows the integer model failing in three ways the capability-aware one
doesn't: wrong type, wrong cost, wrong topology.

## How Sift uses it
`allocator` implements legacy (first-fit, no score) and Sift (filter-then-score).
Tests assert each failure mode: Sift catches it, legacy misses it.

## Gotcha / what confused me
_(your turn — e.g. "Pending is the correct outcome, not a failure"; soft vs hard
constraints; why first-fit order matters for the demo)_

## See also
- `docs/concepts/dra.md`, `docs/concepts/heterogeneous-accelerators.md`
- ADR-0007 (filter-then-score), ADR-0014 (witnessing failures)