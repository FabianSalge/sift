# Structured Parameters

## What it is
The modern, GA DRA path (KEP-4381) for *how device selection happens*. Drivers
publish their devices as ResourceSlices with structured attributes, and the
**kube-scheduler itself** evaluates a ResourceClaim's CEL expressions against
those attributes to pick devices — as part of normal pod scheduling. The driver
does **not** run its own scheduling loop; it mostly publishes good slices and
handles allocation bookkeeping.

## Why it matters for Sift
It decides *where the scheduling logic lives*. In-cluster, Kubernetes' own
scheduler does the selection — so Sift's allocator is **not** the thing running
at bind time on the cluster. The allocator is the reference model: it powers the
benchmark and WASM demo (where there's no kube-scheduler) and is how we design
and validate the matching before expressing it as CEL.

## How Sift uses it
Phase 1 driver = structured-parameters style: publish attribute-rich slices,
express Workload requirements as ResourceClaim CEL, let kube-scheduler match.


## See also
- `docs/concepts/dra.md`, `docs/concepts/capability-aware-scheduling.md`
- ADR-0013