# Dynamic Resource Allocation (DRA)

## What it is
The modern Kubernetes mechanism for requesting and allocating specialized
hardware (GPUs, accelerators, NICs). It replaces the old **device plugin** model,
where hardware was an opaque integer count (`nvidia.com/gpu: 1`), with a richer
model where devices *advertise structured attributes* and workloads *request by
those attributes*. Reached GA in Kubernetes 1.34.

## Why it matters for Sift
DRA is the whole reason Sift is possible. Heterogeneous scheduling needs devices
to describe *what they are* (type, memory, precision, cost) and workloads to ask
for *what they need* — which the integer model literally cannot express. DRA is
the substrate that makes capability-aware scheduling expressible in Kubernetes.

## How Sift uses it
Sift's pure `allocator` is the *reference model* of DRA-style matching; the
driver (Phase 1) expresses that model in real DRA objects against a live cluster.

## Gotcha / what confused me
_(your turn — e.g. DRA being a within-cluster mechanism, or device-plugin vs DRA)_

## See also
- `docs/concepts/resourceslice-resourceclaim.md`
- `docs/concepts/structured-parameters.md`
- ADR-0013 (DRA expression via structured parameters)